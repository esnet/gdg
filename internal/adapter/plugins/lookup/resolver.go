// Package lookup resolves "lookup:<name>:<key>[.<json_field>]" references
// against configured lookup plugin providers (see plugins/lookup/gsm for
// the first implementation, Google Secret Manager).
package lookup

import (
	"fmt"
	"strings"
	"sync"

	"github.com/esnet/gdg/internal/config/config_domain"
	"github.com/esnet/gdg/internal/ports/outbound"
	"github.com/tidwall/gjson"
)

// Compile-time assertions that Resolver satisfies both outbound.LookupResolver
// and outbound.LookupService.
var (
	_ outbound.LookupResolver = (*Resolver)(nil)
	_ outbound.LookupService  = (*Resolver)(nil)
)

// lookupPrefix is the prefix that marks a config value as a lookup
// reference. Defined here (rather than in outbound) because Resolver owns
// the logic of what constitutes a lookup reference.
const lookupPrefix = "lookup:"

// IsLookupRef reports whether raw is a lookup reference.
// Implements outbound.LookupService.
func (r *Resolver) IsLookupRef(raw string) bool {
	return strings.HasPrefix(raw, lookupPrefix)
}

// Prefix returns the prefix string that marks a value as a lookup reference.
// Implements outbound.LookupService.
func (r *Resolver) Prefix() string {
	return lookupPrefix
}

// Resolver resolves "lookup:<name>:<key>[.<json_field>]" values against
// named lookup plugin providers, caching resolved values in memory for
// the life of the process.
//
// This caching is distinct from (and in addition to) the provider-level
// caching each LookupProvider already does: a GSM/Vault-backed provider
// caches its underlying plugin *instance* (the extism.Plugin, created
// once and reused — see gsm.PluginLookupGSM). Resolver additionally
// caches the *resolved value* itself, keyed by the full raw reference
// string, since a Vault/GSM round trip is comparatively expensive and the
// same reference may be looked up more than once in a single gdg
// invocation.
type Resolver struct {
	enabled   bool
	providers map[string]outbound.LookupProvider
	cache     sync.Map // raw reference string -> resolved value
}

// NewResolver builds a Resolver from cfg.
//
// If lookups are disabled (cfg is nil, or cfg.LookupEnabled() is false),
// NewResolver still returns a usable, non-nil Resolver rather than an
// error: it simply has no providers, and Resolve on it will report a
// clear "lookups are disabled" error only if something actually tries to
// resolve a "lookup:" reference. This lets callers unconditionally
// construct and pass around a Resolver without special-casing the
// disabled state everywhere.
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

// ProviderFactory constructs a LookupProvider from a PluginEntity.
// Concrete provider packages (e.g. gsm) supply one of these via RegisterProvider.
type ProviderFactory func(*config_domain.PluginEntity) (outbound.LookupProvider, error)

// providerRegistry maps provider names (e.g. "gsm") to their factory
// functions. Populated by explicit RegisterProvider calls at program startup
// — not via init() — so the caller controls initialization order.
var providerRegistry = map[string]ProviderFactory{}

// RegisterProvider registers a factory for the named lookup provider.
// Call this once at program startup before any NewResolver call, e.g.:
//
//	lookup.RegisterProvider("gsm", gsm.NewPluginLookupGSM)
func RegisterProvider(name string, factory ProviderFactory) {
	providerRegistry[name] = factory
}

// buildProvider looks up the registered factory for name and constructs the provider.
func buildProvider(name string, entity *config_domain.PluginEntity) (outbound.LookupProvider, error) {
	factory, ok := providerRegistry[name]
	if !ok {
		return nil, fmt.Errorf("no lookup plugin implementation registered for provider %q", name)
	}
	return factory(entity)
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
func parseRef(raw string) (name, key, jsonField string, err error) {
	rest, ok := strings.CutPrefix(raw, lookupPrefix)
	if !ok {
		return "", "", "", fmt.Errorf("lookup plugin: %q is not a lookup reference (must start with %q)", raw, lookupPrefix)
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
