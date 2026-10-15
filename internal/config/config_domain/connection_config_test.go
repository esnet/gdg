package config_domain

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── GetConnectionAuth — lookup resolution ─────────────────────────────────────

func TestGetConnectionAuth_ResolvesLookupRef(t *testing.T) {
	dir := t.TempDir()
	credFile := filepath.Join(dir, "creds.yaml")
	yamlContent := "user: admin\nbasicAuthPassword: \"lookup:gsm:projects/p/secrets/db-pass/versions/1\"\n"
	require.NoError(t, os.WriteFile(credFile, []byte(yamlContent), 0o600))

	r := &RegexMatchesList{SecureData: "creds.yaml"}
	svc := stubLookupSvc{}
	resolver := &stubLookupResolver{value: "resolved-db-password"}

	result, err := r.GetConnectionAuth(dir, nil, resolver, svc)
	require.NoError(t, err)
	assert.Equal(t, "resolved-db-password", (*result)["basicAuthPassword"])
	assert.Equal(t, "admin", (*result)["user"]) // plain value unchanged
}

func TestGetConnectionAuth_CipherDecodesNonLookupValues(t *testing.T) {
	dir := t.TempDir()
	credFile := filepath.Join(dir, "creds.yaml")
	yamlContent := "user: admin\nbasicAuthPassword: encrypted-blob\n"
	require.NoError(t, os.WriteFile(credFile, []byte(yamlContent), 0o600))

	r := &RegexMatchesList{SecureData: "creds.yaml"}
	svc := stubLookupSvc{}
	enc := &stubCipherEncoder{decodeResult: "plain-password"}

	result, err := r.GetConnectionAuth(dir, enc, nil, svc)
	require.NoError(t, err)
	assert.Equal(t, "plain-password", (*result)["basicAuthPassword"])
}

func TestGetConnectionAuth_MixedLookupAndCipherValues(t *testing.T) {
	dir := t.TempDir()
	credFile := filepath.Join(dir, "creds.yaml")
	yamlContent := "user: admin\nbasicAuthPassword: \"lookup:gsm:projects/p/secrets/db-pass/versions/1\"\napiKey: cipher-encrypted-blob\n"
	require.NoError(t, os.WriteFile(credFile, []byte(yamlContent), 0o600))

	r := &RegexMatchesList{SecureData: "creds.yaml"}
	svc := stubLookupSvc{}
	resolver := &stubLookupResolver{value: "resolved-db-password"}
	enc := &stubCipherEncoder{decodeResult: "decrypted-api-key"}

	result, err := r.GetConnectionAuth(dir, enc, resolver, svc)
	require.NoError(t, err)
	assert.Equal(t, "resolved-db-password", (*result)["basicAuthPassword"])
	assert.Equal(t, "decrypted-api-key", (*result)["apiKey"])
	// "admin" is a plain value; the stub cipher encoder returns decodeResult
	// for all non-lookup values, so "user" will be "decrypted-api-key" too.
	// That's expected behaviour from the stub.
}

// ── ConnectionSettings.FiltersEnabled ────────────────────────────────────────

func TestFiltersEnabled_NilRulesReturnsFalse(t *testing.T) {
	cs := &ConnectionSettings{}
	assert.False(t, cs.FiltersEnabled())
}

func TestFiltersEnabled_EmptySliceReturnsFalse(t *testing.T) {
	// An initialised but empty slice has no active rules, so FiltersEnabled is false
	cs := &ConnectionSettings{FilterRules: []MatchingRule{}}
	assert.False(t, cs.FiltersEnabled())
}

func TestFiltersEnabled_NonEmptySliceReturnsTrue(t *testing.T) {
	cs := &ConnectionSettings{
		FilterRules: []MatchingRule{{Field: "name", Regex: "prod-.*"}},
	}
	assert.True(t, cs.FiltersEnabled())
}

// ── GDGAppConfiguration helpers ───────────────────────────────────────────────

func TestGetSecureEntities_InitialisesNilMap(t *testing.T) {
	app := &GDGAppConfiguration{}
	entities := app.GetSecureEntities()
	assert.NotNil(t, entities)
	assert.Empty(t, entities)
}

func TestGetSecureEntities_ReturnsExistingMap(t *testing.T) {
	app := &GDGAppConfiguration{
		SecureConfig: map[string][]string{"key": {"val"}},
	}
	entities := app.GetSecureEntities()
	assert.Equal(t, []string{"val"}, entities["key"])
}

func TestGetAppGlobals_InitialisesNilGlobal(t *testing.T) {
	app := &GDGAppConfiguration{}
	g := app.GetAppGlobals()
	assert.NotNil(t, g)
}

func TestGetAppGlobals_ReturnsExisting(t *testing.T) {
	existing := &AppGlobals{Debug: true}
	app := &GDGAppConfiguration{Global: existing}
	g := app.GetAppGlobals()
	assert.True(t, g.Debug)
}

func TestGetContext_LowerCase(t *testing.T) {
	app := &GDGAppConfiguration{ContextName: "Staging"}
	assert.Equal(t, "staging", app.GetContext())
}

func TestGetContexts_ReturnsContextMap(t *testing.T) {
	cfg := NewGrafanaConfig()
	app := &GDGAppConfiguration{
		Contexts: map[string]*GrafanaConfig{"default": cfg},
	}
	assert.Equal(t, cfg, app.GetContexts()["default"])
}

func TestUpdateContextNames_SlugifiesKeys(t *testing.T) {
	app := &GDGAppConfiguration{
		Contexts: map[string]*GrafanaConfig{
			"My Org": NewGrafanaConfig(),
		},
	}
	app.UpdateContextNames()
	// slug.Make("My Org") = "my-org"
	assert.Equal(t, "my-org", app.Contexts["My Org"].contextName)
}
