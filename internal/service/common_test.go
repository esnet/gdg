package service

import (
	"os"
	"testing"

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
	defer assert.NoError(t, os.Unsetenv(envKey))

	svc := NewApiService("dummy")
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
	path := BuildResourceFolder("", config.UserResource)
	assert.Equal(t, "test/data/users/", path)
}

func TestBuildDashboardPath(t *testing.T) {
	fixEnvironment(t)
	result := BuildResourceFolder("General", config.DashboardResource)
	assert.Equal(t, "test/data/org_your-org/dashboards/General", result)
}

func TestBuildFolderSourcePath(t *testing.T) {
	fixEnvironment(t)
	result := buildResourcePath(slug.Make("Some Folder"), config.FolderResource)
	assert.Equal(t, "test/data/org_your-org/folders/some-folder.json", result)
}

func TestBuildDataSourcePath(t *testing.T) {
	fixEnvironment(t)

	result := buildResourcePath(slug.Make("My DS"), config.ConnectionResource)
	assert.Equal(t, "test/data/org_your-org/connections/my-ds.json", result)
}
