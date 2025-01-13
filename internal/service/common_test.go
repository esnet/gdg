package service

import (
	"context"
	"os"
	"testing"

	"github.com/esnet/gdg/internal/storage"

	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/pkg/test_tooling/common"
	"github.com/esnet/gdg/pkg/test_tooling/path"
	"github.com/gosimple/slug"
	"github.com/stretchr/testify/assert"
)

func fixEnvironment(t *testing.T) {
	assert.NoError(t, path.FixTestDir("service", "../.."))
	err := os.Setenv("GDG_CONTEXT_NAME", "qa")
	assert.Nil(t, err)
	config.InitGdgConfig(common.DefaultTestConfig)
}

func TestRelativePathLogin(t *testing.T) {
	envKey := "GDG_CONTEXTS__QA__URL"
	assert.NoError(t, os.Setenv(envKey, "http://localhost:3000/grafana/"))
	fixEnvironment(t)
	config.InitGdgConfig(common.DefaultTestConfig)
	defer func() {
		assert.NoError(t, os.Unsetenv(envKey))
		assert.NoError(t, os.Unsetenv(path.TestEnvKey))
	}()

	localEngine := storage.NewLocalStorage(context.Background())
	svc := NewTestApiService(localEngine, nil)
	_, cfg := svc.(*DashNGoImpl).getNewClient()
	assert.Equal(t, cfg.Host, "localhost:3000")
	assert.Equal(t, cfg.BasePath, "/grafana/api")
}

// Validates the paths for the various entity types using the common
// code used to create folders and generate paths.
func TestSlug(t *testing.T) {
	fixEnvironment(t)
	result := GetSlug("thisTestMoo")
	assert.Equal(t, "thistestmoo", result)
	// This
	result = GetSlug("This Test Moo")
	assert.Equal(t, "this-test-moo", result)
}

func TestUserPath(t *testing.T) {
	fixEnvironment(t)
	userPath := BuildResourceFolder("", config.UserResource, false, false)
	assert.Equal(t, "test/data/users/", userPath)
}

func TestBuildDashboardPath(t *testing.T) {
	fixEnvironment(t)
	result := BuildResourceFolder("General", config.DashboardResource, false, false)
	assert.Equal(t, "test/data/org_your-org/dashboards/General", result)
}

func TestBuildFolderSourcePath(t *testing.T) {
	fixEnvironment(t)
	result := buildResourcePath(slug.Make("Some Folder"), config.FolderResource, false, false)
	assert.Equal(t, "test/data/org_your-org/folders/some-folder.json", result)
}

func TestBuildDataSourcePath(t *testing.T) {
	fixEnvironment(t)

	result := buildResourcePath(slug.Make("My DS"), config.ConnectionResource, false, false)
	assert.Equal(t, "test/data/org_your-org/connections/my-ds.json", result)
}
