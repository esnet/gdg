package test

import (
	"fmt"
	"log/slog"
	"maps"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/esnet/gdg/internal/service"

	"github.com/esnet/gdg/internal/config"
	"github.com/gosimple/slug"

	"github.com/testcontainers/testcontainers-go"

	"github.com/esnet/gdg/internal/types"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/samber/lo"

	"github.com/esnet/gdg/pkg/test_tooling"

	"github.com/stretchr/testify/assert"
)

func TestFolderCRUD(t *testing.T) {
	config.InitGdgConfig("testing")
	apiClient, _, cleanup := test_tooling.InitTest(t, service.DefaultConfigProvider, nil)
	defer cleanup()
	slog.Info("Exporting all folders")
	apiClient.UploadFolders(nil)
	slog.Info("Listing all Folders")
	folders := apiClient.ListFolders(nil)
	assert.Equal(t, len(folders), 2)
	firstDsItem := folders[0]
	assert.Equal(t, firstDsItem.Title, "Ignored")
	secondDsItem := folders[1]
	assert.Equal(t, secondDsItem.Title, "Other")
	// Import Folders
	slog.Info("Importing folders")
	list := apiClient.DownloadFolders(nil)
	assert.Equal(t, len(list), len(folders))
	slog.Info("Deleting Folders")
	deleteList := apiClient.DeleteAllFolders(nil)
	assert.Equal(t, len(deleteList), len(folders))
	slog.Info("List Folders again")
	folders = apiClient.ListFolders(nil)
	assert.Equal(t, len(folders), 0)
}

func TestFolderCRUDInvalidChar(t *testing.T) {
	if os.Getenv(test_tooling.EnableTokenTestsEnv) == "1" {
		t.Skip("Skipping Token configuration, BasicAuth required to setup org structure")
	}
	containerObj, cleanup := test_tooling.InitOrganizations(t)
	defer cleanup()

	//
	config.InitGdgConfig("testing")
	cfg := config.Config()
	cfg.GetDefaultGrafanaConfig().OrganizationName = "Bad Folder"
	cfg.GetDefaultGrafanaConfig().GetDashboardSettings().IgnoreBadFolders = true

	cfgProvider := func() *config.Configuration {
		return cfg
	}

	orgClient := test_tooling.CreateSimpleClientWithConfig(t, cfgProvider, containerObj)
	orgClient.UploadFolders(nil)
	slog.Info("Listing all Folders")
	folders := orgClient.ListFolders(nil)
	assert.Equal(t, len(folders), 2)
	cfg.GetDefaultGrafanaConfig().GetDashboardSettings().IgnoreBadFolders = false
}

// TODO: write a full CRUD validation of folder permissions
func TestFolderPermissions(t *testing.T) {
	config.InitGdgConfig("testing")
	apiClient, _, cleanup := test_tooling.InitTest(t, service.DefaultConfigProvider, nil)
	defer cleanup()
	slog.Info("Exporting all folders")
	apiClient.UploadFolders(nil)
	slog.Info("Listing all Folders")
	folders := apiClient.ListFolders(nil)
	assert.Equal(t, len(folders), 2)
	result := apiClient.ListFolderPermissions(nil)
	assert.True(t, len(result) > 0)
	for key, val := range result {
		assert.NotNil(t, key)
		if os.Getenv(test_tooling.EnableTokenTestsEnv) == "1" {
			assert.Equal(t, 2, len(val))
		} else {
			assert.Equal(t, 3, len(val))
		}
	}

	data := apiClient.DownloadFolderPermissions(nil)
	assert.Equal(t, len(data), 2)
	permissionKeys := lo.Map(slices.Collect(maps.Keys(result)), func(item *types.FolderDetails, index int) string {
		return fmt.Sprintf("test/data/org_main-org/folders-permissions/%s.json", slug.Make(item.NestedPath))
	})
	for _, item := range data {
		assert.True(t, slices.Contains(permissionKeys, item))
	}
}

