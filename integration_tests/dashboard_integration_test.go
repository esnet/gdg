package integration_tests

import (
	"strings"
	"testing"

	"github.com/esnet/gdg/api"
	"github.com/grafana-tools/sdk"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestDashboardCRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	apiClient, _ := initTest(t)
	filters := api.NewDashboardFilter()
	log.Info("Exporting all dashboards")
	apiClient.ExportDashboards(filters)
	log.Info("Listing all dashboards")
	boards := apiClient.ListDashboards(filters)
	log.Infof("Imported %d dashboards", len(boards))
	ignoredSkipped := true
	var generalBoard sdk.FoundBoard
	var otherBoard sdk.FoundBoard
	for ndx, board := range boards {
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
	list := apiClient.ImportDashboards(filters)
	assert.Equal(t, len(list), len(boards))
	log.Info("Deleting Dashboards")
	deleteList := apiClient.DeleteAllDashboards(filters)
	assert.Equal(t, len(deleteList), len(boards))
	log.Info("List Dashboards again")
	boards = apiClient.ListDashboards(filters)
	assert.Equal(t, len(boards), 0)
}

func TestDashboardTagsFilter(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	apiClient, _ := initTest(t)
	emptyFilter := api.NewDashboardFilter()

	filters := api.NewDashboardFilter()
	filters.AddFilter(api.TagsFilter, strings.Join(["flow", "netsage"], ","))

	log.Info("Exporting all dashboards")
	apiClient.ExportDashboards(emptyFilter)

	log.Info("Listing all dashboards")
	boards := apiClient.ListDashboards(filters)

	log.Infof("Imported %d dashboards", len(boards))
	for _, board := range boards {
		validateTags(t, board)
	}

	//Import Dashboards
	log.Info("Importing Dashboards")
	list := apiClient.ImportDashboards(filters)
	assert.Equal(t, len(list), len(boards))

	log.Info("Deleting Dashboards")
	deleteList := apiClient.DeleteAllDashboards(filters)
	assert.Equal(t, len(deleteList), len(boards))

	log.Info("List Dashboards again")
	boards = apiClient.ListDashboards(filters)
	assert.Equal(t, len(boards), 0)
}

func validateOtherBoard(t *testing.T, board sdk.FoundBoard) {
	assert.True(t, board.UID != "")
	assert.Equal(t, board.Title, "Flow Information")
	assert.Equal(t, board.URI, "db/flow-information")
	assert.True(t, strings.Contains(board.URL, board.UID))
	assert.True(t, strings.Contains(board.URL, board.Slug))
	assert.Equal(t, board.Type, "dash-db")
	assert.Equal(t, board.FolderTitle, "Other")
}

func validateGeneralBoard(t *testing.T, board sdk.FoundBoard) {
	assert.True(t, board.UID != "")
	assert.Equal(t, board.Title, "Individual Flows")
	assert.Equal(t, board.URI, "db/individual-flows")
	assert.True(t, strings.Contains(board.URL, board.UID))
	assert.True(t, strings.Contains(board.URL, board.Slug))
	assert.Equal(t, len(board.Tags), 1)
	assert.Equal(t, board.Tags[0], "netsage")
	assert.Equal(t, board.Type, "dash-db")
	assert.Equal(t, board.FolderID, 0)
	assert.Equal(t, board.FolderTitle, "General")

}

func validateTags(t *testing.T, board sdk.FoundBoard) {
	assert.True(t, board.UID != "")
	assert.Equal(t, len(board.Tags), 2)
	assert.Equal(t, board.Tags[0], "netsage")
	assert.Equal(t, board.Tags[1], "flow")
}
