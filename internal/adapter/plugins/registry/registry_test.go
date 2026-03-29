package registry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/esnet/gdg/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// sampleEntries is the registry payload used across tests.
var sampleEntries = []domain.PluginRegistryEntry{
	{
		Name:        "aes-256-gcm",
		Type:        domain.PluginTypeCipher,
		Description: "seeded implementation of aes-256",
		Source:      "https://github.com/esnet/gdg-plugins/tree/main/cipher/aes-256-gcm",
		URLPattern:  "https://github.com/esnet/gdg-plugins/raw/refs/tags/{version}/plugins/cipher_aes256_gcm.wasm",
		Versions: []domain.PluginVersionEntry{
			{Version: "0.1.0", ConfigFields: []string{"passphrase"}},
		},
	},
	{
		Name:        "ansible-vault",
		Type:        domain.PluginTypeCipher,
		Description: "golang implementation of ansible-vault",
		Source:      "https://github.com/esnet/gdg-plugins/tree/main/cipher/ansible",
		URLPattern:  "https://github.com/esnet/gdg-plugins/raw/refs/tags/{version}/plugins/cipher_ansible.wasm",
		Versions: []domain.PluginVersionEntry{
			{Version: "0.1.0", ConfigFields: []string{"vault_password"}},
		},
	},
	{
		Name:    "future-plugin",
		Type:    "future-type",
		Versions: []domain.PluginVersionEntry{
			{Version: "1.0.0", ConfigFields: []string{}},
		},
	},
}

// newTestServer starts an httptest server that serves sampleEntries as JSON.
// The caller is responsible for calling ts.Close().
func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	raw, err := json.Marshal(sampleEntries)
	require.NoError(t, err)
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(raw)
	}))
}

// newLocalFile writes sampleEntries to a temp file and returns its path.
func newLocalFile(t *testing.T) string {
	t.Helper()
	raw, err := json.Marshal(sampleEntries)
	require.NoError(t, err)
	dir := t.TempDir()
	path := filepath.Join(dir, "plugin_registry.json")
	require.NoError(t, os.WriteFile(path, raw, 0o600))
	return path
}

// ── All ──────────────────────────────────────────────────────────────────────

func TestAll_Remote(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	c := NewClient(ClientConfig{URL: ts.URL, HTTPClient: ts.Client()})
	entries, err := c.All()
	require.NoError(t, err)
	assert.Len(t, entries, len(sampleEntries))
}

func TestAll_LocalFile(t *testing.T) {
	path := newLocalFile(t)
	c := NewClient(ClientConfig{FilePath: path})
	entries, err := c.All()
	require.NoError(t, err)
	assert.Len(t, entries, len(sampleEntries))
}

func TestAll_LocalFileTakesPrecedenceOverURL(t *testing.T) {
	path := newLocalFile(t)
	// URL points to a non-existent server — if it were used, the test would fail.
	c := NewClient(ClientConfig{FilePath: path, URL: "http://127.0.0.1:0/should-not-be-called"})
	entries, err := c.All()
	require.NoError(t, err)
	assert.Len(t, entries, len(sampleEntries))
}

func TestAll_CachesResult(t *testing.T) {
	callCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		raw, _ := json.Marshal(sampleEntries)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(raw)
	}))
	defer ts.Close()

	c := NewClient(ClientConfig{URL: ts.URL, HTTPClient: ts.Client()})
	_, _ = c.All()
	_, _ = c.All()
	_, _ = c.All()
	assert.Equal(t, 1, callCount, "registry should only be fetched once per Client instance")
}

func TestAll_HTTP404(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	c := NewClient(ClientConfig{URL: ts.URL, HTTPClient: ts.Client()})
	_, err := c.All()
	assert.ErrorContains(t, err, "HTTP 404")
}

func TestAll_InvalidJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("not json"))
	}))
	defer ts.Close()

	c := NewClient(ClientConfig{URL: ts.URL, HTTPClient: ts.Client()})
	_, err := c.All()
	assert.ErrorContains(t, err, "parsing registry JSON")
}

