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

// ---------------------------------------------------------------------------
// ListS3Configs — additional coverage
// ---------------------------------------------------------------------------

// TestListS3Configs_NilStorageEngine covers the nil-guard branch: when
// app.StorageEngine is explicitly nil the function must return a non-nil
// empty map rather than nil (callers iterate the result without nil checks).
func TestListS3Configs_NilStorageEngine(t *testing.T) {
	app, _ := newS3TestApp(t)
	app.StorageEngine = nil

	result := config.ListS3Configs(app)

	assert.NotNil(t, result, "returned map must never be nil")
	assert.Empty(t, result, "nil StorageEngine should yield an empty result")
}

// TestListS3Configs_SingleEngine_AllFields verifies that a single engine entry
// with all standard storage fields (endpoint, bucket, region, prefix) is
// returned intact.
func TestListS3Configs_SingleEngine_AllFields(t *testing.T) {
	app, _ := newS3TestApp(t)

	app.StorageEngine = map[string]map[string]string{
		"full-engine": {
			storage.CloudType:  storage.Custom,
			storage.Endpoint:   "http://minio.example.com:9000",
			storage.BucketName: "my-bucket",
			storage.Region:     "us-west-2",
			storage.Prefix:     "grafana/",
			storage.InitBucket: "true",
		},
	}

	result := config.ListS3Configs(app)

	require.Contains(t, result, "full-engine")
	eng := result["full-engine"]
	assert.Equal(t, storage.Custom, eng[storage.CloudType])
	assert.Equal(t, "http://minio.example.com:9000", eng[storage.Endpoint])
	assert.Equal(t, "my-bucket", eng[storage.BucketName])
	assert.Equal(t, "us-west-2", eng[storage.Region])
	assert.Equal(t, "grafana/", eng[storage.Prefix])
	assert.Equal(t, "true", eng[storage.InitBucket])
}

// TestListS3Configs_ReturnsSameReference confirms that ListS3Configs returns
// the live map (not a defensive copy).  A mutation made through the returned
// map must be visible via app.StorageEngine.
func TestListS3Configs_ReturnsSameReference(t *testing.T) {
	app, _ := newS3TestApp(t)

	app.StorageEngine = map[string]map[string]string{
		"original": {storage.BucketName: "bucket"},
	}

	result := config.ListS3Configs(app)
	result["new-engine"] = map[string]string{storage.BucketName: "new-bucket"}

	assert.Contains(t, app.StorageEngine, "new-engine",
		"mutation via returned map should be reflected in app.StorageEngine")
}

// ---------------------------------------------------------------------------
// DeleteS3Config — additional coverage
// ---------------------------------------------------------------------------

// TestDeleteS3Config_MultipleContextsBothAssigned ensures that when two
// contexts both reference the same storage engine, both have their Storage
// field cleared after deletion.
func TestDeleteS3Config_MultipleContextsBothAssigned(t *testing.T) {
	app, outputDir := newS3TestApp(t)

	app.StorageEngine = map[string]map[string]string{
		"shared-minio": {storage.CloudType: storage.Custom},
	}

	// Assign the engine to the default "testing" context.
	app.GetDefaultGrafanaConfig().Storage = "shared-minio"

	// Add a second context also referencing the same engine.
	secondCtx := config_domain.NewGrafanaConfig()
	secondCtx.Storage = "shared-minio"
	secondCtx.OutputPath = outputDir
	app.Contexts["staging"] = secondCtx

	config.DeleteS3Config(app, "shared-minio")

	assert.Empty(t, app.GetDefaultGrafanaConfig().Storage,
		"testing context Storage should be cleared")
	assert.Empty(t, app.Contexts["staging"].Storage,
		"staging context Storage should be cleared")
	assert.NotContains(t, app.StorageEngine, "shared-minio")
}

