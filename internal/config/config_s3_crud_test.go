package config_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/esnet/gdg/internal/adapter/storage"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/config/config_domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newS3TestApp writes a minimal gdg.yml to a temp directory and loads it via
// NewConfig. The active context's output_path is set to a separate temp dir so
// SecureLocation() resolves to a controlled path within it.
func newS3TestApp(t *testing.T) (*config_domain.GDGAppConfiguration, string) {
	t.Helper()

	outputDir := t.TempDir()
	cfgDir := t.TempDir()
	cfgPath := filepath.Join(cfgDir, "gdg-s3test.yml")

	yaml := fmt.Sprintf(`
context_name: testing
contexts:
  testing:
    url: http://localhost:3000
    output_path: %s
    watched:
      - General
storage_engine: {}
`, outputDir)

	require.NoError(t, os.WriteFile(cfgPath, []byte(yaml), 0o600))

	app := config.NewConfig(cfgPath)
	return app, outputDir
}

// secureDir returns the path where credentials files are written for the testing context.
// Matches: <output_path>/secure  (SecureSecretsResource = "secure", non-namespaced)
func secureDir(outputDir string) string {
	return filepath.Join(outputDir, "secure")
}

// credFilePath returns the expected credentials file path for a given storage label.
func credFilePath(outputDir, label string) string {
	return filepath.Join(secureDir(outputDir), fmt.Sprintf("%s_%s.yaml", config_domain.CloudAuthPrefix, label))
}

// writeFakeCredFile creates the secure directory and a placeholder credentials file,
// simulating what NewCustomS3Config would write.
func writeFakeCredFile(t *testing.T, outputDir, label string) string {
	t.Helper()
	dir := secureDir(outputDir)
	require.NoError(t, os.MkdirAll(dir, 0o750))
	path := credFilePath(outputDir, label)
	require.NoError(t, os.WriteFile(path, []byte("access_id: test\nsecret_key: secret\n"), 0o600))
	return path
}

// ---------------------------------------------------------------------------
// ListS3Configs
// ---------------------------------------------------------------------------

func TestListS3Configs_Empty(t *testing.T) {
	app, _ := newS3TestApp(t)

	result := config.ListS3Configs(app)
	assert.NotNil(t, result)
	assert.Empty(t, result)
}

func TestListS3Configs_WithEngines(t *testing.T) {
	app, _ := newS3TestApp(t)

	app.StorageEngine = map[string]map[string]string{
		"minio-local": {
			storage.CloudType:  storage.Custom,
			storage.Endpoint:   "http://localhost:9000",
			storage.BucketName: "grafana",
			storage.Region:     "us-east-1",
		},
		"minio-remote": {
			storage.CloudType:  storage.Custom,
			storage.Endpoint:   "http://remote:9000",
			storage.BucketName: "grafana-remote",
			storage.Region:     "eu-west-1",
		},
	}

	result := config.ListS3Configs(app)
	assert.Len(t, result, 2)
	assert.Contains(t, result, "minio-local")
	assert.Contains(t, result, "minio-remote")
	assert.Equal(t, "http://localhost:9000", result["minio-local"][storage.Endpoint])
	assert.Equal(t, "eu-west-1", result["minio-remote"][storage.Region])
}

// ---------------------------------------------------------------------------
// DeleteS3Config
// ---------------------------------------------------------------------------

func TestDeleteS3Config_RemovesEngine(t *testing.T) {
	app, _ := newS3TestApp(t)

	app.StorageEngine = map[string]map[string]string{
		"target": {
			storage.CloudType:  storage.Custom,
			storage.BucketName: "mybucket",
		},
		"other": {
			storage.CloudType:  storage.Custom,
			storage.BucketName: "otherbucket",
		},
	}

	config.DeleteS3Config(app, "target")

	assert.NotContains(t, app.StorageEngine, "target", "deleted engine should be removed")
	assert.Contains(t, app.StorageEngine, "other", "unrelated engine should be preserved")
}

func TestDeleteS3Config_ClearsContextAssignment(t *testing.T) {
	app, _ := newS3TestApp(t)

	app.StorageEngine = map[string]map[string]string{
		"my-minio": {storage.CloudType: storage.Custom},
	}

	// Assign the storage engine to the active context
	app.GetDefaultGrafanaConfig().Storage = "my-minio"
	assert.Equal(t, "my-minio", app.GetDefaultGrafanaConfig().Storage)

	config.DeleteS3Config(app, "my-minio")

	assert.Empty(t, app.GetDefaultGrafanaConfig().Storage,
		"context's Storage field should be cleared after engine deletion")
}

func TestDeleteS3Config_RemovesCredFile(t *testing.T) {
	app, outputDir := newS3TestApp(t)

	app.StorageEngine = map[string]map[string]string{
		"my-minio": {storage.CloudType: storage.Custom},
	}

	credFile := writeFakeCredFile(t, outputDir, "my-minio")
	_, err := os.Stat(credFile)
	require.NoError(t, err, "credentials file should exist before deletion")

	config.DeleteS3Config(app, "my-minio")

	_, err = os.Stat(credFile)
	assert.True(t, os.IsNotExist(err), "credentials file should be removed after engine deletion")
}

func TestDeleteS3Config_NoCredFile(t *testing.T) {
	app, _ := newS3TestApp(t)

	app.StorageEngine = map[string]map[string]string{
		"no-creds": {storage.CloudType: storage.Custom},
	}

	// Should complete without error even when the credentials file is absent
	assert.NotPanics(t, func() {
		config.DeleteS3Config(app, "no-creds")
	})
	assert.NotContains(t, app.StorageEngine, "no-creds")
}

func TestDeleteS3Config_PersistsToDisk(t *testing.T) {
	app, _ := newS3TestApp(t)

	app.StorageEngine = map[string]map[string]string{
		"persist-me": {storage.CloudType: storage.Custom, storage.BucketName: "b"},
		"keep-me":    {storage.CloudType: storage.Custom, storage.BucketName: "c"},
	}
	// Save the added engines first so the file reflects them before deletion
	require.NoError(t, app.SaveToDisk(false))

	config.DeleteS3Config(app, "persist-me")

	// Reload from disk and confirm the deletion was persisted
	cfgFile := app.GetViperConfig().ConfigFileUsed()
	reloaded := config.NewConfig(cfgFile)
	assert.NotContains(t, reloaded.StorageEngine, "persist-me",
		"deleted engine should not appear in reloaded config")
	assert.Contains(t, reloaded.StorageEngine, "keep-me",
		"unrelated engine should still be present in reloaded config")
}
