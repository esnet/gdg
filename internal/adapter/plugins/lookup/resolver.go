// Package lookup resolves "lookup:<name>:<key>[.<json_field>]" references
// against configured lookup plugin providers (see plugins/lookup/gsm for
// the first implementation, Google Secret Manager).
package lookup

import (
	"fmt"
	"strings"
	"sync"

	"github.com/esnet/gdg/internal/adapter/plugins/lookup/gsm"
	"github.com/esnet/gdg/internal/config/config_domain"
	"github.com/esnet/gdg/internal/ports/outbound"
	"github.com/tidwall/gjson"
)

// Compile-time assertion that Resolver satisfies outbound.LookupResolver.
var _ outbound.LookupResolver = (*Resolver)(nil)

// Prefix and IsRef re-export outbound.LookupPrefix / outbound.IsLookupRef
// for convenience within this package. The canonical definitions live in
// outbound rather than here because config_domain also needs them (to
// decide whether a secure field should skip cipher decoding) and this
// package already imports config_domain — defining them here would create
// a circular import.
const Prefix = outbound.LookupPrefix

func IsRef(raw string) bool {
	return outbound.IsLookupRef(raw)
}

// Resolver resolves "lookup:<name>:<key>[.<json_field>]" values against
// named lookup plugin providers, caching resolved values in memory for
// the life of the process.
//
// This caching is distinct from (and in addition to) the provider-level
// caching each LookupProvider already does: a GSM/Vault-backed provider
// caches its underlying plugin *instance* (the extism.Plugin, created
// once and reused — see gsm.PluginLookupGSM), the same way a cipher
// plugin caches its WASM instance. Resolver additionally caches the
// *resolved value* itself, keyed by the full raw reference string, since
// a Vault/GSM round trip is comparatively expensive and the same
// reference may be looked up more than once in a single gdg invocation
// (e.g. referenced from multiple contexts). Neither the plugin instance
// nor the resolved value is ever persisted to disk.
type Resolver struct {
	enabled   bool
	providers map[string]outbound.LookupProvider
	cache     sync.Map // raw reference string -> resolved value
}

// NewResolver builds a Resolver from cfg.
//
// If lookups are disabled (cfg is nil, or cfg.LookupEnabled() is false —
// see PluginConfig.LookupEnabled, which checks the lookup-specific
// "plugins.lookup.disabled" flag), NewResolver still returns a usable, non-nil Resolver
// rather than an error: it simply has no providers, and Resolve on it
// will report a clear "lookups are disabled" error only if something
// actually tries to resolve a "lookup:" reference. This lets callers
// unconditionally construct and pass around a Resolver without special-
// casing the disabled state everywhere.
func NewResolver(cfg *config_domain.PluginConfig) (*Resolver, error) {
	if cfg == nil || !cfg.LookupEnabled() {
		return &Resolver{enabled: false}, nil
	}

	providers := make(map[string]outbound.LookupProvider, len(cfg.Lookup.Plugins))
	for name, entity := range cfg.Lookup.Plugins {
		provider, err := buildProvider(name, entity)
		if err != nil {
			return nil, fmt.Errorf("lookup plugin %q: %w", name, err)
		}
		providers[name] = provider
	}

	return &Resolver{enabled: true, providers: providers}, nil
}

// buildProvider constructs the concrete LookupProvider for a named
// "plugins.lookup.<name>" entry.
//
// There is no "type" or "provider" discriminator field on PluginEntity —
// the provider implementation is selected purely by the map key name
// (e.g. "gsm"), matching the exact config shape already in use:
//
//	lookup:
//	  disabled: false
//	  gsm:
//	    url: https://example.com/lookup_gsm.wasm
//	    config:
//	      credentials: env:GOOGLE_APPLICATION_CREDENTIALS
//
// Adding a new provider (e.g. "vault") means adding a case here once its
// PluginLookup* constructor exists — see gsm_plan.md / vault.md.
func buildProvider(name string, entity *config_domain.PluginEntity) (outbound.LookupProvider, error) {
	switch name {
	case "gsm":
		return gsm.NewPluginLookupGSM(entity)
	default:
		return nil, fmt.Errorf("no lookup plugin implementation registered for provider %q", name)
	}
}

// Resolve resolves raw (a full "lookup:<name>:<key>[.<json_field>]"
// reference) to its concrete value.
//
// If a value for raw has already been resolved during this process's
// lifetime, the cached value is returned without touching the underlying
// provider again.
func (r *Resolver) Resolve(raw string) (string, error) {
	if cached, ok := r.cache.Load(raw); ok {
		return cached.(string), nil
	}

	name, key, jsonField, err := parseRef(raw)
	if err != nil {
		return "", err
	}

	if !r.enabled {
		return "", fmt.Errorf("lookup plugin: cannot resolve %q, lookups are disabled (plugins.lookup.disabled is set)", raw)
	}

	provider, ok := r.providers[name]
	if !ok {
		return "", fmt.Errorf("lookup plugin: no provider configured named %q (referenced by %q)", name, raw)
	}

	value, err := provider.Lookup(key)
	if err != nil {
		return "", fmt.Errorf("lookup plugin %q: %w", name, err)
	}

	if jsonField != "" {
		result := gjson.Get(value, jsonField)
		if !result.Exists() {
			return "", fmt.Errorf("lookup plugin %q: field %q not found in resolved JSON value for %q", name, jsonField, raw)
		}
		value = result.String()
	}

	r.cache.Store(raw, value)
	return value, nil
}

// parseRef parses "lookup:<name>:<key>[.<json_field>]" into its parts.
//
// The optional ".<json_field>" suffix is recognized as everything after
// the LAST "." in the key portion, when present. This is unambiguous for
// GSM resource names (GCP disallows "." in secret IDs, so a real GSM key
// never contains one), but is a known limitation for any future provider
// whose keys can legitimately contain a literal "." that isn't meant as a
// field separator — none of the currently supported providers do.
func parseRef(raw string) (name, key, jsonField string, err error) {
	rest, ok := strings.CutPrefix(raw, Prefix)
	if !ok {
		return "", "", "", fmt.Errorf("lookup plugin: %q is not a lookup reference (must start with %q)", raw, Prefix)
	}

	idx := strings.Index(rest, ":")
	if idx <= 0 {
		return "", "", "", fmt.Errorf("lookup plugin: malformed lookup reference %q, expected \"lookup:<name>:<key>\"", raw)
	}
	name = rest[:idx]

	keyAndField := rest[idx+1:]
	if keyAndField == "" {
		return "", "", "", fmt.Errorf("lookup plugin: malformed lookup reference %q, missing key", raw)
	}

	key = keyAndField
	if dotIdx := strings.LastIndex(keyAndField, "."); dotIdx > 0 && dotIdx < len(keyAndField)-1 {
		key = keyAndField[:dotIdx]
		jsonField = keyAndField[dotIdx+1:]
	}

	return name, key, jsonField, nil
}
