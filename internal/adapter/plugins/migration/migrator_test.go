package migration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/esnet/gdg/internal/adapter/grafana/resources"
	"github.com/esnet/gdg/internal/adapter/plugins/migration"
	"github.com/esnet/gdg/internal/adapter/plugins/secure/noop"
	"github.com/esnet/gdg/internal/adapter/storage"
	"github.com/esnet/gdg/internal/config/config_domain"
	"github.com/esnet/gdg/internal/domain"
	"github.com/esnet/gdg/internal/ports/outbound"
	"github.com/matryer/is"
	"gopkg.in/yaml.v3"
)

// ---------------------------------------------------------------------------
// Test encoder: wraps each value/blob with a deterministic prefix so we can
// verify encode/decode without a real WASM plugin.
// ---------------------------------------------------------------------------

type prefixEncoder struct{ prefix string }

func (e prefixEncoder) EncodeValue(s string) (string, error) { return e.prefix + s, nil }
func (e prefixEncoder) DecodeValue(s string) (string, error) {
	after, found := strings.CutPrefix(s, e.prefix)
	if !found {
		return "", fmt.Errorf("expected prefix %q not found in %q", e.prefix, s)
	}
	return after, nil
}
func (e prefixEncoder) Encode(_ domain.ResourceType, b []byte) ([]byte, error) {
	return append([]byte(e.prefix), b...), nil
}
func (e prefixEncoder) Decode(_ domain.ResourceType, b []byte) ([]byte, error) {
	s := string(b)
	after, found := strings.CutPrefix(s, e.prefix)
	if !found {
		return nil, fmt.Errorf("expected prefix %q not found", e.prefix)
	}
	return []byte(after), nil
}

// ---------------------------------------------------------------------------
// Mock storage: in-memory storage that reports an arbitrary Name() value.
// Used to simulate cloud storage backends.
// ---------------------------------------------------------------------------

type mockStorage struct {
	name  string
	files map[string][]byte
}

func newMockStorage(name string) *mockStorage {
	return &mockStorage{name: name, files: make(map[string][]byte)}
}

func (m *mockStorage) Name() string      { return m.name }
func (m *mockStorage) GetPrefix() string { return "" }
func (m *mockStorage) WriteFile(filename string, data []byte) error {
	m.files[filename] = data
	return nil
}
func (m *mockStorage) FindAllFiles(_ string, _ bool) ([]string, error) { return nil, nil }
func (m *mockStorage) ReadFile(filename string) ([]byte, error) {
	data, ok := m.files[filename]
	if !ok {
		return nil, fmt.Errorf("file not found: %s", filename)
	}
	return data, nil
}

// Ensure mockStorage satisfies the interface.
var _ outbound.Storage = (*mockStorage)(nil)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// newGrafanaConfig returns a minimal GrafanaConfig with the given output path
// and the default organisation name so that org-scoped paths work.
func newGrafanaConfig(outputPath string) *config_domain.GrafanaConfig {
	return &config_domain.GrafanaConfig{
		OutputPath:       outputPath,
		OrganizationName: config_domain.DefaultOrganizationName,
	}
}

// contactsPath returns the expected contact point file path for a given output root.
func contactsPath(t *testing.T, outputPath string) string {
	t.Helper()
	cfg := newGrafanaConfig(outputPath)
	r := resources.NewHelpers()
	return r.BuildResourcePath(cfg, "contacts", domain.AlertingResource, false, false)
}

// writeFile writes data to path, creating all parent directories.
func writeFile(t *testing.T, path string, data []byte) {
	t.Helper()
	is := is.New(t)
	is.NoErr(os.MkdirAll(filepath.Dir(path), 0o750))
	is.NoErr(os.WriteFile(path, data, 0o600))
}

// ---------------------------------------------------------------------------
// Contact point tests
// ---------------------------------------------------------------------------

