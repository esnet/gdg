package service

import (
	"github.com/esnet/gdg/internal/config"
	"github.com/gosimple/slug"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

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
	assert.Equal(t, "qa/users/", path)
}
func TestBuildDashboardPath(t *testing.T) {
	result := BuildResourceFolder("General", config.DashboardResource)
	assert.Equal(t, "qa/org_1/dashboards/General", result)
}

func TestBuildFolderSourcePath(t *testing.T) {
	result := buildResourcePath(slug.Make("Some Folder"), config.FolderResource)
	assert.Equal(t, "qa/org_1/folders/some-folder.json", result)

}

func TestBuildDataSourcePath(t *testing.T) {
	result := buildResourcePath(slug.Make("My DS"), config.ConnectionResource)
	assert.Equal(t, "qa/org_1/connections/my-ds.json", result)
}

func TestBuildAlertNotificationPath(t *testing.T) {
	result := buildResourcePath("SomeNotification", config.AlertNotificationResource)
	assert.Equal(t, "qa/org_1/alertnotifications/SomeNotification.json", result)
}
