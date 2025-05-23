package test

import (
	"encoding/json"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/esnet/gdg/internal/service/filters/v1"

	customModels "github.com/esnet/gdg/internal/types"

	"github.com/esnet/gdg/internal/tools/ptr"

	"github.com/esnet/gdg/internal/types"
	"github.com/esnet/gdg/pkg/test_tooling/containers"
	"github.com/samber/lo"

	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service"
	"github.com/esnet/gdg/pkg/test_tooling"
	"github.com/grafana/grafana-openapi-client-go/models"

	"github.com/stretchr/testify/assert"
)

const (
	DashboardCount       = 16
	IgnoreDashboardCount = DashboardCount + 1
)

func TestDashboardCRUDIgnoreFilters(t *testing.T) {
	config.InitGdgConfig("testing")
	cfgProvider := func() *config.Configuration {
		cfg := config.Config()
		cfg.GetDefaultGrafanaConfig().GetDashboardSettings().IgnoreFilters = true
		return cfg
	}
	apiClient, _, cleanup := test_tooling.InitTest(t, cfgProvider, nil)
	defer func() {
		err := cleanup()
		if err != nil {
			slog.Warn("Unable to clean up after dashboard tests")
		}
	}()
	filtersEntity := service.NewDashboardFilter("", "", "")
	slog.Info("Exporting all dashboards")
	uploadedFiles, err := apiClient.UploadDashboards(filtersEntity)
	assert.NoError(t, err)

	slog.Info("Listing all dashboards")
	boards := apiClient.ListDashboardsLegacy(filtersEntity)
	slog.Info("Imported dashboards", "count", len(boards), "uploadedFiles", len(uploadedFiles))
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
	assert.False(t, ignoredSkipped)
	validateGeneralBoard(t, generalBoard)
	validateOtherBoard(t, otherBoard)
	// Validate filters

	filterFolder := service.NewDashboardFilter("linux%2Fgnu", "", "")
	boards = apiClient.ListDashboardsLegacy(filterFolder)
	assert.Equal(t, 8, len(boards))
	// With Regex filters
	filterFolder = service.NewDashboardFilter("linux%2Fgnu$", "", "")
	boards = apiClient.ListDashboardsLegacy(filterFolder)
	assert.Equal(t, 4, len(boards))
	//
	dashboardFilter := service.NewDashboardFilter("", "flow-information", "")
	boards = apiClient.ListDashboardsLegacy(dashboardFilter)
	assert.Equal(t, 1, len(boards))

	// Import Dashboards
	slog.Info("Importing Dashboards")
	list := apiClient.DownloadDashboards(filtersEntity)
	assert.Equal(t, len(list), IgnoreDashboardCount)
	slog.Info("Deleting Dashboards")
	deleteList := apiClient.DeleteAllDashboards(filtersEntity)
	assert.Equal(t, len(deleteList), IgnoreDashboardCount)
	slog.Info("List Dashboards again")
	boards = apiClient.ListDashboardsLegacy(filtersEntity)
	assert.Equal(t, len(boards), 0)
}

