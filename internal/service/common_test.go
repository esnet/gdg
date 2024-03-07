package service

import (
	"github.com/esnet/gdg/internal/config"
	"github.com/gosimple/slug"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
)

func TestRelativePathLogin(t *testing.T) {
	cwd, err := os.Getwd()
	assert.Nil(t, err)
	if strings.Contains(cwd, "service") {
		os.Chdir("../..")
	}
	os.Setenv("GDG_CONTEXTS__TESTING__URL", "http://localhost:3000/grafana/")
	config.InitConfig("config/testing.yml", "'")
	defer os.Unsetenv("GDG_CONTEXTS__TESTING__URL")

	svc := NewApiService("dummy")
	_, cfg := svc.(*DashNGoImpl).getNewClient()
	assert.Equal(t, cfg.Host, "localhost:3000")
	assert.Equal(t, cfg.BasePath, "/grafana/api")

}

// Validates the paths for the various entity types using the common
// code used to create folders and generate paths.
func TestSlug(t *testing.T) {
	result := GetSlug("thisTestMoo")
	assert.Equal(t, "thistestmoo", result)
	//This
	result = GetSlug("This Test Moo")
	assert.Equal(t, "this-test-moo", result)
}

func TestUserPath(t *testing.T) {
	err := os.Setenv("GDG_CONTEXT_NAME", "qa")
	assert.Nil(t, err)
	config.InitConfig("testing.yml", "'")
	path := BuildResourceFolder("", config.UserResource)
	assert.Equal(t, "test/data/users/", path)
}
func TestBuildDashboardPath(t *testing.T) {
	result := BuildResourceFolder("General", config.DashboardResource)
	assert.Equal(t, "test/data/org_your-org/dashboards/General", result)
}

func TestBuildFolderSourcePath(t *testing.T) {
	result := buildResourcePath(slug.Make("Some Folder"), config.FolderResource)
	assert.Equal(t, "test/data/org_your-org/folders/some-folder.json", result)

}

func TestBuildDataSourcePath(t *testing.T) {
	result := buildResourcePath(slug.Make("My DS"), config.ConnectionResource)
	assert.Equal(t, "test/data/org_your-org/connections/my-ds.json", result)
}
