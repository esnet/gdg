package test

import (
	"context"
	"fmt"
	"log/slog"
	"maps"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/esnet/gdg/internal/service/domain"
	"github.com/esnet/gdg/pkg/test_tooling/containers"

	"github.com/esnet/gdg/pkg/test_tooling/common"

	"github.com/esnet/gdg/internal/service"

	"github.com/esnet/gdg/internal/config"
	"github.com/gosimple/slug"

	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/samber/lo"

	"github.com/esnet/gdg/pkg/test_tooling"

	"github.com/stretchr/testify/assert"
)

func TestFolderCRUD(t *testing.T) {
	cfg := config.InitGdgConfig(common.DefaultTestConfig)
	cfg.GetDefaultGrafanaConfig().GetDashboardSettings().IgnoreFilters = false
	var r *test_tooling.InitContainerResult
	err := Retry(context.Background(), DefaultRetryAttempts, func() error {
		r = test_tooling.InitTest(t, cfg, nil)
		return r.Err
	})
	assert.NotNil(t, r)
	assert.NoError(t, err)
	defer func() {
		err := r.CleanUp()
		if err != nil {
			slog.Warn("Unable to clean up after test", "test", t.Name())
		}
	}()
	apiClient := r.ApiClient
	slog.Info("Exporting all folders")
	apiClient.UploadFolders(nil)
	slog.Info("Listing all Folders")
	folders := apiClient.ListFolders(nil)
	assert.Equal(t, len(folders), 4)
	firstDsItem := lo.FindOrElse(folders, nil, func(item *domain.NestedHit) bool {
		return item.Title == "Ignored"
	})
	assert.Equal(t, firstDsItem.Title, "Ignored")
	assert.Equal(t, firstDsItem.FolderTitle, "")
	secondDsItem := lo.FindOrElse(folders, nil, func(item *domain.NestedHit) bool {
		return item.Title == "Others"
	})
	assert.Equal(t, secondDsItem.Title, "Others")
	assert.Equal(t, secondDsItem.FolderTitle, "linux/gnu")
	// Import Folders
	slog.Info("importing folders")
	list := apiClient.DownloadFolders(nil)
	assert.Equal(t, len(list), len(folders))
	slog.Info("Deleting Folders")
	deleteList := apiClient.DeleteAllFolders(nil)
	assert.Equal(t, len(deleteList), len(folders))
	slog.Info("List Folders again")
	folders = apiClient.ListFolders(nil)
	assert.Equal(t, len(folders), 0)
}

func TestFolderSanityCheck(t *testing.T) {
	cfg := config.InitGdgConfig(common.DefaultTestConfig)
	var r *test_tooling.InitContainerResult
	err := Retry(context.Background(), DefaultRetryAttempts, func() error {
		r = test_tooling.InitTest(t, cfg, nil)
		return r.Err
	})
	assert.NotNil(t, r)
	assert.NoError(t, err)
	defer func() {
		err := r.CleanUp()
		if err != nil {
			slog.Warn("Unable to clean up after test", "test", t.Name())
		}
	}()
	apiClient := r.ApiClient
	newFolders, err := apiClient.(*service.DashNGoImpl).TestCreatedFolders("linux%2Fgnu/Others")
	assert.NoError(t, err)
	assert.Equal(t, len(newFolders), 2)
	folders := apiClient.ListFolders(nil)
	assert.Equal(t, len(folders), 2)
	newFolders, err = apiClient.(*service.DashNGoImpl).TestCreatedFolders("linux%2Fgnu/Others/n%2B_%3D23r")
	assert.NoError(t, err)
	assert.Equal(t, len(newFolders), 1)
	folders = apiClient.ListFolders(nil)
	assert.Equal(t, len(folders), 3)
}