func TestRekeyContactPoints_Success(t *testing.T) {
	is := is.New(t)
	tmp := t.TempDir()

	oldEnc := prefixEncoder{"OLD:"}
	newEnc := prefixEncoder{"NEW:"}

	plaintext := []byte(`[{"uid":"abc","name":"test"}]`)
	encoded, err := oldEnc.Encode(domain.AlertingResource, plaintext)
	is.NoErr(err)

	path := contactsPath(t, tmp)
	writeFile(t, path, encoded)

	cfg := newGrafanaConfig(tmp)
	stor := storage.NewLocalStorage(context.Background())

	m := migration.NewMigrator(oldEnc, newEnc, cfg, stor, resources.NewHelpers())
	report, err := m.Rekey(migration.RekeyOptions{NoBackup: true})
	is.NoErr(err)
	is.Equal(len(report.Errors), 0)
	is.Equal(len(report.ContactPointsFiles), 1)

	// File should now be encoded with the new encoder.
	result, err := os.ReadFile(path)
	is.NoErr(err)
	is.True(strings.HasPrefix(string(result), "NEW:"))

	// Decode with new encoder to verify round-trip integrity.
	decoded, err := newEnc.Decode(domain.AlertingResource, result)
	is.NoErr(err)
	is.Equal(string(decoded), string(plaintext))
}

func TestRekeyContactPoints_FromPlaintext(t *testing.T) {
	is := is.New(t)
	tmp := t.TempDir()

	plaintext := []byte(`[{"uid":"abc","name":"test"}]`)
	path := contactsPath(t, tmp)
	writeFile(t, path, plaintext)

	cfg := newGrafanaConfig(tmp)
	stor := storage.NewLocalStorage(context.Background())
	newEnc := prefixEncoder{"ENC:"}

	m := migration.Migrator{
		OldEncoder:  noop.NoOpEncoder{},
		NewEncoder:  newEnc,
		GrafanaConf: cfg,
		Storage:     stor,
	}
	report, err := m.Rekey(migration.RekeyOptions{NoBackup: true})
	is.NoErr(err)
	is.Equal(len(report.Errors), 0)
	is.Equal(len(report.ContactPointsFiles), 1)

	result, err := os.ReadFile(path)
	is.NoErr(err)
	is.True(strings.HasPrefix(string(result), "ENC:"))
}

func TestRekeyContactPoints_ToPlaintext(t *testing.T) {
	is := is.New(t)
	tmp := t.TempDir()

	oldEnc := prefixEncoder{"ENC:"}
	plaintext := []byte(`[{"uid":"abc"}]`)
	encoded, _ := oldEnc.Encode(domain.AlertingResource, plaintext)

	path := contactsPath(t, tmp)
	writeFile(t, path, encoded)

	cfg := newGrafanaConfig(tmp)
	stor := storage.NewLocalStorage(context.Background())

	m := migration.Migrator{
		OldEncoder:  oldEnc,
		NewEncoder:  noop.NoOpEncoder{},
		GrafanaConf: cfg,
		Storage:     stor,
	}
	report, err := m.Rekey(migration.RekeyOptions{NoBackup: true})
	is.NoErr(err)
	is.Equal(len(report.Errors), 0)

	result, err := os.ReadFile(path)
	is.NoErr(err)
	// noop encoder returns plaintext unchanged.
	is.Equal(string(result), string(plaintext))
}

func TestRekeyContactPoints_FileNotFound_IsSkipped(t *testing.T) {
	is := is.New(t)
	tmp := t.TempDir()

	cfg := newGrafanaConfig(tmp)
	stor := storage.NewLocalStorage(context.Background())

	m := migration.Migrator{
		OldEncoder:  prefixEncoder{"OLD:"},
		NewEncoder:  prefixEncoder{"NEW:"},
		GrafanaConf: cfg,
		Storage:     stor,
	}
	// File was never created — should not produce an error.
	report, err := m.Rekey(migration.RekeyOptions{NoBackup: true})
	is.NoErr(err)
	is.Equal(len(report.Errors), 0)
	is.Equal(len(report.ContactPointsFiles), 0)
}

