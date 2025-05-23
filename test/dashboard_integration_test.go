package test

import (
	"encoding/json"
	"log/slog"
	"os"
	"slices"
	"strings"
	"testing"

	customModels "github.com/esnet/gdg/internal/types"

	"github.com/esnet/gdg/internal/tools/ptr"

	"github.com/esnet/gdg/internal/types"
	"github.com/esnet/gdg/pkg/test_tooling/containers"
	"github.com/samber/lo"

	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service"
	"github.com/esnet/gdg/internal/service/filters"
	"github.com/esnet/gdg/pkg/test_tooling"
	"github.com/grafana/grafana-openapi-client-go/models"

	"github.com/stretchr/testify/assert"
)

const (
	DashboardCount = 16
)

func TestDashboardNestedFolderCRUD(t *testing.T) {
	if os.Getenv(test_tooling.EnableTokenTestsEnv) == test_tooling.FeatureEnabled {
		t.Skip("skipping token based tests")
	}
	containerObj, cleanup := test_tooling.InitOrganizations(t)

	assert.NoError(t, os.Setenv(test_tooling.OrgNameOverride, "testing"))
	assert.NoError(t, os.Setenv(test_tooling.EnableNestedBehavior, "true"))
	assert.NoError(t, os.Setenv(test_tooling.IgnoreDashFilters, "true"))

	defer func() {
		os.Unsetenv(test_tooling.OrgNameOverride)
		os.Unsetenv(test_tooling.EnableNestedBehavior)
		os.Unsetenv(test_tooling.IgnoreDashFilters)
		cleanup()
	}()

	apiClient, _ := test_tooling.CreateSimpleClient(t, nil, containerObj)

	filtersEntity := service.NewDashboardFilter("", "", "")
	slog.Info("Exporting all dashboards")
	assert.NoError(t, apiClient.UploadDashboards(filtersEntity))
	slog.Info("Listing all dashboards")
	boards := apiClient.ListDashboards(filtersEntity)
	slog.Info("Imported dashboards", "count", len(boards))
	var generalBoard *customModels.NestedHit
	var nestedFolder *customModels.NestedHit
	for ndx, board := range boards {

		if board.Slug == "rabbitmq-overview" {
			generalBoard = boards[ndx]
		}
		if board.Slug == "node-exporter-full" {
			nestedFolder = boards[ndx]
		}
	}
	assert.NotNil(t, generalBoard)
	assert.NotNil(t, nestedFolder)
	assert.Equal(t, nestedFolder.NestedPath, "Others/dummy")

	// Import Dashboards
	numBoards := 3
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

func TestDashboardCRUD(t *testing.T) {
	config.InitGdgConfig("testing")
	apiClient, _, cleanup := test_tooling.InitTest(t, service.DefaultConfigProvider, nil)
	defer func() {
		err := cleanup()
		if err != nil {
			slog.Warn("Unable to clean up after dashboard tests")
		}
	}()
	filtersEntity := service.NewDashboardFilter("", "", "")
	slog.Info("Exporting all dashboards")
	assert.NoError(t, apiClient.UploadDashboards(filtersEntity))
	slog.Info("Listing all dashboards")
	boards := apiClient.ListDashboards(filtersEntity)
	slog.Info("Imported dashboards", "count", len(boards))
	ignoredSkipped := true
	var generalBoard *customModels.NestedHit
	var otherBoard *customModels.NestedHit
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
	// Validate filters

	filterFolder := service.NewDashboardFilter("Other", "", "")
	boards = apiClient.ListDashboards(filterFolder)
	assert.Equal(t, 8, len(boards))
	dashboardFilter := service.NewDashboardFilter("", "flow-information", "")
	boards = apiClient.ListDashboards(dashboardFilter)
	assert.Equal(t, 1, len(boards))

	// Import Dashboards
	slog.Info("Importing Dashboards")
	list := apiClient.DownloadDashboards(filtersEntity)
	assert.Equal(t, len(list), DashboardCount)
	slog.Info("Deleting Dashboards")
	deleteList := apiClient.DeleteAllDashboards(filtersEntity)
	assert.Equal(t, len(deleteList), DashboardCount)
	slog.Info("List Dashboards again")
	boards = apiClient.ListDashboards(filtersEntity)
	assert.Equal(t, len(boards), 0)
}

// If a duplicate file with the same UID exists, the upload should fail.  Having a cleanup flag turned on, should
// fix that issue.
func TestDashboardCleanUpCrud(t *testing.T) {
	config.InitGdgConfig("testing")
	cfgProvider := func() *config.Configuration {
		cfg := config.Config()
		cfg.GetDefaultGrafanaConfig().GetDashboardSettings().IgnoreFilters = true
		return cfg
	}
	apiClient, containerObj, cleanup := test_tooling.InitTest(t, cfgProvider, nil)
	defer func() {
		err := cleanup()
		if err != nil {
			slog.Warn("Unable to clean up after dashboard tests")
		}
	}()
	filtersEntity := service.NewDashboardFilter("", "", "")
	slog.Info("Exporting all dashboards")
	assert.NoError(t, apiClient.UploadDashboards(filtersEntity))
	slog.Info("Listing all dashboards")
	boards := apiClient.ListDashboards(filtersEntity)
	assert.Equal(t, len(boards), DashboardCount+1) // Includes the Ignored folder
	// Create another copy of the dashboard json
	// copy file
	data, err := os.ReadFile("test/data/org_main-org/dashboards/General/bandwidth-dashboard.json")
	assert.NoError(t, err)
	err = os.WriteFile("test/data/org_main-org/dashboards/General/bandwidth-dashboard-copy.json", data, 0o644)
	assert.NoError(t, err)
	defer os.Remove("test/data/org_main-org/dashboards/General/bandwidth-dashboard-copy.json")
	cfgProvider = func() *config.Configuration {
		cfg := config.Config()
		cfg.GetDefaultGrafanaConfig().GetDashboardSettings().IgnoreFilters = true
		globals := cfg.GetGDGConfig().Global
		globals.ClearOutput = true
		return cfg
	}
	apiClient = test_tooling.CreateSimpleClientWithConfig(t, cfgProvider, containerObj)
	apiClient.DownloadDashboards(filtersEntity)
	assert.Nil(t, err)
	boards = apiClient.ListDashboards(filtersEntity)
	assert.Equal(t, len(boards), DashboardCount+1) // includes the ignored folder
}

func TestDashboardCRUDTags(t *testing.T) {
	config.InitGdgConfig("testing")
	apiClient, container, cleanup := test_tooling.InitTest(t, service.DefaultConfigProvider, nil)
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
	assert.NoError(t, apiClient.UploadDashboards(filtersEntity))
	slog.Info("Listing all dashboards")
	boards := apiClient.ListDashboards(filtersEntity)
	slog.Info("Removing all dashboards")
	assert.Equal(t, 13, len(boards))
	deleteList := apiClient.DeleteAllDashboards(filtersEntity)
	assert.Equal(t, 13, len(deleteList))
	// Multiple Tags behavior
	slog.Info("Uploading all dashboards, filtered by tags")
	data, err = json.Marshal([]string{"flow"})
	assert.NoError(t, err)
	filtersEntity = service.NewDashboardFilter("", "", string(data))
	assert.NoError(t, apiClient.UploadDashboards(filtersEntity))
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
	assert.NoError(t, apiClient.UploadDashboards(filterNone))
	// Listing with no filter
	boards = apiClient.ListDashboards(filterNone)
	assert.Equal(t, DashboardCount, len(boards))

	data, err = json.Marshal([]string{"flow"})
	assert.NoError(t, err)
	filtersEntity = service.NewDashboardFilter("", "", string(data))

	slog.Info("Listing dashboards by tag")
	boards = apiClient.ListDashboards(filtersEntity)
	assert.Equal(t, 8, len(deleteList))
	// Listing with
	data, err = json.Marshal([]string{"flow", "netsage"})
	assert.NoError(t, err)
	filtersEntity = service.NewDashboardFilter("", "", string(data))

	boards = apiClient.ListDashboards(filtersEntity)
	assert.Equal(t, 13, len(boards))
	deleteList = apiClient.DeleteAllDashboards(filtersEntity)
	assert.Equal(t, 13, len(deleteList))
}

func TestDashboardTagsFilter(t *testing.T) {
	config.InitGdgConfig("testing")
	apiClient, _, cleanup := test_tooling.InitTest(t, service.DefaultConfigProvider, nil)
	defer cleanup()
	emptyFilter := filters.NewBaseFilter()

	data, err := json.Marshal([]string{"flow", "netsage"})
	assert.NoError(t, err)
	filtersEntity := service.NewDashboardFilter("", "", string(data))

	slog.Info("Exporting all dashboards")
	assert.NoError(t, apiClient.UploadDashboards(emptyFilter))

	slog.Info("Listing all dashboards")
	boards := apiClient.ListDashboards(filtersEntity)

	slog.Info("Imported %d dashboards", "count", len(boards))
	for _, board := range boards {
		validateTags(t, board)
	}

	// Import Dashboards
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

func TestDashboardPermissionsCrud(t *testing.T) {
	if os.Getenv(test_tooling.EnableTokenTestsEnv) == "1" {
		t.Skip("Skipping Token configuration, Team and User CRUD requires Basic SecureData")
	}
	config.InitGdgConfig("testing")
	props := containers.DefaultGrafanaEnv()
	err := containers.SetupGrafanaLicense(&props)
	if err != nil {
		slog.Error("no valid grafana license found, skipping enterprise tests")
		t.Skip()
	}
	apiClient, _, cleanup := test_tooling.InitTest(t, service.DefaultConfigProvider, props)
	defer cleanup()
	// Upload all dashboards
	assert.NoError(t, apiClient.UploadDashboards(nil))
	// Upload all users
	newUsers := apiClient.UploadUsers(service.NewUserFilter(""))
	assert.Equal(t, len(newUsers), 2)
	// Upload all teams
	filter := service.NewTeamFilter("")
	teams := apiClient.UploadTeams(filter)
	assert.Equal(t, len(teams), 2)
	// Get current Permissions
	currentPerms, err := apiClient.ListDashboardPermissions(nil)
	assert.Equal(t, len(currentPerms), DashboardCount)
	entry := ptr.Of(lo.FirstOrEmpty(lo.Filter(currentPerms, func(item types.DashboardAndPermissions, index int) bool {
		return item.Dashboard.Title == "Bandwidth Dashboard"
	})))
	assert.NotNil(t, entry)
	assert.Equal(t, len(entry.Permissions), 3)

	assert.NoError(t, apiClient.ClearDashboardPermissions(nil))
	currentPerms, err = apiClient.ListDashboardPermissions(nil)
	assert.NoError(t, err)
	assert.Equal(t, len(currentPerms), DashboardCount)
	assert.Equal(t, len(currentPerms[0].Permissions), 0)
	addPerms, err := apiClient.UploadDashboardPermissions(nil)
	assert.NoError(t, err)
	assert.Equal(t, len(addPerms), DashboardCount)
	currentPerms, err = apiClient.ListDashboardPermissions(nil)
	entry = nil
	entry = ptr.Of(lo.FirstOrEmpty(lo.Filter(currentPerms, func(item types.DashboardAndPermissions, index int) bool {
		return item.Dashboard.Title == "Bandwidth Dashboard"
	})))
	assert.NotNil(t, entry)
	assert.Equal(t, 5, len(entry.Permissions))
	var bobPerm *models.DashboardACLInfoDTO
	var teamMusic *models.DashboardACLInfoDTO
	for ndx, entryPerm := range entry.Permissions {
		if entryPerm.Team == "musicians" {
			teamMusic = entry.Permissions[ndx]
		}
		if entryPerm.UserLogin == "bob" {
			bobPerm = entry.Permissions[ndx]
		}
	}
	assert.NotNil(t, bobPerm)
	assert.NotNil(t, teamMusic)
	// validate bob
	assert.Equal(t, bobPerm.PermissionName, "Edit")
	assert.Equal(t, bobPerm.UserEmail, "bob@aol.com")
	assert.Equal(t, bobPerm.UserID, int64(2))
	assert.Equal(t, bobPerm.Permission, models.PermissionType(2))
	// validate team permission
	assert.Equal(t, teamMusic.PermissionName, "Admin")
	assert.Equal(t, teamMusic.TeamID, int64(2))
	assert.Equal(t, teamMusic.Permission, models.PermissionType(4))
}

func TestWildcardFilter(t *testing.T) {
	// Setup Filters
	config.InitGdgConfig("testing")
	apiClient, _, cleanup := test_tooling.InitTest(t, service.DefaultConfigProvider, nil)
	defer cleanup()
	emptyFilter := service.NewDashboardFilter("", "", "")

	data, err := json.Marshal([]string{"flow", "netsage"})
	assert.NoError(t, err)
	filtersEntity := service.NewDashboardFilter("", "", string(data))

	// Enable Wildcard
	testingContext := config.Config().GetGDGConfig().GetContexts()["testing"]
	testingContext.GetDashboardSettings().IgnoreFilters = true
	assert.True(t, testingContext.GetDashboardSettings().IgnoreFilters)

	// Testing Exporting with Wildcard
	assert.NoError(t, apiClient.UploadDashboards(emptyFilter))
	boards := apiClient.ListDashboards(emptyFilter)

	assert.NoError(t, apiClient.UploadDashboards(filtersEntity))
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

func validateOtherBoard(t *testing.T, board *customModels.NestedHit) {
	assert.True(t, board.UID != "")
	assert.Equal(t, board.Title, "Flow Information")
	assert.Equal(t, board.URI, "db/flow-information")
	assert.True(t, strings.Contains(board.URL, board.UID))
	assert.True(t, strings.Contains(board.URL, board.Slug))
	assert.Equal(t, board.Type, models.HitType("dash-db"))
	assert.Equal(t, board.FolderTitle, "Other")
}

func validateGeneralBoard(t *testing.T, board *customModels.NestedHit) {
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

func validateTags(t *testing.T, board *customModels.NestedHit) {
	assert.True(t, board.UID != "")
	assert.True(t, len(board.Tags) > 0)
	allTags := []string{"netsage", "flow"}
	for _, tag := range board.Tags {
		assert.True(t, slices.Contains(allTags, tag))
	}
}
