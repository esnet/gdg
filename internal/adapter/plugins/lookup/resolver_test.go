package lookup

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/esnet/gdg/internal/config/config_domain"
	"github.com/esnet/gdg/internal/ports/outbound"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── test doubles ────────────────────────────────────────────────────────

// fakeProvider is a minimal outbound.LookupProvider for testing Resolve's
// own logic (caching, ref parsing, JSON-field extraction) independently
// of any real plugin implementation.
type fakeProvider struct {
	value string
	err   error
	calls *int
}

func (f fakeProvider) Lookup(_ string) (string, error) {
	if f.calls != nil {
		*f.calls++
	}
	if f.err != nil {
		return "", f.err
	}
	return f.value, nil
}

var _ outbound.LookupProvider = fakeProvider{}

// ── parseRef ────────────────────────────────────────────────────────────

func TestParseRef(t *testing.T) {
	tests := []struct {
		name          string
		raw           string
		wantProvider  string
		wantKey       string
		wantJSONField string
		wantErr       bool
	}{
		{
			name:         "simple key, no field",
			raw:          "lookup:gsm:projects/p/secrets/s/versions/1",
			wantProvider: "gsm",
			wantKey:      "projects/p/secrets/s/versions/1",
		},
		{
			name:          "key with json field suffix",
			raw:           "lookup:gsm:projects/p/secrets/s/versions/1.key_name",
			wantProvider:  "gsm",
			wantKey:       "projects/p/secrets/s/versions/1",
			wantJSONField: "key_name",
		},
		{
			name:          "key with multiple dots splits on the last one",
			raw:           "lookup:gsm:a.b.c",
			wantProvider:  "gsm",
			wantKey:       "a.b",
			wantJSONField: "c",
		},
		{
			name:         "trailing dot with nothing after it is kept as part of the key",
			raw:          "lookup:gsm:a.",
			wantProvider: "gsm",
			wantKey:      "a.",
		},
		{
			name:    "missing prefix",
			raw:     "gsm:a",
			wantErr: true,
		},
		{
			name:    "missing name/key separator",
			raw:     "lookup:gsmonly",
			wantErr: true,
		},
		{
			name:    "empty name",
			raw:     "lookup::key",
			wantErr: true,
		},
		{
			name:    "empty key",
			raw:     "lookup:gsm:",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, key, jsonField, err := parseRef(tt.raw)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantProvider, name)
			assert.Equal(t, tt.wantKey, key)
			assert.Equal(t, tt.wantJSONField, jsonField)
		})
	}
}

// ── IsRef ───────────────────────────────────────────────────────────────

func TestIsRef(t *testing.T) {
	assert.True(t, IsRef("lookup:gsm:foo"))
	assert.False(t, IsRef("env:FOO"))
	assert.False(t, IsRef("file:/foo"))
	assert.False(t, IsRef("plain-value"))
	assert.False(t, IsRef(""))
}

// ── Resolve ─────────────────────────────────────────────────────────────

func TestResolve_Success_NoJSONField(t *testing.T) {
	r := &Resolver{enabled: true, providers: map[string]outbound.LookupProvider{
		"gsm": fakeProvider{value: "super-secret-token"},
	}}

	got, err := r.Resolve("lookup:gsm:projects/p/secrets/s/versions/1")
	require.NoError(t, err)
	assert.Equal(t, "super-secret-token", got)
}

func TestResolve_Success_WithJSONFieldExtraction(t *testing.T) {
	r := &Resolver{enabled: true, providers: map[string]outbound.LookupProvider{
		"gsm": fakeProvider{value: `{"api_key": "super-secret-token", "other": "ignored"}`},
	}}

	got, err := r.Resolve("lookup:gsm:projects/p/secrets/s/versions/1.api_key")
	require.NoError(t, err)
	assert.Equal(t, "super-secret-token", got)
}

func TestResolve_JSONFieldNotFound_ReturnsError(t *testing.T) {
	r := &Resolver{enabled: true, providers: map[string]outbound.LookupProvider{
		"gsm": fakeProvider{value: `{"other": "value"}`},
	}}

	_, err := r.Resolve("lookup:gsm:projects/p/secrets/s/versions/1.missing_field")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `field "missing_field" not found`)
}

func TestResolve_UnknownProvider_ReturnsError(t *testing.T) {
	r := &Resolver{enabled: true, providers: map[string]outbound.LookupProvider{
		"gsm": fakeProvider{value: "irrelevant"},
	}}

	_, err := r.Resolve("lookup:vault:some/path")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `no provider configured named "vault"`)
}

