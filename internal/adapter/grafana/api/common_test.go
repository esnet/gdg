package api

import (
	"context"
	"os"
	"testing"

	"github.com/esnet/gdg/internal/adapter/grafana/extended"
	"github.com/esnet/gdg/internal/adapter/plugins/secure/noop"
	"github.com/esnet/gdg/internal/adapter/storage"
	"github.com/esnet/gdg/internal/config"
	configDomain "github.com/esnet/gdg/internal/config/config_domain"
	"github.com/esnet/gdg/internal/domain"
	"github.com/esnet/gdg/pkg/test_tooling/common"
	"github.com/esnet/gdg/pkg/test_tooling/path"
	"github.com/gosimple/slug"
	"github.com/stretchr/testify/assert"
)

func fixEnvironment(t *testing.T) {
	assert.NoError(t, path.FixTestDir("api", "../../../.."))
	err := os.Setenv(common.ContextNameEnv, "qa")
	assert.Nil(t, err)
	config.NewConfig(common.DefaultTestConfig)
}

func TestRelativePathLogin(t *testing.T) {
	envKey := "GDG_CONTEXTS__QA__URL"
	assert.NoError(t, os.Setenv(envKey, "http://localhost:3000/grafana/"))
	fixEnvironment(t)
	defer os.Unsetenv(common.ContextNameEnv)
	cfg := config.NewConfig(common.DefaultTestConfig)
	defer func() {
		assert.NoError(t, os.Unsetenv(envKey))
	}()

	localEngine := storage.NewLocalStorage(context.Background())
	svc := NewDashNGo(cfg, noop.NoOpEncoder{}, localEngine, extended.NewExtendedApi(cfg))
	_, clientCfg := svc.(*DashNGoImpl).getNewClient()
	assert.Equal(t, clientCfg.Host, "localhost:3000")
	assert.Equal(t, clientCfg.BasePath, "/grafana/api")
}

// Validates the paths for the various entity types using the common
// code used to create folders and generate paths.
func TestSlug(t *testing.T) {
	fixEnvironment(t)
	defer os.Unsetenv(common.ContextNameEnv)
	result := GetSlug("thisTestMoo")
	assert.Equal(t, "thistestmoo", result)
	// This
	result = GetSlug("This Test Moo")
	assert.Equal(t, "this-test-moo", result)
}

func TestUserPath(t *testing.T) {
	fixEnvironment(t)
	defer os.Unsetenv(common.ContextNameEnv)
	cfg := &configDomain.GrafanaConfig{
		OutputPath: "test/data",
	}
	userPath := BuildResourceFolder(cfg, "", domain.UserResource, false, false)
	assert.Equal(t, "test/data/users/", userPath)
}

func TestBuildDashboardPath(t *testing.T) {
	fixEnvironment(t)
	defer os.Unsetenv(common.ContextNameEnv)
	cfg := &configDomain.GrafanaConfig{
		OutputPath:       "test/data",
		OrganizationName: "Your Org",
	}
	result := BuildResourceFolder(cfg, "General", domain.DashboardResource, false, false)
	assert.Equal(t, "test/data/org_your-org/dashboards/General", result)
}

func TestBuildFolderSourcePath(t *testing.T) {
	fixEnvironment(t)
	defer os.Unsetenv(common.ContextNameEnv)
	cfg := &configDomain.GrafanaConfig{
		OutputPath:       "test/data",
		OrganizationName: "Your Org",
	}
	result := BuildResourcePath(cfg, slug.Make("Some Folder"), domain.FolderResource, false, false)
	assert.Equal(t, "test/data/org_your-org/folders/some-folder.json", result)
}

func TestBuildDataSourcePath(t *testing.T) {
	fixEnvironment(t)
	defer os.Unsetenv(common.ContextNameEnv)
	cfg := &configDomain.GrafanaConfig{
		OutputPath:       "test/data",
		OrganizationName: "Your Org",
	}
	result := BuildResourcePath(cfg, slug.Make("My DS"), domain.ConnectionResource, false, false)
	assert.Equal(t, "test/data/org_your-org/connections/my-ds.json", result)
}