func TestAll_MissingLocalFile(t *testing.T) {
	c := NewClient(ClientConfig{FilePath: "/nonexistent/path/registry.json"})
	_, err := c.All()
	assert.ErrorContains(t, err, "reading registry file")
}

func TestAll_FallsBackToDefaultURL(t *testing.T) {
	// Provide neither FilePath nor URL — the client must use domain.RegistryDefaultURL.
	// We can't hit GitHub in tests, so we just verify the error message contains the URL.
	c := NewClient(ClientConfig{HTTPClient: &http.Client{Transport: &blockingTransport{}}})
	_, err := c.All()
	assert.ErrorContains(t, err, domain.RegistryDefaultURL)
}

// blockingTransport is an http.RoundTripper that always returns an error,
// used to verify which URL the client attempts to connect to.
type blockingTransport struct{}

func (b *blockingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return nil, fmt.Errorf("blocked: %s", req.URL.String())
}

// ── CipherPlugins ─────────────────────────────────────────────────────────────

func TestCipherPlugins_FiltersNonCipher(t *testing.T) {
	path := newLocalFile(t)
	c := NewClient(ClientConfig{FilePath: path})
	ciphers, err := c.CipherPlugins()
	require.NoError(t, err)
	assert.Len(t, ciphers, 2)
	for _, e := range ciphers {
		assert.Equal(t, domain.PluginTypeCipher, e.Type)
	}
}

// ── Find ──────────────────────────────────────────────────────────────────────

func TestFind_KnownPlugin(t *testing.T) {
	path := newLocalFile(t)
	c := NewClient(ClientConfig{FilePath: path})
	entry, err := c.Find("aes-256-gcm")
	require.NoError(t, err)
	assert.Equal(t, "aes-256-gcm", entry.Name)
}

func TestFind_CaseInsensitive(t *testing.T) {
	path := newLocalFile(t)
	c := NewClient(ClientConfig{FilePath: path})
	entry, err := c.Find("AES-256-GCM")
	require.NoError(t, err)
	assert.Equal(t, "aes-256-gcm", entry.Name)
}

func TestFind_UnknownPlugin(t *testing.T) {
	path := newLocalFile(t)
	c := NewClient(ClientConfig{FilePath: path})
	_, err := c.Find("does-not-exist")
	assert.ErrorContains(t, err, "not found in registry")
}

// ── ResolvePlugin ─────────────────────────────────────────────────────────────

func TestResolvePlugin_ExplicitVersion(t *testing.T) {
	path := newLocalFile(t)
	c := NewClient(ClientConfig{FilePath: path})
	entry, vEntry, wasmURL, err := c.ResolvePlugin("aes-256-gcm", "0.1.0")
	require.NoError(t, err)
	assert.Equal(t, "aes-256-gcm", entry.Name)
	assert.Equal(t, "0.1.0", vEntry.Version)
	assert.Equal(t,
		"https://github.com/esnet/gdg-plugins/raw/refs/tags/0.1.0/plugins/cipher_aes256_gcm.wasm",
		wasmURL,
	)
}

func TestResolvePlugin_DefaultsToLatest(t *testing.T) {
	path := newLocalFile(t)
	c := NewClient(ClientConfig{FilePath: path})
	_, vEntry, _, err := c.ResolvePlugin("aes-256-gcm", "")
	require.NoError(t, err)
	// sampleEntries has one version — 0.1.0 is both first and latest.
	assert.Equal(t, "0.1.0", vEntry.Version)
}

func TestResolvePlugin_UnknownVersion(t *testing.T) {
	path := newLocalFile(t)
	c := NewClient(ClientConfig{FilePath: path})
	_, _, _, err := c.ResolvePlugin("aes-256-gcm", "9.9.9")
	assert.ErrorContains(t, err, "version")
	assert.ErrorContains(t, err, "9.9.9")
}

func TestResolvePlugin_UnknownPlugin(t *testing.T) {
	path := newLocalFile(t)
	c := NewClient(ClientConfig{FilePath: path})
	_, _, _, err := c.ResolvePlugin("unknown-plugin", "")
	assert.ErrorContains(t, err, "not found in registry")
}