func TestRekeyContactPoints_CloudStorage_IsSkipped(t *testing.T) {
	is := is.New(t)
	tmp := t.TempDir()

	cloudStore := newMockStorage("S3Storage")
	path := contactsPath(t, tmp)
	// Populate the in-memory store so we can verify it is NOT changed.
	original := []byte("OLD:some data")
	cloudStore.files[path] = original

	cfg := newGrafanaConfig(tmp)
	m := migration.Migrator{
		OldEncoder:  prefixEncoder{"OLD:"},
		NewEncoder:  prefixEncoder{"NEW:"},
		GrafanaConf: cfg,
		Storage:     cloudStore,
	}
	report, err := m.Rekey(migration.RekeyOptions{NoBackup: true})
	is.NoErr(err)
	is.Equal(len(report.Errors), 0)
	is.Equal(len(report.ContactPointsFiles), 0)
	// The in-memory store must be untouched.
	is.Equal(string(cloudStore.files[path]), "OLD:some data")
}

// ---------------------------------------------------------------------------
// SecureData file tests
// ---------------------------------------------------------------------------

func TestRekeySecureDataFiles_YAML(t *testing.T) {
	is := is.New(t)
	tmp := t.TempDir()

	oldEnc := prefixEncoder{"OLD:"}
	newEnc := prefixEncoder{"NEW:"}

	// Write a YAML credential file with encoded values.
	secureDir := filepath.Join(tmp, "secure")
	is.NoErr(os.MkdirAll(secureDir, 0o750))
	creds := map[string]string{
		"user":     oldEnc.mustEncodeValue("alice"),
		"password": oldEnc.mustEncodeValue("secret123"),
	}
	raw, err := yaml.Marshal(creds)
	is.NoErr(err)
	credFile := filepath.Join(secureDir, "my_creds.yml")
	is.NoErr(os.WriteFile(credFile, raw, 0o600))

	cfg := newGrafanaConfigWithRule(tmp, "my_creds.yml")
	stor := storage.NewLocalStorage(context.Background())

	m := migration.Migrator{
		OldEncoder:  oldEnc,
		NewEncoder:  newEnc,
		GrafanaConf: cfg,
		Storage:     stor,
	}
	report, err := m.Rekey(migration.RekeyOptions{NoBackup: true})
	is.NoErr(err)
	is.Equal(len(report.Errors), 0)
	is.Equal(len(report.SecureDataFiles), 1)

	// Read the migrated file and check each value uses the new prefix.
	updated := make(map[string]string)
	data, err := os.ReadFile(credFile)
	is.NoErr(err)
	is.NoErr(yaml.Unmarshal(data, updated))
	for k, v := range updated {
		is.True(strings.HasPrefix(v, "NEW:")) // key should be re-encoded
		decoded, err := newEnc.DecodeValue(v)
		is.NoErr(err)
		original, err := oldEnc.DecodeValue(creds[k])
		is.NoErr(err)
		is.Equal(decoded, original) // round-trip must preserve plaintext
	}
}

func TestRekeySecureDataFiles_JSON(t *testing.T) {
	is := is.New(t)
	tmp := t.TempDir()

	oldEnc := prefixEncoder{"ENC:"}
	newEnc := noop.NoOpEncoder{} // remove encryption

	secureDir := filepath.Join(tmp, "secure")
	is.NoErr(os.MkdirAll(secureDir, 0o750))
	creds := map[string]string{
		"token": "ENC:supersecret",
	}
	raw, err := json.Marshal(creds)
	is.NoErr(err)
	credFile := filepath.Join(secureDir, "cloud_auth.json")
	is.NoErr(os.WriteFile(credFile, raw, 0o600))

	cfg := newGrafanaConfigWithRule(tmp, "cloud_auth.json")
	stor := storage.NewLocalStorage(context.Background())

	m := migration.Migrator{
		OldEncoder:  oldEnc,
		NewEncoder:  newEnc,
		GrafanaConf: cfg,
		Storage:     stor,
	}
	report, err := m.Rekey(migration.RekeyOptions{NoBackup: true})
	is.NoErr(err)
	is.Equal(len(report.Errors), 0)

	result := make(map[string]string)
	data, err := os.ReadFile(credFile)
	is.NoErr(err)
	is.NoErr(json.Unmarshal(data, &result))
	is.Equal(result["token"], "supersecret") // noop decrypts to plaintext
}

