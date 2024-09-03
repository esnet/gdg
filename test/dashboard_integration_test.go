package test

import (
	"encoding/json"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service"
	"github.com/esnet/gdg/internal/service/filters"
	"github.com/esnet/gdg/pkg/test_tooling"
	"github.com/grafana/grafana-openapi-client-go/models"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"log/slog"
	"slices"
)

func TestDashboardCRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	apiClient, _, _, cleanup := test_tooling.InitTest(t, nil, false)
	defer func() {
		err := cleanup()
		if err != nil {
			slog.Warn("Unable to clean up after dashboard tests")
		}
	}()
	filtersEntity := service.NewDashboardFilter("", "", "")
	slog.Info("Exporting all dashboards")
	apiClient.UploadDashboards(filtersEntity)
	slog.Info("Listing all dashboards")
	boards := apiClient.ListDashboards(filtersEntity)
	slog.Info("Imported dashboards", "count", len(boards))
	ignoredSkipped := true
	var generalBoard *models.Hit
	var otherBoard *models.Hit
	for ndx, board := range boards {
		slog.Info(board.Slug)
		if board.Slug == "latency-patterns" {
			ignoredSkipped = false
		}
		if board.Slug == "individual-flows" {
			generalBoard = boards[ndx]
		}
		if board.Slug == "flow-information" {
			otherBoard = boards[ndx]
		}
	}
	assert.NotNil(t, otherBoard)
	assert.NotNil(t, generalBoard)
	assert.True(t, ignoredSkipped)
	validateGeneralBoard(t, generalBoard)
	validateOtherBoard(t, otherBoard)
	//Validate filters

	filterFolder := service.NewDashboardFilter("Other", "", "")
	boards = apiClient.ListDashboards(filterFolder)
	assert.Equal(t, 8, len(boards))
	dashboardFilter := service.NewDashboardFilter("", "flow-information", "")
	boards = apiClient.ListDashboards(dashboardFilter)
	assert.Equal(t, 1, len(boards))

	//Import Dashboards
	numBoards := 16
	slog.Info("Importing Dashboards")
	list := apiClient.DownloadDashboards(filtersEntity)
	assert.Equal(t, len(list), numBoards)
	slog.Info("Deleting Dashboards")
	deleteList := apiClient.DeleteAllDashboards(filtersEntity)
	assert.Equal(t, len(deleteList), numBoards)
	slog.Info("List Dashboards again")
	boards = apiClient.ListDashboards(filtersEntity)
	assert.Equal(t, len(boards), 0)
}

func TestDashboardCRUDTags(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	apiClient, _, container, cleanup := test_tooling.InitTest(t, nil, false)
	defer func() {
		err := cleanup()
		if err != nil {
			slog.Warn("Unable to clean up after dashboard tests")
		}
	}()

	data, err := json.Marshal([]string{"netsage"})
	assert.NoError(t, err)
	filtersEntity := service.NewDashboardFilter("", "", string(data))

	slog.Info("Uploading all dashboards, filtered by tags")
	apiClient.UploadDashboards(filtersEntity)
	slog.Info("Listing all dashboards")
	boards := apiClient.ListDashboards(filtersEntity)
	slog.Info("Removing all dashboards")
	assert.Equal(t, 13, len(boards))
	deleteList := apiClient.DeleteAllDashboards(filtersEntity)
	assert.Equal(t, 13, len(deleteList))
	//Multiple Tags behavior
	slog.Info("Uploading all dashboards, filtered by tags")
	data, err = json.Marshal([]string{"flow"})
	assert.NoError(t, err)
	filtersEntity = service.NewDashboardFilter("", "", string(data))
	apiClient.UploadDashboards(filtersEntity)
	slog.Info("Listing all dashboards")
	boards = apiClient.ListDashboards(filtersEntity)
	assert.Equal(t, 8, len(boards))
	slog.Info("Removing all dashboards")
	deleteList = apiClient.DeleteAllDashboards(filtersEntity)
	assert.Equal(t, 8, len(deleteList))
	//
	os.Setenv("GDG_CONTEXTS__TESTING__IGNORE_FILTERS", "true")
	defer os.Unsetenv("")
	apiClient, _ = test_tooling.CreateSimpleClient(t, nil, container)
	filterNone := service.NewDashboardFilter("", "", "")
	apiClient.UploadDashboards(filterNone)
	//Listing with no filter
	boards = apiClient.ListDashboards(filterNone)
	assert.Equal(t, 16, len(boards))

	data, err = json.Marshal([]string{"flow"})
	assert.NoError(t, err)
	filtersEntity = service.NewDashboardFilter("", "", string(data))

	slog.Info("Listing dashboards by tag")
	boards = apiClient.ListDashboards(filtersEntity)
	assert.Equal(t, 8, len(deleteList))
	//Listing with
	data, err = json.Marshal([]string{"flow", "netsage"})
	assert.NoError(t, err)
	filtersEntity = service.NewDashboardFilter("", "", string(data))

	boards = apiClient.ListDashboards(filtersEntity)
	assert.Equal(t, 13, len(boards))
	deleteList = apiClient.DeleteAllDashboards(filtersEntity)
	assert.Equal(t, 13, len(deleteList))
}

