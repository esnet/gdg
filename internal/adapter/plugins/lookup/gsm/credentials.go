// Package gsm implements a Google Secret Manager backed lookup plugin.
//
// Credential handling is intentionally kept on the host (this package),
// never inside the WASM guest: resolving Application Default Credentials
// (reading a service-account JSON file, or its raw contents, and turning
// that into a usable OAuth2 token source) requires filesystem and crypto
// operations that a tinygo/WASI guest cannot perform cleanly. The guest is
// only responsible for making the already-authenticated HTTPS call to the
// Secret Manager REST API (see gsm_lookup.go / host_functions.go).
package gsm

import (
	"context"
	"encoding/json/v2"
	"fmt"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// SecretManagerScope is the OAuth2 scope required to call the Secret
// Manager REST API.
const SecretManagerScope = "https://www.googleapis.com/auth/cloud-platform" // #nosec G101

// loadCredentialsJSON resolves the "credentials" plugin config value into
// raw service-account JSON bytes.
//
// The value is expected to have already passed through
// config_domain.PluginEntity.GetPluginConfig(), which resolves "env:" and
// "file:" prefixes the same way every other plugin's config does. That
// leaves two possible shapes here:
//
//   - The value IS the raw JSON already (this happens when the config used
//     "file:/path/to/key.json" — GetPluginConfig reads the file itself and
//     the config value becomes the file's contents).
//   - The value is a filesystem path to the JSON key file (this happens
//     when the config used "env:GOOGLE_APPLICATION_CREDENTIALS", since
//     that environment variable conventionally holds a path, not JSON
//     content, per Google's Application Default Credentials convention).
//
// loadCredentialsJSON distinguishes the two by checking whether the
// trimmed value looks like a JSON object; if not, it treats the value as a
// file path and reads it.
func loadCredentialsJSON(raw string) ([]byte, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, fmt.Errorf("gsm lookup plugin: no credentials configured (expected \"credentials\" config field)")
	}

	if strings.HasPrefix(trimmed, "{") {
		return []byte(trimmed), nil
	}

	data, err := os.ReadFile(trimmed) // #nosec G304 -- path comes from operator-controlled plugin config, same trust
	// level as other plugin config fields (e.g. cipher's file: resolution)
	if err != nil {
		return nil, fmt.Errorf("gsm lookup plugin: unable to read credentials file %q: %w", trimmed, err)
	}
	return data, nil
}

// allowedCredentialTypes are the only Google credential JSON shapes this
// plugin will load, keyed by their "type" field. Both are legitimate,
// commonly-used Application Default Credentials shapes:
//
//   - "service_account": a downloaded service-account key file — the
//     normal choice for anything unattended (CI, a deployed gdg instance).
//   - "authorized_user": written by `gcloud auth application-default
//     login` — a developer's own credentials, handy for local testing
//     without minting/managing a service-account key.
//
// Anything else (e.g. "external_account", "impersonated_service_account",
// "gdch_service_account") is rejected explicitly. This is still an
// allowlist, not the deprecated, fully type-agnostic
// google.CredentialsFromJSON: that function accepts every credential shape
// the oauth2/google package understands, which is broader than this
// plugin ever intends to support given the "credentials" config value
// could originate from an untrusted source.
var allowedCredentialTypes = map[string]google.CredentialsType{
	string(google.ServiceAccount): google.ServiceAccount,
	string(google.AuthorizedUser): google.AuthorizedUser,
}

// credentialsTypeField mirrors just the "type" field every Google
// credentials JSON shape carries, used to pick which google.CredentialsType
// to parse with.
type credentialsTypeField struct {
	Type string `json:"type"`
}

// NewTokenSource builds an OAuth2 token source from the resolved
// "credentials" plugin config value, scoped for Secret Manager access.
//
// The returned oauth2.TokenSource handles token expiry/refresh internally
// (via golang.org/x/oauth2), so callers do not need to track token
// lifetimes themselves — this is what lets the plugin instance (and its
// underlying extism.Plugin) stay cached and live for the lifetime of the
// process, the same way a cipher plugin's WASM instance is created once
// and reused, while still always handing the guest a valid token.
func NewTokenSource(ctx context.Context, credentialsConfig string) (oauth2.TokenSource, error) {
	jsonData, err := loadCredentialsJSON(credentialsConfig)
	if err != nil {
		return nil, err
	}

	var typed credentialsTypeField
	if err := json.Unmarshal(jsonData, &typed); err != nil {
		return nil, fmt.Errorf("gsm lookup plugin: unable to parse credentials JSON: %w", err)
	}

	credType, ok := allowedCredentialTypes[typed.Type]
	if !ok {
		return nil, fmt.Errorf(
			"gsm lookup plugin: unsupported credentials type %q (expected %q or %q)",
			typed.Type, google.ServiceAccount, google.AuthorizedUser,
		)
	}

	creds, err := google.CredentialsFromJSONWithType(ctx, jsonData, credType, SecretManagerScope)
	if err != nil {
		return nil, fmt.Errorf("gsm lookup plugin: unable to build credentials: %w", err)
	}

	return creds.TokenSource, nil
}
