package tools_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/esnet/gdg/cli"
	"github.com/esnet/gdg/cli/domain"
	gdgdomain "github.com/esnet/gdg/internal/domain"
	"github.com/esnet/gdg/internal/ports/outbound/mocks"
	"github.com/esnet/gdg/pkg/test_tooling"
	"github.com/matryer/is"
)

// registryFixture is the minimal registry JSON written to a temp file for tests
// that need to exercise the list command without network access.
var registryFixture = []gdgdomain.PluginRegistryEntry{
	{
		Name:        "aes-256-gcm",
		Type:        gdgdomain.PluginTypeCipher,
		Description: "seeded implementation of aes-256",
		Source:      "https://example.com/aes",
		URLPattern:  "https://example.com/releases/{version}/aes.wasm",
		Versions: []gdgdomain.PluginVersionEntry{
			{Version: "0.1.0", ConfigFields: []string{"passphrase"}},
		},
	},
	{
		Name:        "ansible-vault",
		Type:        gdgdomain.PluginTypeCipher,
		Description: "golang implementation of ansible-vault",
		Source:      "https://example.com/ansible",
		URLPattern:  "https://example.com/releases/{version}/ansible.wasm",
		Versions: []gdgdomain.PluginVersionEntry{
			{Version: "0.1.0", ConfigFields: []string{"vault_password"}},
		},
	},
	{
		// Non-cipher entry — must be filtered out of the list output.
		Name:     "future-tool",
		Type:     "future-type",
		Versions: []gdgdomain.PluginVersionEntry{{Version: "1.0.0"}},
	},
}

// writeRegistryFixture marshals registryFixture to a temp file and returns its path.
func writeRegistryFixture(t *testing.T) string {
	t.Helper()
	raw, err := json.Marshal(registryFixture)
	if err != nil {
		t.Fatalf("marshal registry fixture: %v", err)
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "registry.json")
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatalf("write registry fixture: %v", err)
	}
	return path
}

// ── plugins (parent) ──────────────────────────────────────────────────────────

func TestPluginsParentPrintsHelp(t *testing.T) {
	is := is.New(t)
	regFile := writeRegistryFixture(t)

	rootSvc := cli.NewRootService()
	execMe := func(_ *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		return cli.Execute(rootSvc, []string{"tools", "plugins"}, optionMockSvc())
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()
	_ = regFile // fixture not needed for the parent help command

	lower := strings.ToLower(outStr)
	is.True(strings.Contains(lower, "list"))
}

func TestPluginsAliasPlugin(t *testing.T) {
	is := is.New(t)
	rootSvc := cli.NewRootService()
	execMe := func(_ *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		return cli.Execute(rootSvc, []string{"tools", "plugin"}, optionMockSvc())
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	lower := strings.ToLower(outStr)
	is.True(strings.Contains(lower, "list"))
}

// ── plugins list ──────────────────────────────────────────────────────────────

func TestPluginsListShowsCipherPlugins(t *testing.T) {
	is := is.New(t)
	regFile := writeRegistryFixture(t)

	rootSvc := cli.NewRootService()
	execMe := func(_ *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		return cli.Execute(rootSvc,
			[]string{"tools", "plugins", "list", "--registry-file", regFile},
			optionMockSvc())
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	lower := strings.ToLower(outStr)
	is.True(strings.Contains(lower, "aes-256-gcm"))
	is.True(strings.Contains(lower, "ansible-vault"))
	is.True(strings.Contains(lower, "passphrase"))
	is.True(strings.Contains(lower, "vault_password"))
}

func TestPluginsListFiltersNonCipherTypes(t *testing.T) {
	is := is.New(t)
	regFile := writeRegistryFixture(t)

	rootSvc := cli.NewRootService()
	execMe := func(_ *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		return cli.Execute(rootSvc,
			[]string{"tools", "plugins", "list", "--registry-file", regFile},
			optionMockSvc())
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	is.True(!strings.Contains(outStr, "future-tool"))
}

func TestPluginsListShowsVersionColumn(t *testing.T) {
	is := is.New(t)
	regFile := writeRegistryFixture(t)

	rootSvc := cli.NewRootService()
	execMe := func(_ *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		return cli.Execute(rootSvc,
			[]string{"tools", "plugins", "list", "--registry-file", regFile},
			optionMockSvc())
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	is.True(strings.Contains(outStr, "0.1.0"))
}

func TestPluginsListMissingFileFails(t *testing.T) {
	is := is.New(t)
	rootSvc := cli.NewRootService()

	var execErr error
	execMe := func(_ *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		execErr = cli.Execute(rootSvc,
			[]string{"tools", "plugins", "list", "--registry-file", "/nonexistent/registry.json"},
			optionMockSvc())
		// Return nil so SetupAndExecuteMockingServices' assert.Nil doesn't fail the test;
		// we inspect execErr directly.
		return nil
	}
	_, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	is.True(execErr != nil)
}
