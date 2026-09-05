package gsm

import (
	"context"
	"fmt"

	"github.com/esnet/gdg/internal/config/config_domain"
	"github.com/esnet/gdg/internal/ports/outbound"
	extism "github.com/extism/go-sdk"
)

// secretManagerHost is the only host the GSM lookup guest plugin is
// permitted to reach over HTTP, enforced via the extism manifest's
// AllowedHosts — the guest gets outbound access to Secret Manager and
// nothing else.
const secretManagerHost = "secretmanager.googleapis.com"

// credentialsConfigKey is the "config" field name (under
// plugins.lookup.gsm.config) holding the resolved credentials value —
// either a path to a service-account JSON file or the JSON itself,
// depending on whether the operator used an "env:" or "file:" prefix (see
// PluginEntity.GetPluginConfig() and loadCredentialsJSON in credentials.go).
const credentialsConfigKey = "credentials"

// PluginLookupGSM resolves "lookup:<name>:<key>" values against Google
// Secret Manager via a WASM guest plugin, following the same shape as
// cipher.PluginCipherEncoder: the extism.Plugin instance (wasmExec) is
// created once in NewPluginLookupGSM and cached for the life of the
// process, reused across every Lookup call rather than re-initialized
// per call.
type PluginLookupGSM struct {
	cfg      *config_domain.PluginEntity
	wasmExec *extism.Plugin
}

// Lookup resolves key (a Secret Manager resource name, e.g.
// "projects/<id>/secrets/<name>/versions/<n>") to its secret value by
// invoking the guest plugin's exported Lookup function.
func (p PluginLookupGSM) Lookup(key string) (string, error) {
	exit, out, err := p.wasmExec.Call(outbound.LookupOperation, []byte(key))
	if err != nil {
		return "", err
	}
	if exit != 0 {
		return "", fmt.Errorf("gsm lookup plugin returned non-zero exit code, failed to resolve %q", key)
	}
	return string(out), nil
}

// resolveWasmSource picks the extism.Wasm source for plugCfg, mirroring
// cipher.NewPluginCipherEncoder's precedence: an explicit FilePath wins
// over a Url, and at least one of the two must be set.
//
// Pulled out as its own function (rather than inlined in
// NewPluginLookupGSM) so this branch of config handling is directly unit
// testable without needing to construct a real extism.Plugin.
func resolveWasmSource(plugCfg *config_domain.PluginEntity) (extism.Wasm, error) {
	switch {
	case plugCfg.FilePath != "":
		return extism.WasmFile{Path: plugCfg.FilePath}, nil
	case plugCfg.Url != "":
		return extism.WasmUrl{Url: plugCfg.Url}, nil
	default:
		return nil, fmt.Errorf("gsm lookup plugin configuration is invalid: no URL or file path was provided")
	}
}

// NewPluginLookupGSM constructs a GSM-backed LookupProvider from plugCfg.
//
// Credential resolution follows the same env:/file: convention every
// other plugin's config uses (see PluginEntity.GetPluginConfig()); the
// resulting OAuth2 token source is handed to the guest on demand via the
// get_gcp_access_token host function (host_functions.go), never baked
// into the static manifest config — so the cached plugin instance never
// goes stale even though GCP access tokens themselves expire.
//
// Credential resolution happens before the WASM module is ever fetched or
// instantiated, so a misconfigured "credentials" value fails fast without
// requiring network access or a reachable plugin URL.
func NewPluginLookupGSM(plugCfg *config_domain.PluginEntity) (outbound.LookupProvider, error) {
	if plugCfg == nil {
		return nil, fmt.Errorf("gsm lookup plugin: no configuration provided")
	}

	wasmSource, err := resolveWasmSource(plugCfg)
	if err != nil {
		return nil, err
	}

	resolvedConfig := plugCfg.GetPluginConfig()

	ctx := context.Background()
	tokenSource, err := NewTokenSource(ctx, resolvedConfig[credentialsConfigKey])
	if err != nil {
		return nil, err
	}

	manifest := extism.Manifest{
		Wasm:         []extism.Wasm{wasmSource},
		Config:       resolvedConfig,
		AllowedHosts: []string{secretManagerHost},
	}

	pluginConfig := extism.PluginConfig{
		EnableWasi: true,
	}

	hostFn := newGetAccessTokenHostFunction(tokenSource)

	plugin, err := extism.NewPlugin(ctx, manifest, pluginConfig, []extism.HostFunction{hostFn})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize gsm lookup plugin: %w", err)
	}

	return PluginLookupGSM{cfg: plugCfg, wasmExec: plugin}, nil
}
