package test

import (
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service"
	"github.com/esnet/gdg/internal/service/filters"
	"github.com/esnet/grafana-swagger-api-golang/goclient/models"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slices"
	"log/slog"
)

//TODO: with full CRUD.
// - Add single dashboard test -d <>
// - Add Folder dashboard test -f <>

func TestDashboardCRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	apiClient, _, cleanup := initTest(t, nil)
	defer cleanup()
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

func TestDashboardTagsFilter(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	apiClient, _, cleanup := initTest(t, nil)
	defer cleanup()
	emptyFilter := filters.NewBaseFilter()

	filtersEntity := service.NewDashboardFilter("", "", "")
	filtersEntity.AddFilter(filters.TagsFilter, strings.Join([]string{"flow", "netsage"}, ","))

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
	apiClient, _, cleanup := initTest(t, nil)
	defer cleanup()
	emptyFilter := filters.NewBaseFilter()

	filtersEntity := service.NewDashboardFilter("", "", "")
	filtersEntity.AddFilter(filters.TagsFilter, strings.Join([]string{"flow", "netsage"}, ","))

	// Enable Wildcard
	testingContext := config.Config().GetAppConfig().GetContexts()["testing"]
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

	assert.Equal(t, len(boards), len(boards_filtered))

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
	assert.Equal(t, len(board.Tags), 2)
	assert.True(t, slices.Contains(board.Tags, "netsage"))
	assert.True(t, slices.Contains(board.Tags, "flow"))
}
