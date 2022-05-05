package api

import (
	"github.com/esnet/gdg/config"
	"github.com/gosimple/slug"
	"github.com/stretchr/testify/assert"
	"testing"
)

//Validates the paths for the various entity types using the common
// code used to create folders and generate paths.
func TestSlug(t *testing.T) {
	result := GetSlug("thisTestMoo")
	assert.Equal(t, "thistestmoo", result)
	//This
	result = GetSlug("This Test Moo")
	assert.Equal(t, "this-test-moo", result)
}

func TestUserPath(t *testing.T) {
	config.InitConfig("testing.yml", "'")
	path := buildResourceFolder("", config.UserResource)
	assert.Equal(t, "qa/users/", path)
}
func TestBuildDashboardPath(t *testing.T) {
	result := buildResourceFolder("General", config.DashboardResource)
	assert.Equal(t, "qa/dashboards/General", result)
}

func TestBuildFolderSourcePath(t *testing.T) {
	result := buildResourcePath(slug.Make("Some Folder"), config.FolderResource)
	assert.Equal(t, "qa/folders/some-folder.json", result)

}

func TestBuildDataSourcePath(t *testing.T) {
	result := buildResourcePath(slug.Make("My DS"), config.DataSourceResource)
	assert.Equal(t, "qa/datasources/my-ds.json", result)
}

func TestBuildAlertNotificationPath(t *testing.T) {
	result := buildResourcePath("SomeNotification", config.AlertNotificationResource)
	assert.Equal(t, "qa/alertnotifications/SomeNotification.json", result)
}
