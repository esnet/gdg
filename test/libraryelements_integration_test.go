package test

import (
	"log/slog"
	"testing"

	"github.com/esnet/gdg/internal/service"
	"github.com/esnet/gdg/pkg/test_tooling"
	"github.com/gosimple/slug"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/stretchr/testify/assert"
)

func TestLibraryElementsCRUD(t *testing.T) {
	apiClient, _, _, cleanup := test_tooling.InitTestLegacy(t, nil, nil)
	defer cleanup()
	filtersEntity := service.NewDashboardFilter("", "", "")
	slog.Info("Exporting all Library Elements")
	apiClient.UploadLibraryElements(filtersEntity)
	slog.Info("Listing all library elements")
	boards := apiClient.ListLibraryElements(filtersEntity)
	slog.Info("Imported library elements", "count", len(boards))
	var generalBoard *models.LibraryElementDTO
	var otherBoard *models.LibraryElementDTO
	for ndx, board := range boards {
		slog.Info(board.Name)
		if slug.Make(board.Name) == "dashboard-makeover-extra-cleaning-duty-assignment-today" {
			generalBoard = boards[ndx]
		}
		if slug.Make(board.Name) == "extreme-dashboard-makeover-mac-oven" {
			otherBoard = boards[ndx]
		}
	}
	assert.NotNil(t, otherBoard)
	assert.NotNil(t, generalBoard)
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
	apiClient.UploadDashboards(service.NewDashboardFilter("", "", ""))
	// List connection
	connections := apiClient.ListLibraryElementsConnections(nil, "T47RSwQnz")
	assert.Equal(t, len(connections), 1)
	connection := connections[0]

	assert.Equal(t, connection.Meta.FolderTitle, "Other")
	assert.True(t, len(connection.Meta.FolderUID) > 0)
	assert.Equal(t, connection.Meta.Slug, "dashboard-makeover-challenge")
	assert.Equal(t, connection.Dashboard.(map[string]any)["uid"].(string), "F3eInwQ7z")
	assert.Equal(t, connection.Dashboard.(map[string]any)["title"].(string), "Dashboard Makeover Challenge")

	// Delete All Dashboards
	apiClient.DeleteAllDashboards(service.NewDashboardFilter("", "", ""))
	slog.Info("Deleting Library Elements")
	deleteList := apiClient.DeleteAllLibraryElements(filtersEntity)
	assert.Equal(t, len(deleteList), len(boards))
	slog.Info("List Dashboards again")
	boards = apiClient.ListLibraryElements(filtersEntity)
	assert.Equal(t, len(boards), 0)
}

func validateLibraryElement(t *testing.T, board *models.LibraryElementDTO, data map[string]any) {
	assert.Equal(t, board.Name, data["Name"].(string))
	assert.Equal(t, board.Type, data["Type"].(string))
	assert.Equal(t, board.UID, data["UID"].(string))
	assert.Equal(t, board.Kind, data["Kind"].(int64))
}