func TestRekeySecureDataFiles_Deduplicates(t *testing.T) {
	is := is.New(t)
	tmp := t.TempDir()

	oldEnc := prefixEncoder{"OLD:"}
	newEnc := prefixEncoder{"NEW:"}

	secureDir := filepath.Join(tmp, "secure")
	is.NoErr(os.MkdirAll(secureDir, 0o750))
	creds := map[string]string{"key": "OLD:value"}
	raw, _ := yaml.Marshal(creds)
	credFile := filepath.Join(secureDir, "shared.yml")
	is.NoErr(os.WriteFile(credFile, raw, 0o600))

	// Two rules pointing to the same file.
	cfg := newGrafanaConfigWithRules(tmp, "shared.yml", "shared.yml")
	stor := storage.NewLocalStorage(context.Background())

	m := migration.Migrator{
		OldEncoder:  oldEnc,
		NewEncoder:  newEnc,
		GrafanaConf: cfg,
		Storage:     stor,
	}
	report, err := m.Rekey(migration.RekeyOptions{NoBackup: true})
	is.NoErr(err)
	is.Equal(len(report.Errors), 0)
	// File should only appear once despite two rules pointing to it.
	is.Equal(len(report.SecureDataFiles), 1)
}

func TestRekeySecureDataFiles_UnsupportedExtension(t *testing.T) {
	is := is.New(t)
	tmp := t.TempDir()

	secureDir := filepath.Join(tmp, "secure")
	is.NoErr(os.MkdirAll(secureDir, 0o750))
	credFile := filepath.Join(secureDir, "creds.toml")
	is.NoErr(os.WriteFile(credFile, []byte(`key = "value"`), 0o600))

	cfg := newGrafanaConfigWithRule(tmp, "creds.toml")
	stor := storage.NewLocalStorage(context.Background())

	m := migration.Migrator{
		OldEncoder:  noop.NoOpEncoder{},
		NewEncoder:  prefixEncoder{"ENC:"},
		GrafanaConf: cfg,
		Storage:     stor,
	}
	report, err := m.Rekey(migration.RekeyOptions{NoBackup: true})
	is.NoErr(err)
	// Unsupported extension should be reported as an error.
	is.Equal(len(report.Errors), 1)
	is.Equal(len(report.SecureDataFiles), 0)
}

// ---------------------------------------------------------------------------
// GDG credentials tests
// ---------------------------------------------------------------------------

func TestRekeyGdgCredentials_YAML(t *testing.T) {
	is := is.New(t)
	tmp := t.TempDir()

	oldEnc := prefixEncoder{"OLD:"}
	newEnc := prefixEncoder{"NEW:"}

	// Write a YAML auth file.
	cfg := newGrafanaConfig(tmp)
	authBase := cfg.GetAuthLocation()
	is.NoErr(os.MkdirAll(filepath.Dir(authBase), 0o750))
	sm := config_domain.SecureModel{
		Password: "OLD:mypassword",
		Token:    "OLD:mytoken",
	}
	raw, err := yaml.Marshal(sm)
	is.NoErr(err)
	is.NoErr(os.WriteFile(authBase+".yaml", raw, 0o600))

	stor := storage.NewLocalStorage(context.Background())
	m := migration.Migrator{
		OldEncoder:  oldEnc,
		NewEncoder:  newEnc,
		GrafanaConf: cfg,
		Storage:     stor,
	}
	report, err := m.Rekey(migration.RekeyOptions{
		NoBackup:              true,
		IncludeGdgCredentials: true,
	})
	is.NoErr(err)
	is.Equal(len(report.Errors), 0)
	is.True(report.GdgCredentialsMigrated)

	// Check the written file.
	data, err := os.ReadFile(authBase + ".yaml")
	is.NoErr(err)
	var result config_domain.SecureModel
	is.NoErr(yaml.Unmarshal(data, &result))
	is.True(strings.HasPrefix(result.Password, "NEW:"))
	is.True(strings.HasPrefix(result.Token, "NEW:"))
}

