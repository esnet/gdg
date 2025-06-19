package test

import (
	"context"
	"log/slog"
	"testing"

	customModels "github.com/esnet/gdg/internal/service/domain"

	"github.com/esnet/gdg/pkg/test_tooling/common"

	"github.com/esnet/gdg/internal/config"

	"github.com/esnet/gdg/internal/service"
	"github.com/esnet/gdg/pkg/test_tooling"
	"github.com/gosimple/slug"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/stretchr/testify/assert"
)

func TestLibraryElementsCRUD(t *testing.T) {
	config.InitGdgConfig(common.DefaultTestConfig)
	testCase := []struct {
		name          string
		ignore        bool
		expectedCount int
		disabled      bool
	}{
		{
			ignore:        true,
			name:          "ignore Enabled Test",
			expectedCount: 10,
		},
		{
			ignore:        false,
			name:          "ignore Disabled Test",
			expectedCount: 9,
		},
	}

	for _, tc := range testCase {
		if tc.disabled {
			t.Log("Skipping test test", tc.name)
			continue
		}
		t.Log("Running test", tc.name)
		var apiClient service.GrafanaService
		var cleanup func() error
		wrapTest(func() {
			config.InitGdgConfig(common.DefaultTestConfig)
		})
		cfgProvider := func() *config.Configuration {
			// Needed to unset filters
			cfg := config.Config()
			cfg.GetDefaultGrafanaConfig().GetDashboardSettings().IgnoreFilters = tc.ignore
			return cfg
		}
		var r *test_tooling.InitContainerResult
		err := Retry(context.Background(), DefaultRetryAttempts, func() error {
			r = test_tooling.InitTest(t, cfgProvider, nil)
			return r.Err
		})
		assert.NotNil(t, r)
		assert.NoError(t, err)
		apiClient = r.ApiClient
		cleanup = r.CleanUp
		filtersEntity := service.NewLibraryElementFilter()
		dashFilter := service.NewDashboardFilter("", "", "")
		slog.Info("Exporting all Library Elements")
		uploadCount := apiClient.UploadLibraryElements(filtersEntity)
		assert.Equal(t, len(uploadCount), tc.expectedCount)
		slog.Info("Listing all library elements")
		boards := apiClient.ListLibraryElements(filtersEntity)
		slog.Info("Imported library elements", "count", len(boards))
		assert.Equal(t, len(boards), tc.expectedCount)
		var generalBoard *customModels.WithNested[models.LibraryElementDTO]
		var otherBoard *customModels.WithNested[models.LibraryElementDTO]
		var ignoreBoard *customModels.WithNested[models.LibraryElementDTO]
		for ndx, board := range boards {
			slog.Info(board.Entity.Name)
			if slug.Make(board.Entity.Name) == "dashboard-makeover-extra-cleaning-duty-assignment-today" {
				generalBoard = boards[ndx]
			}
			if slug.Make(board.Entity.Name) == "extreme-dashboard-makeover-mac-oven" {
				otherBoard = boards[ndx]
			}
			if board.Entity.Name == "Dashboard Makeover - Ignored" {
				ignoreBoard = boards[ndx]
			}
		}
		assert.NotNil(t, otherBoard)
		assert.NotNil(t, generalBoard)
		if tc.ignore {
			assert.NotNil(t, ignoreBoard)
			validateLibraryElement(t, ignoreBoard, map[string]any{
				"Name": "Dashboard Makeover - Ignored",
				"Type": "table", "UID": "1DTh3UQ7k", "Kind": int64(1),
			})

		} else {
			assert.Nil(t, ignoreBoard)
		}
		validateLibraryElement(t, generalBoard, map[string]any{
			"Name": "Dashboard Makeover - Extra Cleaning Duty Assignment Today",
			"Type": "table", "UID": "T47RSwQnz", "Kind": int64(1),
		})
		validateLibraryElement(t, otherBoard, map[string]any{
			"Name": "Extreme Dashboard Makeover - Mac Oven",
			"Type": "stat", "UID": "VvzpJ5X7z", "Kind": int64(1),
		})

		// Import Library Elements
		slog.Info("Importing Library Elements")
		list := apiClient.DownloadLibraryElements(filtersEntity)
		assert.Equal(t, len(list), len(boards))
		// Export all Dashboards
		_, dashErr := apiClient.UploadDashboards(dashFilter)
		assert.NoError(t, dashErr)
		// List connection
		connections := apiClient.ListLibraryElementsConnections(filtersEntity, "T47RSwQnz")
		assert.Equal(t, len(connections), 1)
		connection := connections[0]

		assert.Equal(t, connection.Meta.FolderTitle, "n+_=23r")
		assert.True(t, len(connection.Meta.FolderUID) > 0)
		assert.Equal(t, connection.Meta.Slug, "dashboard-makeover-challenge")
		assert.Equal(t, connection.Dashboard.(map[string]any)["uid"].(string), "F3eInwQ7z")
		assert.Equal(t, connection.Dashboard.(map[string]any)["title"].(string), "Dashboard Makeover Challenge")

		// Delete All Dashboards
		apiClient.DeleteAllDashboards(dashFilter)
		slog.Info("Deleting Library Elements")
		deleteList := apiClient.DeleteAllLibraryElements(filtersEntity)
		assert.Equal(t, len(deleteList), len(boards))
		slog.Info("List Dashboards again")
		boards = apiClient.ListLibraryElements(filtersEntity)
		assert.Equal(t, len(boards), 0)

		assert.NoError(t, cleanup(), "Failed to cleanup container for test"+tc.name)

	}
}

func validateLibraryElement(t *testing.T, board *customModels.WithNested[models.LibraryElementDTO], data map[string]any) {
	assert.Equal(t, board.Entity.Name, data["Name"].(string))
	assert.Equal(t, board.Entity.Type, data["Type"].(string))
	assert.Equal(t, board.Entity.UID, data["UID"].(string))
	assert.Equal(t, board.Entity.Kind, data["Kind"].(int64))
}
