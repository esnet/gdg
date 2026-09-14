package lookup_test

// Integration test for NewResolver dispatching to a real provider constructor.
// Lives in package lookup_test (external test package) so it can import
// lookup/gsm without creating an import cycle.

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/esnet/gdg/internal/adapter/plugins/lookup"
	"github.com/esnet/gdg/internal/adapter/plugins/lookup/gsm"
	"github.com/esnet/gdg/internal/config/config_domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// minimalWasmModule is the smallest possible valid WebAssembly binary.
var minimalWasmModule = []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}

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

func TestNewResolver_KnownProvider_DispatchesToRealConstructor(t *testing.T) {
	lookup.RegisterProvider("gsm", gsm.NewPluginLookupGSM)

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

	r, err := lookup.NewResolver(cfg)
	require.NoError(t, err, "expected buildProvider to dispatch \"gsm\" to gsm.NewPluginLookupGSM and succeed")
	assert.NotNil(t, r)
}