func TestRekeyGdgCredentials_NoFile_IsSkipped(t *testing.T) {
	is := is.New(t)
	tmp := t.TempDir()

	cfg := newGrafanaConfig(tmp)
	stor := storage.NewLocalStorage(context.Background())
	m := migration.Migrator{
		OldEncoder:  prefixEncoder{"OLD:"},
		NewEncoder:  prefixEncoder{"NEW:"},
		GrafanaConf: cfg,
		Storage:     stor,
	}
	report, err := m.Rekey(migration.RekeyOptions{
		NoBackup:              true,
		IncludeGdgCredentials: true,
	})
	is.NoErr(err)
	is.Equal(len(report.Errors), 0)
	is.True(!report.GdgCredentialsMigrated)
}

func TestRekeyGdgCredentials_ExcludedByDefault(t *testing.T) {
	is := is.New(t)
	tmp := t.TempDir()

	cfg := newGrafanaConfig(tmp)
	// Write a valid auth file.
	authBase := cfg.GetAuthLocation()
	is.NoErr(os.MkdirAll(filepath.Dir(authBase), 0o750))
	sm := config_domain.SecureModel{Password: "OLD:pw"}
	raw, _ := yaml.Marshal(sm)
	is.NoErr(os.WriteFile(authBase+".yaml", raw, 0o600))

	stor := storage.NewLocalStorage(context.Background())
	m := migration.Migrator{
		OldEncoder:  prefixEncoder{"OLD:"},
		NewEncoder:  prefixEncoder{"NEW:"},
		GrafanaConf: cfg,
		Storage:     stor,
	}
	// IncludeGdgCredentials defaults to false — file should not be touched.
	report, err := m.Rekey(migration.RekeyOptions{NoBackup: true})
	is.NoErr(err)
	is.True(!report.GdgCredentialsMigrated)

	// File should be unchanged.
	data, err := os.ReadFile(authBase + ".yaml")
	is.NoErr(err)
	var result config_domain.SecureModel
	is.NoErr(yaml.Unmarshal(data, &result))
	is.Equal(result.Password, "OLD:pw")
}

// ---------------------------------------------------------------------------
// Backup tests
// ---------------------------------------------------------------------------

func TestRekey_BackupCreated(t *testing.T) {
	is := is.New(t)
	tmp := t.TempDir()
	backupDir := t.TempDir()

	oldEnc := prefixEncoder{"OLD:"}
	newEnc := prefixEncoder{"NEW:"}

	// Set up a contact points file.
	plaintext := []byte(`[{"uid":"abc"}]`)
	encoded, _ := oldEnc.Encode(domain.AlertingResource, plaintext)
	path := contactsPath(t, tmp)
	writeFile(t, path, encoded)

	cfg := newGrafanaConfig(tmp)
	stor := storage.NewLocalStorage(context.Background())

	m := migration.Migrator{
		OldEncoder:  oldEnc,
		NewEncoder:  newEnc,
		GrafanaConf: cfg,
		Storage:     stor,
	}
	report, err := m.Rekey(migration.RekeyOptions{
		NoBackup:  false,
		BackupDir: backupDir,
	})
	is.NoErr(err)
	is.Equal(len(report.Errors), 0)
	is.Equal(report.BackupDir, backupDir)

	// A backup of the contacts file must exist.
	// The backup mirrors the full path under backupDir.
	rel := strings.TrimPrefix(path, string(filepath.Separator))
	backupPath := filepath.Join(backupDir, rel)
	backup, err := os.ReadFile(backupPath)
	is.NoErr(err)
	is.Equal(string(backup), string(encoded)) // backup holds the OLD content
}