// TestDeleteS3Config_MultipleContexts_OtherStorageUntouched verifies that a
// context pointing to a *different* storage engine is not disturbed when an
// unrelated engine is deleted.
func TestDeleteS3Config_MultipleContexts_OtherStorageUntouched(t *testing.T) {
	app, outputDir := newS3TestApp(t)

	app.StorageEngine = map[string]map[string]string{
		"target":  {storage.CloudType: storage.Custom},
		"keeper":  {storage.CloudType: storage.Custom},
	}

	// Default context is assigned to the engine being deleted.
	app.GetDefaultGrafanaConfig().Storage = "target"

	// Second context uses a different engine — must be left alone.
	secondCtx := config_domain.NewGrafanaConfig()
	secondCtx.Storage = "keeper"
	secondCtx.OutputPath = outputDir
	app.Contexts["staging"] = secondCtx

	config.DeleteS3Config(app, "target")

	assert.Empty(t, app.GetDefaultGrafanaConfig().Storage,
		"testing context (assigned to deleted engine) should be cleared")
	assert.Equal(t, "keeper", app.Contexts["staging"].Storage,
		"staging context (assigned to a different engine) must not be modified")
}

// TestDeleteS3Config_MultipleEngines_OnlyTargetCredRemoved checks that when
// two engines each have a credentials file only the deleted engine's file is
// removed; the other file must remain on disk.
func TestDeleteS3Config_MultipleEngines_OnlyTargetCredRemoved(t *testing.T) {
	app, outputDir := newS3TestApp(t)

	app.StorageEngine = map[string]map[string]string{
		"delete-me": {storage.CloudType: storage.Custom},
		"keep-me":   {storage.CloudType: storage.Custom},
	}

	deletedCred := writeFakeCredFile(t, outputDir, "delete-me")
	keptCred := writeFakeCredFile(t, outputDir, "keep-me")

	config.DeleteS3Config(app, "delete-me")

	_, errDeleted := os.Stat(deletedCred)
	assert.True(t, os.IsNotExist(errDeleted),
		"deleted engine's credentials file should be removed")

	_, errKept := os.Stat(keptCred)
	assert.NoError(t, errKept,
		"surviving engine's credentials file must still exist")
}

// TestDeleteS3Config_SurvivingEngineIntact confirms that after one engine is
// deleted the remaining engine's configuration map is fully preserved.
func TestDeleteS3Config_SurvivingEngineIntact(t *testing.T) {
	app, _ := newS3TestApp(t)

	app.StorageEngine = map[string]map[string]string{
		"gone": {storage.CloudType: storage.Custom, storage.BucketName: "gone-bucket"},
		"stay": {
			storage.CloudType:  storage.Custom,
			storage.Endpoint:   "http://stay.example.com",
			storage.BucketName: "stay-bucket",
			storage.Region:     "eu-central-1",
			storage.Prefix:     "stay/",
		},
	}

	config.DeleteS3Config(app, "gone")

	require.Contains(t, app.StorageEngine, "stay",
		"surviving engine must still be present")
	stay := app.StorageEngine["stay"]
	assert.Equal(t, storage.Custom, stay[storage.CloudType])
	assert.Equal(t, "http://stay.example.com", stay[storage.Endpoint])
	assert.Equal(t, "stay-bucket", stay[storage.BucketName])
	assert.Equal(t, "eu-central-1", stay[storage.Region])
	assert.Equal(t, "stay/", stay[storage.Prefix])
}

// TestDeleteS3Config_CredFilePathUsesCloudAuthPrefix verifies that the
// credentials file is looked up (and removed) using the expected naming
// convention: <secure_dir>/<CloudAuthPrefix>_<label>.yaml.
func TestDeleteS3Config_CredFilePathUsesCloudAuthPrefix(t *testing.T) {
	app, outputDir := newS3TestApp(t)

	const label = "path-check"
	app.StorageEngine = map[string]map[string]string{
		label: {storage.CloudType: storage.Custom},
	}

	expectedPath := filepath.Join(
		outputDir, "secure",
		fmt.Sprintf("%s_%s.yaml", config_domain.CloudAuthPrefix, label),
	)
	// Write the file at exactly the path the implementation should target.
	require.NoError(t, os.MkdirAll(filepath.Dir(expectedPath), 0o750))
	require.NoError(t, os.WriteFile(expectedPath, []byte("access_id: x\n"), 0o600))

	config.DeleteS3Config(app, label)

	_, err := os.Stat(expectedPath)
	assert.True(t, os.IsNotExist(err),
		"credential file at the CloudAuthPrefix-based path should be removed")
}
