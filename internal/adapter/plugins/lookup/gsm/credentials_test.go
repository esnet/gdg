package gsm

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeServiceAccountJSON returns a syntactically valid GCP service-account
// key JSON document backed by a freshly generated RSA key, so that
// google.CredentialsFromJSON can successfully parse and build a
// TokenSource without ever making a network call (no Token() is invoked
// in these tests, only construction).
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

// ── loadCredentialsJSON ─────────────────────────────────────────────────

func TestLoadCredentialsJSON_InlineJSON(t *testing.T) {
	raw := `{"type": "service_account"}`
	data, err := loadCredentialsJSON(raw)
	require.NoError(t, err)
	assert.JSONEq(t, raw, string(data))
}

func TestLoadCredentialsJSON_InlineJSON_WithSurroundingWhitespace(t *testing.T) {
	raw := "  \n{\"type\": \"service_account\"}\n  "
	data, err := loadCredentialsJSON(raw)
	require.NoError(t, err)
	assert.JSONEq(t, `{"type": "service_account"}`, string(data))
}

func TestLoadCredentialsJSON_EmptyReturnsError(t *testing.T) {
	_, err := loadCredentialsJSON("")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no credentials configured")
}

func TestLoadCredentialsJSON_WhitespaceOnlyReturnsError(t *testing.T) {
	_, err := loadCredentialsJSON("   ")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no credentials configured")
}

func TestLoadCredentialsJSON_FilePath_ReadsFileContents(t *testing.T) {
	dir := t.TempDir()
	keyFile := filepath.Join(dir, "key.json")
	want := `{"type": "service_account", "project_id": "from-file"}`
	require.NoError(t, os.WriteFile(keyFile, []byte(want), 0o600))

	data, err := loadCredentialsJSON(keyFile)
	require.NoError(t, err)
	assert.JSONEq(t, want, string(data))
}

func TestLoadCredentialsJSON_MissingFileReturnsError(t *testing.T) {
	_, err := loadCredentialsJSON(filepath.Join(t.TempDir(), "does-not-exist.json"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unable to read credentials file")
}

// ── NewTokenSource ──────────────────────────────────────────────────────

func TestNewTokenSource_ValidServiceAccountJSON_Succeeds(t *testing.T) {
	ts, err := NewTokenSource(context.Background(), fakeServiceAccountJSON(t))
	require.NoError(t, err)
	assert.NotNil(t, ts)
}

func TestNewTokenSource_ValidServiceAccountFile_Succeeds(t *testing.T) {
	dir := t.TempDir()
	keyFile := filepath.Join(dir, "key.json")
	require.NoError(t, os.WriteFile(keyFile, []byte(fakeServiceAccountJSON(t)), 0o600))

	ts, err := NewTokenSource(context.Background(), keyFile)
	require.NoError(t, err)
	assert.NotNil(t, ts)
}

func TestNewTokenSource_EmptyCredentials_ReturnsError(t *testing.T) {
	_, err := NewTokenSource(context.Background(), "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no credentials configured")
}

func TestNewTokenSource_MalformedJSON_ReturnsError(t *testing.T) {
	_, err := NewTokenSource(context.Background(), "{not valid json")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unable to parse credentials JSON")
}

func TestNewTokenSource_MissingFile_ReturnsError(t *testing.T) {
	_, err := NewTokenSource(context.Background(), filepath.Join(t.TempDir(), "missing.json"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unable to read credentials file")
}

// TestNewTokenSource_AuthorizedUserType_Succeeds verifies that credentials
// written by `gcloud auth application-default login` — an "authorized_user"
// JSON shape, a developer's own OAuth2 refresh token rather than a service
// account key — are accepted. This is a deliberately loosened allowlist
// (see allowedCredentialTypes), not a return to the deprecated, fully
// type-agnostic google.CredentialsFromJSON.
func TestNewTokenSource_AuthorizedUserType_Succeeds(t *testing.T) {
	authorizedUserJSON := `{
		"type": "authorized_user",
		"client_id": "test-client-id",
		"client_secret": "test-client-secret",
		"refresh_token": "test-refresh-token"
	}`
	ts, err := NewTokenSource(context.Background(), authorizedUserJSON)
	require.NoError(t, err)
	assert.NotNil(t, ts)
}

// TestNewTokenSource_RejectsUnsupportedType verifies that credential shapes
// outside the explicit allowlist (service_account, authorized_user) are
// still rejected — the allowlist was widened by one entry, not removed.
func TestNewTokenSource_RejectsUnsupportedType(t *testing.T) {
	externalAccountJSON := `{
		"type": "external_account",
		"audience": "test-audience"
	}`
	_, err := NewTokenSource(context.Background(), externalAccountJSON)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported credentials type")
}