func TestRekey_AutoBackupDir(t *testing.T) {
	is := is.New(t)
	tmp := t.TempDir()

	oldEnc := prefixEncoder{"OLD:"}
	newEnc := prefixEncoder{"NEW:"}

	plaintext := []byte(`[{"uid":"abc"}]`)
	encoded, _ := oldEnc.Encode(domain.AlertingResource, plaintext)
	path := contactsPath(t, tmp)
	writeFile(t, path, encoded)

	cfg := newGrafanaConfig(tmp)
	stor := storage.NewLocalStorage(context.Background())

	m := migration.Migrator{
		OldEncoder:  oldEnc,
		NewEncoder:  newEnc,
		GrafanaConf: cfg,
		Storage:     stor,
	}
	// BackupDir is empty → must be auto-generated.
	report, err := m.Rekey(migration.RekeyOptions{NoBackup: false})
	is.NoErr(err)
	is.True(report.BackupDir != "")
	// Verify the auto-generated dir exists.
	_, err = os.Stat(report.BackupDir)
	is.NoErr(err)
}

func TestRekey_NoBackup(t *testing.T) {
	is := is.New(t)
	tmp := t.TempDir()
	backupDir := t.TempDir()

	oldEnc := prefixEncoder{"OLD:"}
	newEnc := prefixEncoder{"NEW:"}

	plaintext := []byte(`[{"uid":"abc"}]`)
	encoded, _ := oldEnc.Encode(domain.AlertingResource, plaintext)
	path := contactsPath(t, tmp)
	writeFile(t, path, encoded)

	cfg := newGrafanaConfig(tmp)
	stor := storage.NewLocalStorage(context.Background())

	m := migration.Migrator{
		OldEncoder:  oldEnc,
		NewEncoder:  newEnc,
		GrafanaConf: cfg,
		Storage:     stor,
	}
	report, err := m.Rekey(migration.RekeyOptions{
		NoBackup:  true,
		BackupDir: backupDir, // ignored when NoBackup is true
	})
	is.NoErr(err)
	is.Equal(report.BackupDir, "") // not set in report when NoBackup is true

	// The explicitly supplied backupDir must remain empty.
	entries, err := os.ReadDir(backupDir)
	is.NoErr(err)
	is.Equal(len(entries), 0)
}

// ---------------------------------------------------------------------------
// Full migration integration test
// ---------------------------------------------------------------------------