// TODO: write a full CRUD validation of folder permissions
func TestFolderPermissions(t *testing.T) {
	cfg := config.InitGdgConfig(common.DefaultTestConfig)
	var r *test_tooling.InitContainerResult
	err := Retry(context.Background(), DefaultRetryAttempts, func() error {
		r = test_tooling.InitTest(t, cfg, nil)
		return r.Err
	})
	assert.NotNil(t, r)
	assert.NoError(t, err)
	defer func() {
		err := r.CleanUp()
		if err != nil {
			slog.Warn("Unable to clean up after test", "test", t.Name())
		}
	}()
	apiClient := r.ApiClient
	slog.Info("Exporting all folders")
	apiClient.UploadFolders(nil)
	slog.Info("Listing all Folders")
	folders := apiClient.ListFolders(nil)
	assert.Equal(t, len(folders), 4)
	result := apiClient.ListFolderPermissions(nil)
	assert.True(t, len(result) > 0)
	// this behavior is inconsistent across versions, disabled for now
	/*
		for key, val := range result {
			assert.NotNil(t, key)
			if os.Getenv(test_tooling.EnableTokenTestsEnv) == "1" {
				assert.Equal(t, 2, len(val))
			} else {
				assert.Equal(t, 3, len(val))
			}
		}

	*/

	data := apiClient.DownloadFolderPermissions(nil)
	assert.Equal(t, len(data), 4)
	permissionKeys := lo.Map(slices.Collect(maps.Keys(result)), func(item *domain.NestedHit, index int) string {
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
	cfg := config.InitGdgConfig(common.DefaultTestConfig)
	containerObj, cleanup := test_tooling.InitOrganizations(t, cfg)

	assert.NoError(t, os.Setenv(test_tooling.OrgNameOverride, "testing"))
	defer func() {
		os.Unsetenv(test_tooling.OrgNameOverride)
		cleanup()
	}()

	apiClient, _ := test_tooling.CreateSimpleClient(t, cfg, nil, containerObj)
	slog.Info("Exporting all folders")
	apiClient.UploadFolders(nil)
	slog.Info("Listing all Folders")
	folders := apiClient.ListFolders(nil)
	assert.Equal(t, len(folders), 4)
	result := apiClient.ListFolderPermissions(nil)
	assert.True(t, len(result) > 0)
	grafanaVersion := getGrafanaVersion(":" + containers.GetGrafanaVersion())

	for key, val := range result {
		assert.NotNil(t, key)
		if strings.Contains(key.NestedPath, "/") {
			switch grafanaVersion {
			case 11:
				assert.Equal(t, 1, len(val))
			case 12:
				assert.Equal(t, 0, len(val))
			}
		} else {
			assert.Equal(t, 3, len(val))
		}
	}

	data := apiClient.DownloadFolderPermissions(nil)
	assert.Equal(t, len(data), 4)
	permissionKeys := lo.Map(slices.Collect(maps.Keys(result)), func(item *domain.NestedHit, index int) string {
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
	cfg := config.InitGdgConfig(common.DefaultTestConfig)

	containerObj, cleanup := test_tooling.InitOrganizations(t, cfg)

	assert.NoError(t, os.Setenv(test_tooling.OrgNameOverride, "testing"))
	defer func() {
		os.Unsetenv(test_tooling.OrgNameOverride)
		cleanup()
	}()

	apiClient, _ := test_tooling.CreateSimpleClient(t, cfg, nil, containerObj)

	slog.Info("Exporting all folders")
	apiClient.UploadFolders(nil)
	slog.Info("Listing all Folders")
	folders := apiClient.ListFolders(nil)
	assert.Equal(t, len(folders), 4)
	firstDsItem := lo.FirstOrEmpty(lo.Filter(folders, func(item *domain.NestedHit, index int) bool {
		return item.NestedPath == "Others/dummy"
	}))
	assert.Equal(t, firstDsItem.Title, "dummy")
	assert.Equal(t, firstDsItem.FolderTitle, "Others")
	assert.Equal(t, firstDsItem.Type, models.HitType("dash-folder"))
	secondDsItem := lo.FirstOrEmpty(lo.Filter(folders, func(item *domain.NestedHit, index int) bool {
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
