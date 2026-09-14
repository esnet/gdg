// Package lookup provides the base service for WASM-backed lookup plugin
// implementations and the Resolver that dispatches to them.
package lookup

import (
	"context"
	"fmt"

	configDomain "github.com/esnet/gdg/internal/config/config_domain"
	extism "github.com/extism/go-sdk"
)

// Operation is the exported WASM guest function name every lookup
// plugin must implement. It is specific to the lookup plugin contract and
// lives here rather than in ports/outbound.
const Operation = "Lookup"

// Plugin is the base service for WASM-backed lookup provider
// implementations. Concrete providers (e.g. PluginLookupGSM) embed this
// struct and add any provider-specific fields (credentials, host
// allowlists, etc.) on top.
type Plugin struct {
	Cfg      *configDomain.PluginEntity
	WasmExec *extism.Plugin
}

// ResolveWasmSource picks the extism.Wasm source from plugCfg: an
// explicit FilePath wins over a Url, and at least one must be set.
func ResolveWasmSource(plugCfg *configDomain.PluginEntity) (extism.Wasm, error) {
	switch {
	case plugCfg.FilePath != "":
		return extism.WasmFile{Path: plugCfg.FilePath}, nil
	case plugCfg.Url != "":
		return extism.WasmUrl{Url: plugCfg.Url}, nil
	default:
		return nil, fmt.Errorf("plugin configuration is invalid: no URL or file path was provided")
	}
}

// NewLookupPlugin constructs a Plugin from plugCfg using the
// provided manifest and host functions. Callers build the manifest
// themselves (to add AllowedHosts, custom Config entries, etc.); this
// function handles the common extism.NewPlugin call.
func NewLookupPlugin(
	plugCfg *configDomain.PluginEntity,
	manifest extism.Manifest,
	hostFunctions []extism.HostFunction,
) (Plugin, error) {
	plugin, err := extism.NewPlugin(
		context.Background(),
		manifest,
		extism.PluginConfig{EnableWasi: true},
		hostFunctions,
	)
	if err != nil {
		return Plugin{}, fmt.Errorf("failed to initialize lookup plugin: %w", err)
	}
	return Plugin{Cfg: plugCfg, WasmExec: plugin}, nil
}

// Call invokes the named WASM guest export with input and returns the
// output as a string. Shared by all providers that embed Plugin.
func (p *Plugin) Call(operation string, input []byte) (string, error) {
	exit, out, err := p.WasmExec.Call(operation, input)
	if err != nil {
		return "", err
	}
	if exit != 0 {
		return "", fmt.Errorf("lookup plugin %q returned non-zero exit code", operation)
	}
	return string(out), nil
}