// TODO: write a full CRUD validation of folder permissions
func TestFolderNestedPermissions(t *testing.T) {
	if os.Getenv(test_tooling.EnableTokenTestsEnv) == "1" {
		t.Skip("skipping token based tests")
	}
	containerObj, cleanup := test_tooling.InitOrganizations(t)
	dockerContainer := containerObj.(*testcontainers.DockerContainer)
	if getGrafanaVersion(dockerContainer.Image) < minimumNestedFoldersVersion {
		t.Log("Nested folders not supported prior to v11.0, skipping test")
		t.Skip()
	}
	assert.NoError(t, os.Setenv(test_tooling.OrgNameOverride, "testing"))
	assert.NoError(t, os.Setenv(test_tooling.EnableNestedBehavior, "true"))
	defer func() {
		os.Unsetenv(test_tooling.OrgNameOverride)
		os.Unsetenv(test_tooling.EnableNestedBehavior)
		cleanup()
	}()

	apiClient, _ := test_tooling.CreateSimpleClient(t, nil, containerObj)
	slog.Info("Exporting all folders")
	apiClient.UploadFolders(nil)
	slog.Info("Listing all Folders")
	folders := apiClient.ListFolders(nil)
	assert.Equal(t, len(folders), 4)
	result := apiClient.ListFolderPermissions(nil)
	assert.True(t, len(result) > 0)
	for key, val := range result {
		assert.NotNil(t, key)
		if strings.Contains(key.NestedPath, "/") {
			assert.Equal(t, 1, len(val))
		} else {
			assert.Equal(t, 3, len(val))
		}
	}

	data := apiClient.DownloadFolderPermissions(nil)
	assert.Equal(t, len(data), 4)
	permissionKeys := lo.Map(slices.Collect(maps.Keys(result)), func(item *types.FolderDetails, index int) string {
		return fmt.Sprintf("test/data/org_testing/folders-permissions/%s.json", slug.Make(item.NestedPath))
	})
	for _, item := range data {
		assert.True(t, slices.Contains(permissionKeys, item))
	}
}

func TestFolderNestedCRUD(t *testing.T) {
	if os.Getenv(test_tooling.EnableTokenTestsEnv) == "1" {
		t.Skip("skipping token based tests")
	}

	containerObj, cleanup := test_tooling.InitOrganizations(t)

	dockerContainer := containerObj.(*testcontainers.DockerContainer)

	if getGrafanaVersion(dockerContainer.Image) < minimumNestedFoldersVersion {
		t.Log("Nested folders not supported prior to v11.0, skipping test")
		t.Skip()
	}

	assert.NoError(t, os.Setenv(test_tooling.OrgNameOverride, "testing"))
	assert.NoError(t, os.Setenv(test_tooling.EnableNestedBehavior, "true"))
	defer func() {
		os.Unsetenv(test_tooling.OrgNameOverride)
		os.Unsetenv(test_tooling.EnableNestedBehavior)
		cleanup()
	}()

	apiClient, _ := test_tooling.CreateSimpleClient(t, nil, containerObj)

	slog.Info("Exporting all folders")
	apiClient.UploadFolders(nil)
	slog.Info("Listing all Folders")
	folders := apiClient.ListFolders(nil)
	assert.Equal(t, len(folders), 4)
	firstDsItem := lo.FirstOrEmpty(lo.Filter(folders, func(item *types.FolderDetails, index int) bool {
		return item.NestedPath == "Others/dummy"
	}))
	assert.Equal(t, firstDsItem.Title, "dummy")
	assert.Equal(t, firstDsItem.FolderTitle, "Others")
	assert.Equal(t, firstDsItem.Type, models.HitType("dash-folder"))
	secondDsItem := lo.FirstOrEmpty(lo.Filter(folders, func(item *types.FolderDetails, index int) bool {
		return item.NestedPath == "Others"
	}))
	assert.Equal(t, secondDsItem.Title, "Others")
	assert.Equal(t, secondDsItem.FolderTitle, "")
	// Import Folders
	slog.Info("Importing folders")
	list := apiClient.DownloadFolders(nil)
	assert.Equal(t, len(list), len(folders))
	slog.Info("Deleting Folders")
	deleteList := apiClient.DeleteAllFolders(nil)
	assert.Equal(t, len(deleteList), len(folders))
	slog.Info("List Folders again")
	folders = apiClient.ListFolders(nil)
	assert.Equal(t, len(folders), 0)
}
