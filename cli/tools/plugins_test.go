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
	{
		Name:        "gsm-secretmanager",
		Type:        gdgdomain.PluginTypeLookup,
		Description: "Google Secret Manager lookup plugin",
		Source:      "https://example.com/gsm",
		URLPattern:  "https://example.com/releases/{version}/lookup_gsm.wasm",
		Versions: []gdgdomain.PluginVersionEntry{
			{Version: "0.1.0", ConfigFields: []string{"credentials"}},
		},
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

// ── plugins rekey ─────────────────────────────────────────────────────────────

// TestPluginsParentHelp_ListsRekeySubcommand verifies that the "rekey"
// subcommand is registered and appears in the "plugins" parent help output.
func TestPluginsParentHelp_ListsRekeySubcommand(t *testing.T) {
	is := is.New(t)

	rootSvc := cli.NewRootService()
	execMe := func(_ *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		return cli.Execute(rootSvc, []string{"tools", "plugins"}, optionMockSvc())
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	lower := strings.ToLower(outStr)
	is.True(strings.Contains(lower, "rekey"))
}

// TestPluginsRekeyHelpShowsFlags verifies that the rekey command exposes the
// expected --registry-file and --registry-url flags (checked via --help, which
// does not invoke RunFunc and therefore does not launch the TUI).
func TestPluginsRekeyHelpShowsFlags(t *testing.T) {
	is := is.New(t)

	rootSvc := cli.NewRootService()
	execMe := func(_ *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		return cli.Execute(rootSvc, []string{"tools", "plugins", "rekey", "--help"}, optionMockSvc())
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	is.True(strings.Contains(outStr, "--registry-file"))
	is.True(strings.Contains(outStr, "--registry-url"))
}

// TestPluginsRekeyHelpDescribesMigration verifies that the rekey command's
// help text communicates its purpose (re-encrypting files / cipher migration).
func TestPluginsRekeyHelpDescribesMigration(t *testing.T) {
	is := is.New(t)

	rootSvc := cli.NewRootService()
	execMe := func(_ *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		return cli.Execute(rootSvc, []string{"tools", "plugins", "rekey", "--help"}, optionMockSvc())
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	lower := strings.ToLower(outStr)
	is.True(strings.Contains(lower, "re-encrypt") || strings.Contains(lower, "cipher"))
}

// ── plugins cipher encode/decode (--value path) ────────────────────────────────
//
// Moved here from "tools helpers cipher" so cipher-plugin interaction
// (list, rekey, and now encode/decode) all lives under "tools plugins".

func TestPluginsCipherEncodeByValue(t *testing.T) {
	is := is.New(t)
	rootSvc := cli.NewRootService()
	execMe := func(mock *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		mock.EXPECT().EncodeValue("super-secret").Return("cipher:encoded-value")
		return cli.Execute(rootSvc, []string{"tools", "plugins", "cipher", "encode", "--value", "super-secret"}, optionMockSvc())
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	is.True(strings.Contains(outStr, "cipher:encoded-value"))
}

func TestPluginsCipherDecodeByValue(t *testing.T) {
	is := is.New(t)
	rootSvc := cli.NewRootService()
	execMe := func(mock *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		mock.EXPECT().DecodeValue("cipher:encoded-value").Return("super-secret")
		return cli.Execute(rootSvc, []string{"tools", "plugins", "cipher", "decode", "--value", "cipher:encoded-value"}, optionMockSvc())
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	is.True(strings.Contains(outStr, "super-secret"))
}

func TestPluginsCipherAlias(t *testing.T) {
	is := is.New(t)
	rootSvc := cli.NewRootService()
	execMe := func(_ *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		return cli.Execute(rootSvc, []string{"tools", "plugins", "c"}, optionMockSvc())
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	lower := strings.ToLower(outStr)
	is.True(strings.Contains(lower, "encode"))
	is.True(strings.Contains(lower, "decode"))
}

// ── plugins list --type ────────────────────────────────────────────────────────

func TestPluginsListTypeLookupShowsOnlyLookupPlugins(t *testing.T) {
	is := is.New(t)
	regFile := writeRegistryFixture(t)

	rootSvc := cli.NewRootService()
	execMe := func(_ *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		return cli.Execute(rootSvc,
			[]string{"tools", "plugins", "list", "--registry-file", regFile, "--type", "lookup"},
			optionMockSvc())
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	lower := strings.ToLower(outStr)
	is.True(strings.Contains(lower, "gsm-secretmanager"))
	is.True(strings.Contains(lower, "credentials"))
	is.True(!strings.Contains(outStr, "aes-256-gcm"))
	is.True(!strings.Contains(outStr, "ansible-vault"))
}

func TestPluginsListTypeAllShowsEveryType(t *testing.T) {
	is := is.New(t)
	regFile := writeRegistryFixture(t)

	rootSvc := cli.NewRootService()
	execMe := func(_ *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		return cli.Execute(rootSvc,
			[]string{"tools", "plugins", "list", "--registry-file", regFile, "--type", "all"},
			optionMockSvc())
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	is.True(strings.Contains(outStr, "aes-256-gcm"))
	is.True(strings.Contains(outStr, "gsm-secretmanager"))
	is.True(strings.Contains(outStr, "future-tool"))
}

func TestPluginsListDefaultTypeIsCipher(t *testing.T) {
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

	is.True(strings.Contains(outStr, "aes-256-gcm"))
	is.True(!strings.Contains(outStr, "gsm-secretmanager"))
}

func TestPluginsListUnknownTypeFails(t *testing.T) {
	is := is.New(t)
	regFile := writeRegistryFixture(t)

	rootSvc := cli.NewRootService()
	var execErr error
	execMe := func(_ *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		execErr = cli.Execute(rootSvc,
			[]string{"tools", "plugins", "list", "--registry-file", regFile, "--type", "bogus"},
			optionMockSvc())
		// Return nil so SetupAndExecuteMockingServices' assert.Nil doesn't fail the test;
		// we inspect execErr directly.
		return nil
	}
	_, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	is.True(execErr != nil)
}

// ── plugins lookup ────────────────────────────────────────────────────────────

// TestPluginsParentHelp_ListsLookupSubcommand verifies that the "lookup"
// subcommand is registered and appears in the "plugins" parent help output.
func TestPluginsParentHelp_ListsLookupSubcommand(t *testing.T) {
	is := is.New(t)

	rootSvc := cli.NewRootService()
	execMe := func(_ *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		return cli.Execute(rootSvc, []string{"tools", "plugins"}, optionMockSvc())
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	lower := strings.ToLower(outStr)
	is.True(strings.Contains(lower, "lookup"))
}

// TestPluginsLookupHelp_ListsTestSubcommand verifies that "test" is
// registered under "gdg tools plugins lookup".
func TestPluginsLookupHelp_ListsTestSubcommand(t *testing.T) {
	is := is.New(t)

	rootSvc := cli.NewRootService()
	execMe := func(_ *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		return cli.Execute(rootSvc, []string{"tools", "plugins", "lookup"}, optionMockSvc())
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	lower := strings.ToLower(outStr)
	is.True(strings.Contains(lower, "test"))
}

// TestPluginsLookupTest_MissingArgsFails verifies the command requires both
// a provider and a key argument.
func TestPluginsLookupTest_MissingArgsFails(t *testing.T) {
	is := is.New(t)

	rootSvc := cli.NewRootService()
	var execErr error
	execMe := func(_ *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		execErr = cli.Execute(rootSvc, []string{"tools", "plugins", "lookup", "test", "gsm"}, optionMockSvc())
		return nil
	}
	_, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	is.True(execErr != nil)
}

// TestPluginsLookupTest_UnknownProviderFails verifies that resolving against
// a provider that isn't configured in gdg.yml (the "testing" fixture config
// has no plugins.lookup block at all) returns a clear error rather than
// requiring a live Grafana connection or a real lookup plugin.
func TestPluginsLookupTest_UnknownProviderFails(t *testing.T) {
	is := is.New(t)

	rootSvc := cli.NewRootService()
	var execErr error
	execMe := func(_ *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		execErr = cli.Execute(rootSvc,
			[]string{"tools", "plugins", "lookup", "test", "gsm", "projects/1/secrets/x/versions/1"},
			optionMockSvc())
		return nil
	}
	_, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	is.True(execErr != nil)
	is.True(strings.Contains(execErr.Error(), "gsm"))
}

// TestPluginsParentHelp_ListsCipherSubcommand verifies that the "cipher"
// subcommand is registered and appears in the "plugins" parent help output,
// alongside "list" and "rekey".
func TestPluginsParentHelp_ListsCipherSubcommand(t *testing.T) {
	is := is.New(t)

	rootSvc := cli.NewRootService()
	execMe := func(_ *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		return cli.Execute(rootSvc, []string{"tools", "plugins"}, optionMockSvc())
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	lower := strings.ToLower(outStr)
	is.True(strings.Contains(lower, "cipher"))
}