func TestDashboardTagsFilter(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	apiClient, _, _, cleanup := test_tooling.InitTest(t, nil, false)
	defer cleanup()
	emptyFilter := filters.NewBaseFilter()

	data, err := json.Marshal([]string{"flow", "netsage"})
	assert.NoError(t, err)
	filtersEntity := service.NewDashboardFilter("", "", string(data))

	slog.Info("Exporting all dashboards")
	apiClient.UploadDashboards(emptyFilter)

	slog.Info("Listing all dashboards")
	boards := apiClient.ListDashboards(filtersEntity)

	slog.Info("Imported %d dashboards", "count", len(boards))
	for _, board := range boards {
		validateTags(t, board)
	}

	//Import Dashboards
	slog.Info("Importing Dashboards")
	list := apiClient.DownloadDashboards(filtersEntity)
	assert.Equal(t, len(list), len(boards))

	slog.Info("Deleting Dashboards")
	deleteList := apiClient.DeleteAllDashboards(filtersEntity)
	assert.Equal(t, len(deleteList), len(boards))

	slog.Info("List Dashboards again")
	boards = apiClient.ListDashboards(filtersEntity)
	assert.Equal(t, len(boards), 0)
}

func TestWildcardFilter(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Setup Filters
	apiClient, _, _, cleanup := test_tooling.InitTest(t, nil, false)
	defer cleanup()
	emptyFilter := service.NewDashboardFilter("", "", "")

	data, err := json.Marshal([]string{"flow", "netsage"})
	assert.NoError(t, err)
	filtersEntity := service.NewDashboardFilter("", "", string(data))

	// Enable Wildcard
	testingContext := config.Config().GetGDGConfig().GetContexts()["testing"]
	testingContext.GetFilterOverrides().IgnoreDashboardFilters = true
	assert.True(t, testingContext.GetFilterOverrides().IgnoreDashboardFilters)

	// Testing Exporting with Wildcard
	apiClient.UploadDashboards(emptyFilter)
	boards := apiClient.ListDashboards(emptyFilter)

	apiClient.UploadDashboards(filtersEntity)
	boards_filtered := apiClient.ListDashboards(emptyFilter)

	assert.Equal(t, len(boards), len(boards_filtered))

	// Testing Listing with Wildcard
	slog.Info("Listing all dashboards without filter")
	boards = apiClient.ListDashboards(emptyFilter)

	slog.Info("Listing all dashboards ignoring filter")
	boards_filtered = apiClient.ListDashboards(filtersEntity)

	assert.Equal(t, 14, len(boards_filtered))

	slog.Info("Importing Dashboards")
	list := apiClient.DownloadDashboards(emptyFilter)
	assert.Equal(t, len(list), len(boards))

	slog.Info("Deleting Dashboards")
	deleteList := apiClient.DeleteAllDashboards(emptyFilter)
	assert.Equal(t, len(deleteList), len(boards))

	slog.Info("List Dashboards again")
	boards = apiClient.ListDashboards(filtersEntity)
	assert.Equal(t, len(boards), 0)
}

func validateOtherBoard(t *testing.T, board *models.Hit) {
	assert.True(t, board.UID != "")
	assert.Equal(t, board.Title, "Flow Information")
	assert.Equal(t, board.URI, "db/flow-information")
	assert.True(t, strings.Contains(board.URL, board.UID))
	assert.True(t, strings.Contains(board.URL, board.Slug))
	assert.Equal(t, board.Type, models.HitType("dash-db"))
	assert.Equal(t, board.FolderTitle, "Other")
}

func validateGeneralBoard(t *testing.T, board *models.Hit) {
	assert.True(t, board.UID != "")
	assert.Equal(t, board.Title, "Individual Flows")
	assert.Equal(t, board.URI, "db/individual-flows")
	assert.True(t, strings.Contains(board.URL, board.UID))
	assert.True(t, strings.Contains(board.URL, board.Slug))
	assert.Equal(t, len(board.Tags), 1)
	assert.Equal(t, board.Tags[0], "netsage")
	assert.Equal(t, board.Type, models.HitType("dash-db"))
	assert.Equal(t, board.FolderID, int64(0))
	assert.Equal(t, board.FolderTitle, "General")

}

func validateTags(t *testing.T, board *models.Hit) {
	assert.True(t, board.UID != "")
	assert.True(t, len(board.Tags) > 0)
	allTags := []string{"netsage", "flow"}
	for _, tag := range board.Tags {
		assert.True(t, slices.Contains(allTags, tag))

	}
}