func TestRekey_AllCategories(t *testing.T) {
	is := is.New(t)
	tmp := t.TempDir()

	oldEnc := prefixEncoder{"OLD:"}
	newEnc := prefixEncoder{"NEW:"}

	// 1. Contact points
	plainCP := []byte(`[{"uid":"1"}]`)
	encCP, _ := oldEnc.Encode(domain.AlertingResource, plainCP)
	cpPath := contactsPath(t, tmp)
	writeFile(t, cpPath, encCP)

	// 2. SecureData file
	secureDir := filepath.Join(tmp, "secure")
	is.NoErr(os.MkdirAll(secureDir, 0o750))
	creds := map[string]string{"user": "OLD:alice", "pass": "OLD:secret"}
	credRaw, _ := yaml.Marshal(creds)
	credFile := filepath.Join(secureDir, "creds.yml")
	is.NoErr(os.WriteFile(credFile, credRaw, 0o600))

	// 3. Auth file
	cfg := newGrafanaConfigWithRule(tmp, "creds.yml")
	authBase := cfg.GetAuthLocation()
	is.NoErr(os.MkdirAll(filepath.Dir(authBase), 0o750))
	sm := config_domain.SecureModel{Password: "OLD:pass", Token: "OLD:tok"}
	smRaw, _ := yaml.Marshal(sm)
	is.NoErr(os.WriteFile(authBase+".yaml", smRaw, 0o600))

	stor := storage.NewLocalStorage(context.Background())
	m := migration.Migrator{
		OldEncoder:  oldEnc,
		NewEncoder:  newEnc,
		GrafanaConf: cfg,
		Storage:     stor,
	}
	report, err := m.Rekey(migration.RekeyOptions{
		NoBackup:              true,
		IncludeGdgCredentials: true,
	})
	is.NoErr(err)
	is.Equal(len(report.Errors), 0)
	is.Equal(len(report.ContactPointsFiles), 1)
	is.Equal(len(report.SecureDataFiles), 1)
	is.True(report.GdgCredentialsMigrated)

	// Verify contact points file is now NEW-encoded.
	cpData, _ := os.ReadFile(cpPath)
	is.True(strings.HasPrefix(string(cpData), "NEW:"))

	// Verify SecureData file values are now NEW-encoded.
	updatedCreds := make(map[string]string)
	credData, _ := os.ReadFile(credFile)
	is.NoErr(yaml.Unmarshal(credData, updatedCreds))
	for _, v := range updatedCreds {
		is.True(strings.HasPrefix(v, "NEW:"))
	}

	// Verify auth file values are now NEW-encoded.
	var updatedSM config_domain.SecureModel
	authData, _ := os.ReadFile(authBase + ".yaml")
	is.NoErr(yaml.Unmarshal(authData, &updatedSM))
	is.True(strings.HasPrefix(updatedSM.Password, "NEW:"))
	is.True(strings.HasPrefix(updatedSM.Token, "NEW:"))
}

// ---------------------------------------------------------------------------
// Helper constructors used only in tests
// ---------------------------------------------------------------------------

// newGrafanaConfigWithRule returns a GrafanaConfig with one credential rule
// referencing the supplied secure data filename.
func newGrafanaConfigWithRule(outputPath, secureDataFile string) *config_domain.GrafanaConfig {
	return newGrafanaConfigWithRules(outputPath, secureDataFile)
}

// newGrafanaConfigWithRules returns a GrafanaConfig with one credential rule
// per supplied filename.
func newGrafanaConfigWithRules(outputPath string, secureDataFiles ...string) *config_domain.GrafanaConfig {
	rules := make([]*config_domain.RegexMatchesList, 0, len(secureDataFiles))
	for _, f := range secureDataFiles {
		rules = append(rules, &config_domain.RegexMatchesList{
			SecureData: f,
		})
	}
	return &config_domain.GrafanaConfig{
		OutputPath:       outputPath,
		OrganizationName: config_domain.DefaultOrganizationName,
		ConnectionSettings: &config_domain.ConnectionSettings{
			MatchingRules: rules,
		},
	}
}

// mustEncodeValue is a test helper that panics on encoding errors.
func (e prefixEncoder) mustEncodeValue(s string) string {
	v, err := e.EncodeValue(s)
	if err != nil {
		panic(err)
	}
	return v
}

// ---------------------------------------------------------------------------
// Dry-run tests
// ---------------------------------------------------------------------------

func TestRekey_DryRun_ContactPoints_PopulatesPreviews(t *testing.T) {
	is := is.New(t)
	tmp := t.TempDir()

	oldEnc := prefixEncoder{"OLD:"}
	plaintext := []byte(`[{"uid":"abc","name":"test"}]`)
	encoded, _ := oldEnc.Encode(domain.AlertingResource, plaintext)

	path := contactsPath(t, tmp)
	writeFile(t, path, encoded)

	cfg := newGrafanaConfig(tmp)
	stor := storage.NewLocalStorage(context.Background())

	m := migration.Migrator{
		OldEncoder:  oldEnc,
		NewEncoder:  prefixEncoder{"NEW:"},
		GrafanaConf: cfg,
		Storage:     stor,
	}
	report, err := m.Rekey(migration.RekeyOptions{DryRun: true})
	is.NoErr(err)

	// DryRun must not modify the file.
	result, _ := os.ReadFile(path)
	is.Equal(string(result), string(encoded)) // file unchanged

	// Report must contain a FilePreview for the contacts file.
	is.Equal(len(report.ContactPointsFiles), 0) // no writes
	is.Equal(len(report.Previews), 1)
	is.Equal(report.Previews[0].Category, "contact_points")
	is.Equal(report.Previews[0].Path, path)
	is.True(report.Previews[0].DecodedOK)
}