func TestResolve_MalformedRef_ReturnsError(t *testing.T) {
	r := &Resolver{enabled: true, providers: map[string]outbound.LookupProvider{}}

	_, err := r.Resolve("not-a-lookup-ref")
	require.Error(t, err)
}

func TestResolve_ProviderError_IsWrapped(t *testing.T) {
	r := &Resolver{enabled: true, providers: map[string]outbound.LookupProvider{
		"gsm": fakeProvider{err: errors.New("permission denied")},
	}}

	_, err := r.Resolve("lookup:gsm:projects/p/secrets/s/versions/1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "permission denied")
}

func TestResolve_Disabled_ReturnsError(t *testing.T) {
	r := &Resolver{enabled: false}

	_, err := r.Resolve("lookup:gsm:projects/p/secrets/s/versions/1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "lookups are disabled")
}

func TestResolve_CachesResolvedValue(t *testing.T) {
	calls := 0
	r := &Resolver{enabled: true, providers: map[string]outbound.LookupProvider{
		"gsm": fakeProvider{value: "cached-value", calls: &calls},
	}}

	ref := "lookup:gsm:projects/p/secrets/s/versions/1"

	got1, err := r.Resolve(ref)
	require.NoError(t, err)
	got2, err := r.Resolve(ref)
	require.NoError(t, err)

	assert.Equal(t, "cached-value", got1)
	assert.Equal(t, "cached-value", got2)
	assert.Equal(t, 1, calls, "the underlying provider should only be called once; the second Resolve must hit the cache")
}

// ── NewResolver ─────────────────────────────────────────────────────────

func TestNewResolver_NilConfig_ReturnsDisabledResolver(t *testing.T) {
	r, err := NewResolver(nil)
	require.NoError(t, err)
	require.NotNil(t, r)

	_, err = r.Resolve("lookup:gsm:foo")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "lookups are disabled")
}

func TestNewResolver_LookupSpecificallyDisabled_ReturnsDisabledResolver(t *testing.T) {
	cfg := &config_domain.PluginConfig{
		Lookup: config_domain.LookupConfig{Disabled: true},
	}

	r, err := NewResolver(cfg)
	require.NoError(t, err)

	_, err = r.Resolve("lookup:gsm:foo")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "lookups are disabled")
}

func TestNewResolver_UnknownProviderName_ReturnsError(t *testing.T) {
	cfg := &config_domain.PluginConfig{
		Lookup: config_domain.LookupConfig{
			Plugins: map[string]*config_domain.PluginEntity{
				"vault": {Url: "https://example.com/vault.wasm"},
			},
		},
	}

	_, err := NewResolver(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no lookup plugin implementation registered")
}

// fakeServiceAccountJSON builds a syntactically valid GCP service-account
// key JSON document backed by a freshly generated RSA key, mirroring the
// helper in the gsm package's own tests -- duplicated here (rather than
// exported from gsm) since it's a small, test-only fixture and this test
// wants to exercise the real gsm.NewPluginLookupGSM path via buildProvider
// without depending on gsm's unexported test internals.
func fakeServiceAccountJSON(t *testing.T) string {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	der, err := x509.MarshalPKCS8PrivateKey(key)
	require.NoError(t, err)

	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})

	return fmt.Sprintf(`{
		"type": "service_account",
		"project_id": "test-project",
		"private_key_id": "test-key-id",
		"private_key": %q,
		"client_email": "test@test-project.iam.gserviceaccount.com",
		"client_id": "1234567890",
		"token_uri": "https://oauth2.googleapis.com/token"
	}`, string(pemBytes))
}

// minimalWasmModule is the smallest possible valid WebAssembly binary --
// see gsm's own tests for the full explanation of why this is sufficient
// to exercise a real extism.NewPlugin construction.
var minimalWasmModule = []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}

func TestNewResolver_KnownProvider_DispatchesToRealConstructor(t *testing.T) {
	dir := t.TempDir()
	wasmPath := filepath.Join(dir, "empty.wasm")
	require.NoError(t, os.WriteFile(wasmPath, minimalWasmModule, 0o600))

	cfg := &config_domain.PluginConfig{
		Lookup: config_domain.LookupConfig{
			Plugins: map[string]*config_domain.PluginEntity{
				"gsm": {
					FilePath: wasmPath,
					PluginConfig: map[string]string{
						"credentials": fakeServiceAccountJSON(t),
					},
				},
			},
		},
	}

	r, err := NewResolver(cfg)
	require.NoError(t, err, "expected buildProvider to dispatch \"gsm\" to gsm.NewPluginLookupGSM and succeed")
	require.NotNil(t, r)
	_, ok := r.providers["gsm"]
	assert.True(t, ok, "expected a \"gsm\" provider to be registered")
}
