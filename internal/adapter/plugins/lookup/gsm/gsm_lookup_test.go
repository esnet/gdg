package gsm

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/esnet/gdg/internal/adapter/plugins/lookup"
	"github.com/esnet/gdg/internal/config/config_domain"
	extism "github.com/extism/go-sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// minimalWasmModule is the smallest possible valid WebAssembly binary:
// just the 4-byte magic number and 4-byte version, with zero sections —
// no imports, no exports, no memory. A module with no sections is legal
// per the WebAssembly binary format spec, and extism's runtime (wazero
// under the hood) compiles and instantiates it without issue since it
// requires nothing to be satisfied.
//
// This is used to exercise NewPluginLookupGSM's actual success path
// (extism.NewPlugin really succeeding, with AllowedHosts and our host
// function registered) and Lookup()'s "function not found" error path —
// for real, without needing a real compiled guest plugin, which doesn't
// exist yet (see gsm_plan.md step 11).
var minimalWasmModule = []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}

func writeMinimalWasmModule(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.wasm")
	require.NoError(t, os.WriteFile(path, minimalWasmModule, 0o600))
	return path
}

// ── plugins.ResolveWasmSource ───────────────────────────────────────────
// These tests live here (adjacent to the GSM plugin) because ResolveWasmSource
// was extracted from this package; keeping them here avoids a separate test
// file for a one-function helper.

func TestResolveWasmSource_FilePathTakesPrecedenceOverURL(t *testing.T) {
	cfg := &config_domain.PluginEntity{
		FilePath: "/opt/gdg/plugins/lookup_gsm.wasm",
		Url:      "https://example.com/lookup_gsm.wasm",
	}

	src, err := lookup.ResolveWasmSource(cfg)
	require.NoError(t, err)

	file, ok := src.(extism.WasmFile)
	require.True(t, ok, "expected extism.WasmFile when FilePath is set, got %T", src)
	assert.Equal(t, "/opt/gdg/plugins/lookup_gsm.wasm", file.Path)
}

func TestResolveWasmSource_URLUsedWhenNoFilePath(t *testing.T) {
	cfg := &config_domain.PluginEntity{
		Url: "https://example.com/lookup_gsm.wasm",
	}

	src, err := lookup.ResolveWasmSource(cfg)
	require.NoError(t, err)

	u, ok := src.(extism.WasmUrl)
	require.True(t, ok, "expected extism.WasmUrl when only Url is set, got %T", src)
	assert.Equal(t, "https://example.com/lookup_gsm.wasm", u.Url)
}

func TestResolveWasmSource_NeitherSetReturnsError(t *testing.T) {
	cfg := &config_domain.PluginEntity{}

	_, err := lookup.ResolveWasmSource(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no URL or file path was provided")
}

// ── NewPluginLookupGSM: validation branches ─────────────────────────────
//
// These cover the failure branches that run before the WASM module is
// ever fetched or instantiated. The success path (and Lookup itself) are
// covered further below using a minimal real WASM module.

func TestNewPluginLookupGSM_NilConfig_ReturnsError(t *testing.T) {
	_, err := NewPluginLookupGSM(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no configuration provided")
}

func TestNewPluginLookupGSM_MissingURLAndFilePath_ReturnsError(t *testing.T) {
	_, err := NewPluginLookupGSM(&config_domain.PluginEntity{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no URL or file path was provided")
}

func TestNewPluginLookupGSM_InvalidCredentials_ReturnsErrorBeforeFetchingWasm(t *testing.T) {
	// A syntactically valid (if unreachable) URL is set, so the only thing
	// that should make this fail is credential resolution -- proving that
	// happens before any attempt to fetch/instantiate the WASM module
	// (no network call for the "wasm" URL should occur here).
	cfg := &config_domain.PluginEntity{
		Url: "https://example.invalid/lookup_gsm.wasm",
		PluginConfig: map[string]string{
			"credentials": "", // empty -> loadCredentialsJSON fails fast
		},
	}

	_, err := NewPluginLookupGSM(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no credentials configured")
}

func TestNewPluginLookupGSM_MalformedCredentialsJSON_ReturnsError(t *testing.T) {
	cfg := &config_domain.PluginEntity{
		Url: "https://example.invalid/lookup_gsm.wasm",
		PluginConfig: map[string]string{
			"credentials": "{not valid json",
		},
	}

	_, err := NewPluginLookupGSM(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unable to parse credentials JSON")
}

// ── NewPluginLookupGSM / Lookup — real (minimal) WASM instantiation ─────

func TestNewPluginLookupGSM_Succeeds_WithValidConfigAndMinimalGuest(t *testing.T) {
	cfg := &config_domain.PluginEntity{
		FilePath: writeMinimalWasmModule(t),
		PluginConfig: map[string]string{
			"credentials": fakeServiceAccountJSON(t),
		},
	}

	provider, err := NewPluginLookupGSM(cfg)
	require.NoError(t, err, "expected extism.NewPlugin to succeed with a valid (if empty) guest module and valid credentials")
	require.NotNil(t, provider)
}

func TestPluginLookupGSM_Lookup_ReturnsErrorWhenGuestExportsNoLookupFunction(t *testing.T) {
	cfg := &config_domain.PluginEntity{
		FilePath: writeMinimalWasmModule(t),
		PluginConfig: map[string]string{
			"credentials": fakeServiceAccountJSON(t),
		},
	}

	provider, err := NewPluginLookupGSM(cfg)
	require.NoError(t, err)

	// The minimal module exports nothing, so calling "Lookup" must fail
	// cleanly (extism.Plugin.Call returns "unknown function: Lookup")
	// rather than panicking -- this exercises Lookup()'s error-return path.
	_, err = provider.Lookup("projects/test-project/secrets/foo/versions/1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown function")
}