// If a duplicate file with the same UID exists, the upload should fail.  Having a cleanup flag turned on, should
// fix that issue.
func TestDashboardCleanUpCrud(t *testing.T) {
	config.InitGdgConfig("testing")

	apiClient, containerObj, cleanup := test_tooling.InitTest(t, service.DefaultConfigProvider, nil)
	defer func() {
		err := cleanup()
		if err != nil {
			slog.Warn("Unable to clean up after dashboard tests")
		}
	}()
	filtersEntity := service.NewDashboardFilter("", "", "")
	slog.Info("Exporting all dashboards")
	uploadedFiles, err := apiClient.UploadDashboards(filtersEntity)
	assert.NoError(t, err)
	assert.Equal(t, len(uploadedFiles), DashboardCount)
	slog.Info("Listing all dashboards")
	boards := apiClient.ListDashboardsLegacy(filtersEntity)
	assert.Equal(t, len(boards), DashboardCount) // Includes the Ignored folder
	// Create another copy of the dashboard json
	// copy file
	data, err := os.ReadFile("test/data/org_main-org/dashboards/General/bandwidth-dashboard.json")
	assert.NoError(t, err)
	err = os.WriteFile("test/data/org_main-org/dashboards/General/bandwidth-dashboard-copy.json", data, 0o644)
	assert.NoError(t, err)
	defer os.Remove("test/data/org_main-org/dashboards/General/bandwidth-dashboard-copy.json")
	cfgProvider := func() *config.Configuration {
		cfg := config.Config()
		cfg.GetDefaultGrafanaConfig().GetDashboardSettings().IgnoreFilters = true
		globals := cfg.GetGDGConfig().Global
		globals.ClearOutput = true
		return cfg
	}
	apiClient = test_tooling.CreateSimpleClientWithConfig(t, cfgProvider, containerObj)
	apiClient.DownloadDashboards(filtersEntity)
	assert.Nil(t, err)
	boards = apiClient.ListDashboardsLegacy(filtersEntity)
	assert.Equal(t, len(boards), IgnoreDashboardCount) // includes the ignored folder
}

// Download relies on Listing behavior so we only need to check that the dashboard listing works properly
func TestDashListFilters(t *testing.T) {
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
	encodeTags := func(tags ...string) string {
		raw, err := json.Marshal(tags)
		assert.NoError(t, err, "unable to encode tags")
		return string(raw)
	}
	uploadedFiles, err := apiClient.UploadDashboards(nil)
	assert.NoError(t, err)
	assert.Equal(t, len(uploadedFiles), IgnoreDashboardCount)
	// folder test
	filtersEntity := service.NewDashboardFilter("linux%2Fgnu/Ot*", "", "")
	boards := apiClient.ListDashboardsLegacy(filtersEntity)
	assert.Equal(t, len(boards), 4)
	//
	filtersEntity = service.NewDashboardFilter("", "", encodeTags("flow"))
	boards = apiClient.ListDashboardsLegacy(filtersEntity)
	assert.Equal(t, len(boards), 8)
	// Dash filter
	filtersEntity = service.NewDashboardFilter("individual-flows-per-country", "", "")
	boards = apiClient.ListDashboardsLegacy(filtersEntity)
	assert.Equal(t, len(boards), 1)
	// Filtering without ignore flags
	cfgProvider = func() *config.Configuration {
		cfg := config.Config()
		cfg.GetDefaultGrafanaConfig().GetDashboardSettings().IgnoreFilters = false
		return cfg
	}
	apiClient = test_tooling.CreateSimpleClientWithConfig(t, cfgProvider, containerObj)
	// no additional filter
	boards = apiClient.ListDashboardsLegacy(filtersEntity)
	assert.Equal(t, len(boards), DashboardCount)
	// folder test
	filtersEntity = service.NewDashboardFilter("linux%2Fgnu/Ot*", "", "")
	boards = apiClient.ListDashboardsLegacy(filtersEntity)
	assert.Equal(t, len(boards), 4)
	//
	filtersEntity = service.NewDashboardFilter("", "", encodeTags("flow"))
	boards = apiClient.ListDashboardsLegacy(filtersEntity)
	assert.Equal(t, len(boards), 8)
	// Dash filter
	filtersEntity = service.NewDashboardFilter("individual-flows-per-country", "", "")
	boards = apiClient.ListDashboardsLegacy(filtersEntity)
	assert.Equal(t, len(boards), 1)
}

