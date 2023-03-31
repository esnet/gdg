package integration_tests

import (
	"github.com/esnet/gdg/api"
	"github.com/esnet/gdg/api/filters"
	"github.com/esnet/grafana-swagger-api-golang/goclient/models"
	"strings"
	"testing"

	"github.com/esnet/gdg/config"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slices"
)

//TODO: with full CRUD.
// - Add single dashboard test -d <>
// - Add Folder dashboard test -f <>

func TestDashboardCRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	apiClient, _ := initTest(t)
	filtersEntity := api.NewDashboardFilter("", "", "")
	log.Info("Exporting all dashboards")
	apiClient.ExportDashboards(filtersEntity)
	log.Info("Listing all dashboards")
	boards := apiClient.ListDashboards(filtersEntity)
	log.Infof("Imported %d dashboards", len(boards))
	ignoredSkipped := true
	var generalBoard *models.Hit
	var otherBoard *models.Hit
	for ndx, board := range boards {
		log.Infof(board.Slug)
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
	log.Info("Importing Dashboards")
	list := apiClient.ImportDashboards(filtersEntity)
	assert.Equal(t, len(list), len(boards))
	log.Info("Deleting Dashboards")
	deleteList := apiClient.DeleteAllDashboards(filtersEntity)
	assert.Equal(t, len(deleteList), len(boards))
	log.Info("List Dashboards again")
	boards = apiClient.ListDashboards(filtersEntity)
	assert.Equal(t, len(boards), 0)
}

func TestDashboardTagsFilter(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	apiClient, _ := initTest(t)
	emptyFilter := filters.NewBaseFilter()

	filtersEntity := api.NewDashboardFilter("", "", "")
	filtersEntity.AddFilter(filters.TagsFilter, strings.Join([]string{"flow", "netsage"}, ","))

	log.Info("Exporting all dashboards")
	apiClient.ExportDashboards(emptyFilter)

	log.Info("Listing all dashboards")
	boards := apiClient.ListDashboards(filtersEntity)

	log.Infof("Imported %d dashboards", len(boards))
	for _, board := range boards {
		validateTags(t, board)
	}

	//Import Dashboards
	log.Info("Importing Dashboards")
	list := apiClient.ImportDashboards(filtersEntity)
	assert.Equal(t, len(list), len(boards))

	log.Info("Deleting Dashboards")
	deleteList := apiClient.DeleteAllDashboards(filtersEntity)
	assert.Equal(t, len(deleteList), len(boards))

	log.Info("List Dashboards again")
	boards = apiClient.ListDashboards(filtersEntity)
	assert.Equal(t, len(boards), 0)
}

func TestWildcardFilter(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Setup Filters
	apiClient, _ := initTest(t)
	emptyFilter := filters.NewBaseFilter()

	filtersEntity := api.NewDashboardFilter("", "", "")
	filtersEntity.AddFilter(filters.TagsFilter, strings.Join([]string{"flow", "netsage"}, ","))

	// Enable Wildcard
	testingContext := config.Config().Contexts()["testing"]
	testingContext.GetFilterOverrides().IgnoreDashboardFilters = true
	assert.True(t, testingContext.GetFilterOverrides().IgnoreDashboardFilters)

	// Testing Exporting with Wildcard
	apiClient.ExportDashboards(emptyFilter)
	boards := apiClient.ListDashboards(emptyFilter)

	apiClient.ExportDashboards(filtersEntity)
	boards_filtered := apiClient.ListDashboards(emptyFilter)

	assert.Equal(t, len(boards), len(boards_filtered))

	// Testing Listing with Wildcard
	log.Info("Listing all dashboards without filter")
	boards = apiClient.ListDashboards(emptyFilter)

	log.Info("Listing all dashboards ignoring filter")
	boards_filtered = apiClient.ListDashboards(filtersEntity)

	assert.Equal(t, len(boards), len(boards_filtered))

	log.Info("Importing Dashboards")
	list := apiClient.ImportDashboards(emptyFilter)
	assert.Equal(t, len(list), len(boards))

	log.Info("Deleting Dashboards")
	deleteList := apiClient.DeleteAllDashboards(emptyFilter)
	assert.Equal(t, len(deleteList), len(boards))

	log.Info("List Dashboards again")
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
