package gsm

import (
	"context"
	"fmt"

	"github.com/esnet/gdg/internal/adapter/plugins/lookup"
	"github.com/esnet/gdg/internal/config/config_domain"
	"github.com/esnet/gdg/internal/ports/outbound"
	extism "github.com/extism/go-sdk"
)

// secretManagerHost is the only host the GSM lookup guest plugin is
// permitted to reach over HTTP, enforced via the extism manifest's
// AllowedHosts.
const secretManagerHost = "secretmanager.googleapis.com"

// credentialsConfigKey is the "config" field name holding the resolved
// credentials value (a path to a service-account JSON file or the JSON
// itself, depending on whether the operator used an "env:" or "file:"
// prefix).
const credentialsConfigKey = "credentials"

// PluginLookupGSM resolves "lookup:<name>:<key>" values against Google
// Secret Manager via a WASM guest plugin. It embeds lookup.Plugin
// for shared WASM lifecycle management and satisfies
// outbound.LookupProvider. GSM-specific concerns (OAuth2 credentials,
// AllowedHosts, host function registration) are handled here.
type PluginLookupGSM struct {
	lookup.Plugin
}

// Compile-time assertion that PluginLookupGSM satisfies outbound.LookupProvider.
var _ outbound.LookupProvider = (*PluginLookupGSM)(nil)

// Lookup resolves key (a Secret Manager resource name, e.g.
// "projects/<id>/secrets/<name>/versions/<n>") to its secret value.
func (p *PluginLookupGSM) Lookup(key string) (string, error) {
	return p.Call(lookup.Operation, []byte(key))
}

// NewPluginLookupGSM constructs a GSM-backed LookupProvider from plugCfg.
//
// Credential resolution happens before the WASM module is ever fetched or
// instantiated, so a misconfigured "credentials" value fails fast without
// requiring network access or a reachable plugin URL.
func NewPluginLookupGSM(plugCfg *config_domain.PluginEntity) (outbound.LookupProvider, error) {
	if plugCfg == nil {
		return nil, fmt.Errorf("gsm lookup plugin: no configuration provided")
	}

	wasmSource, err := lookup.ResolveWasmSource(plugCfg)
	if err != nil {
		return nil, err
	}

	resolvedConfig := plugCfg.GetPluginConfig()

	tokenSource, err := NewTokenSource(context.Background(), resolvedConfig[credentialsConfigKey])
	if err != nil {
		return nil, err
	}

	manifest := extism.Manifest{
		Wasm:         []extism.Wasm{wasmSource},
		Config:       resolvedConfig,
		AllowedHosts: []string{secretManagerHost},
	}

	base, err := lookup.NewLookupPlugin(plugCfg, manifest, []extism.HostFunction{
		newGetAccessTokenHostFunction(tokenSource),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize gsm lookup plugin: %w", err)
	}

	return &PluginLookupGSM{Plugin: base}, nil
}