func TestUploadDashboardsBehavior(t *testing.T) {
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
	cleanupDash := func(count int) {
		items := apiClient.DeleteAllDashboards(service.NewDashboardFilter("", "", ""))
		assert.Equal(t, len(items), count)
	}

	encodeTags := func(tags ...string) string {
		raw, err := json.Marshal(tags)
		assert.NoError(t, err, "unable to encode tags")
		return string(raw)
	}

	uploadedFiles, err := apiClient.UploadDashboards(nil)
	assert.NoError(t, err)
	assert.Equal(t, len(uploadedFiles), IgnoreDashboardCount)
	cleanupDash(IgnoreDashboardCount)
	// folder test
	// filtersEntity := service.NewDashboardFilter("linux%2Fgnu/Ot*", "", "")
	// uploadedFiles, err = apiClient.UploadDashboards(filtersEntity)
	// assert.NoError(t, err)
	// assert.Equal(t, len(uploadedFiles), 4)
	// cleanupDash(len(uploadedFiles))
	// Tags filter
	filtersEntity := service.NewDashboardFilter("", "", encodeTags("flow"))
	uploadedFiles, err = apiClient.UploadDashboards(filtersEntity)
	assert.NoError(t, err)
	assert.Equal(t, len(uploadedFiles), 8)
	cleanupDash(len(uploadedFiles))
	// Dash filter
	//filtersEntity = service.NewDashboardFilter("individual-flows-per-country", "", "")
	//uploadedFiles, err = apiClient.UploadDashboards(filtersEntity)
	//assert.NoError(t, err)
	//assert.Equal(t, len(uploadedFiles), 1)
	//cleanupDash(len(uploadedFiles))
	//
	cfgProvider = func() *config.Configuration {
		cfg := config.Config()
		cfg.GetDefaultGrafanaConfig().GetDashboardSettings().IgnoreFilters = false
		return cfg
	}
	apiClient = test_tooling.CreateSimpleClientWithConfig(t, cfgProvider, containerObj)
	uploadedFiles, err = apiClient.UploadDashboards(nil)
	assert.NoError(t, err)
	assert.Equal(t, len(uploadedFiles), DashboardCount)
	cleanupDash(DashboardCount)
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
	uploadedFiles, err := apiClient.UploadDashboards(filtersEntity)
	assert.NoError(t, err)
	assert.Equal(t, len(uploadedFiles), 13)
	slog.Info("Listing all dashboards")
	boards := apiClient.ListDashboardsLegacy(filtersEntity)
	slog.Info("Removing all dashboards")
	assert.Equal(t, 13, len(boards))
	deleteList := apiClient.DeleteAllDashboards(filtersEntity)
	assert.Equal(t, 13, len(deleteList))
	// Multiple Tags behavior
	slog.Info("Uploading all dashboards, filtered by tags")
	data, err = json.Marshal([]string{"flow"})
	assert.NoError(t, err)
	filtersEntity = service.NewDashboardFilter("", "", string(data))
	uploadedFiles, err = apiClient.UploadDashboards(filtersEntity)
	assert.NoError(t, err)
	assert.Equal(t, len(uploadedFiles), 8)
	slog.Info("Listing all dashboards")
	boards = apiClient.ListDashboardsLegacy(filtersEntity)
	assert.Equal(t, 8, len(boards))
	slog.Info("Removing all dashboards")
	deleteList = apiClient.DeleteAllDashboards(filtersEntity)
	assert.Equal(t, 8, len(deleteList))
	//
	os.Setenv("GDG_CONTEXTS__TESTING__IGNORE_FILTERS", "true")
	defer os.Unsetenv("")
	apiClient, _ = test_tooling.CreateSimpleClient(t, nil, container)
	filterNone := service.NewDashboardFilter("", "", "")
	uploadedFiles, err = apiClient.UploadDashboards(filterNone)
	assert.NoError(t, err)
	assert.Equal(t, len(uploadedFiles), DashboardCount)
	// Listing with no filter
	boards = apiClient.ListDashboardsLegacy(filterNone)
	assert.Equal(t, DashboardCount, len(boards))

	data, err = json.Marshal([]string{"flow"})
	assert.NoError(t, err)
	filtersEntity = service.NewDashboardFilter("", "", string(data))

	slog.Info("Listing dashboards by tag")
	boards = apiClient.ListDashboardsLegacy(filtersEntity)
	assert.Equal(t, 8, len(deleteList))
	// Listing with
	data, err = json.Marshal([]string{"flow", "netsage"})
	assert.NoError(t, err)
	filtersEntity = service.NewDashboardFilter("", "", string(data))

	boards = apiClient.ListDashboardsLegacy(filtersEntity)
	assert.Equal(t, 13, len(boards))
	deleteList = apiClient.DeleteAllDashboards(filtersEntity)
	assert.Equal(t, 13, len(deleteList))
}