func TestRekey_DryRun_SecureData_PopulatesPreviews(t *testing.T) {
	is := is.New(t)
	tmp := t.TempDir()

	oldEnc := prefixEncoder{"OLD:"}
	secureDir := filepath.Join(tmp, "secure")
	is.NoErr(os.MkdirAll(secureDir, 0o750))

	creds := map[string]string{"user": "OLD:alice", "pass": "OLD:secret"}
	credRaw, _ := yaml.Marshal(creds)
	credFile := filepath.Join(secureDir, "creds.yml")
	writeFile(t, credFile, credRaw)

	cfg := newGrafanaConfigWithRule(tmp, "creds.yml")
	stor := storage.NewLocalStorage(context.Background())

	m := migration.Migrator{
		OldEncoder:  oldEnc,
		NewEncoder:  prefixEncoder{"NEW:"},
		GrafanaConf: cfg,
		Storage:     stor,
	}
	report, err := m.Rekey(migration.RekeyOptions{DryRun: true})
	is.NoErr(err)

	// File must be unchanged.
	result, _ := os.ReadFile(credFile)
	is.Equal(string(result), string(credRaw))

	// One FilePreview for the creds file.
	is.Equal(len(report.SecureDataFiles), 0)
	is.Equal(len(report.Previews), 1)
	is.Equal(report.Previews[0].Category, "secure_data")
	is.True(report.Previews[0].DecodedOK)
	is.Equal(len(report.Previews[0].Keys), 2) // "pass" and "user"
}

func TestRekey_AllowList_RestrictsProcessing(t *testing.T) {
	is := is.New(t)
	tmp := t.TempDir()

	oldEnc := prefixEncoder{"OLD:"}
	newEnc := prefixEncoder{"NEW:"}

	secureDir := filepath.Join(tmp, "secure")
	is.NoErr(os.MkdirAll(secureDir, 0o750))

	file1 := filepath.Join(secureDir, "creds1.yml")
	file2 := filepath.Join(secureDir, "creds2.yml")
	raw1, _ := yaml.Marshal(map[string]string{"key": "OLD:val1"})
	raw2, _ := yaml.Marshal(map[string]string{"key": "OLD:val2"})
	writeFile(t, file1, raw1)
	writeFile(t, file2, raw2)

	cfg := newGrafanaConfigWithRules(tmp, "creds1.yml", "creds2.yml")
	stor := storage.NewLocalStorage(context.Background())

	m := migration.Migrator{
		OldEncoder:  oldEnc,
		NewEncoder:  newEnc,
		GrafanaConf: cfg,
		Storage:     stor,
	}
	// Only allow file1.
	report, err := m.Rekey(migration.RekeyOptions{
		NoBackup:  true,
		AllowList: []string{file1},
	})
	is.NoErr(err)
	is.Equal(len(report.Errors), 0)

	// file1 must be re-encrypted; file2 must be untouched.
	is.Equal(len(report.SecureDataFiles), 1)
	is.Equal(report.SecureDataFiles[0], file1)

	result1, _ := os.ReadFile(file1)
	result2, _ := os.ReadFile(file2)
	is.True(strings.Contains(string(result1), "NEW:")) // re-encrypted
	is.Equal(string(result2), string(raw2))            // unchanged
}