func TestDashboardTagsFilter(t *testing.T) {
	config.InitGdgConfig("testing")
	apiClient, _, cleanup := test_tooling.InitTest(t, service.DefaultConfigProvider, nil)
	defer cleanup()
	emptyFilter := v1.NewBaseFilter()

	data, err := json.Marshal([]string{"flow", "netsage"})
	assert.NoError(t, err)
	filtersEntity := service.NewDashboardFilter("", "", string(data))

	slog.Info("Exporting all dashboards")
	_, err = apiClient.UploadDashboards(emptyFilter)
	assert.NoError(t, err)

	slog.Info("Listing all dashboards")
	boards := apiClient.ListDashboardsLegacy(filtersEntity)

	slog.Info("Filtered Count is", "count", len(boards))
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
	boards = apiClient.ListDashboardsLegacy(filtersEntity)
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
	_, err = apiClient.UploadDashboards(nil)
	assert.NoError(t, err)
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
	_, err = apiClient.UploadDashboards(emptyFilter)
	assert.NoError(t, err)
	boards := apiClient.ListDashboardsLegacy(emptyFilter)

	_, err = apiClient.UploadDashboards(filtersEntity)
	assert.NoError(t, err)
	boards_filtered := apiClient.ListDashboardsLegacy(emptyFilter)

	assert.Equal(t, len(boards), len(boards_filtered))

	// Testing Listing with Wildcard
	slog.Info("Listing all dashboards without filter")
	boards = apiClient.ListDashboardsLegacy(emptyFilter)

	slog.Info("Listing all dashboards ignoring filter")
	boards_filtered = apiClient.ListDashboardsLegacy(filtersEntity)

	assert.Equal(t, 14, len(boards_filtered))

	slog.Info("Importing Dashboards")
	list := apiClient.DownloadDashboards(emptyFilter)
	assert.Equal(t, len(list), len(boards))

	slog.Info("Deleting Dashboards")
	deleteList := apiClient.DeleteAllDashboards(emptyFilter)
	assert.Equal(t, len(deleteList), len(boards))

	slog.Info("List Dashboards again")
	boards = apiClient.ListDashboardsLegacy(filtersEntity)
	assert.Equal(t, len(boards), 0)
}

func validateOtherBoard(t *testing.T, board *customModels.NestedHit) {
	assert.True(t, board.UID != "")
	assert.Equal(t, board.Title, "Flow Information")
	assert.Equal(t, board.URI, "db/flow-information")
	assert.True(t, strings.Contains(board.URL, board.UID))
	assert.True(t, strings.Contains(board.URL, board.Slug))
	assert.Equal(t, board.Type, models.HitType("dash-db"))
	assert.Equal(t, board.FolderTitle, "linux/gnu")
	assert.Equal(t, board.NestedPath, "linux%2Fgnu")
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
	assert.Equal(t, board.NestedPath, "General")
}

func validateTags(t *testing.T, board *customModels.NestedHit) {
	assert.True(t, board.UID != "")
	assert.True(t, len(board.Tags) > 0)
	allTags := []string{"netsage", "flow"}
	common := lo.Intersect(board.Tags, allTags)
	assert.True(t, len(common) > 0)
}
