package test

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"strings"
	"testing"

	customModels "github.com/esnet/gdg/internal/service/domain"

	"github.com/esnet/gdg/pkg/test_tooling/common"

	"github.com/esnet/gdg/internal/service/filters/v2"
	"github.com/samber/lo"

	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service"
	"github.com/esnet/gdg/pkg/test_tooling"
	"github.com/grafana/grafana-openapi-client-go/models"

	"github.com/stretchr/testify/assert"
)

const (
	DashboardCount       = 17
	IgnoreDashboardCount = DashboardCount + 1
	FolderCount          = 4
)

func TestDashboardCRUDIgnoreFilters(t *testing.T) {
	config.InitGdgConfig(common.DefaultTestConfig)
	var r *test_tooling.InitContainerResult
	err := Retry(context.Background(), DefaultRetryAttempts, func() error {
		r = test_tooling.InitTest(t, service.DefaultConfigProvider, nil)
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
	filtersEntity := service.NewDashboardFilter("", "", "")
	slog.Info("Exporting all dashboards")
	uploadedFiles, err := apiClient.UploadDashboards(filtersEntity)
	assert.NoError(t, err)
	assert.Equal(t, len(uploadedFiles), DashboardCount)
	folders := apiClient.ListFolders(service.NewFolderFilter())
	assert.Equal(t, len(folders), FolderCount)

	slog.Info("Listing all dashboards")
	boards := apiClient.ListDashboards(filtersEntity)
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
	assert.True(t, ignoredSkipped)
	validateGeneralBoard(t, generalBoard)
	validateOtherBoard(t, otherBoard)
	// Validate filters

	filterFolder := service.NewDashboardFilter("linux%2Fgnu", "", "")
	boards = apiClient.ListDashboards(filterFolder)
	assert.Equal(t, 8, len(boards))
	// With Regex filters
	filterFolder = service.NewDashboardFilter("linux%2Fgnu$", "", "")
	boards = apiClient.ListDashboards(filterFolder)
	assert.Equal(t, 4, len(boards))
	//
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
	config.InitGdgConfig(common.DefaultTestConfig)
	ctx := context.Background()
	cfgProvider := func() *config.Configuration {
		cfg := config.Config()
		cfg.GetDefaultGrafanaConfig().GetDashboardSettings().IgnoreFilters = true
		return cfg
	}
	var r *test_tooling.InitContainerResult
	err := Retry(ctx, DefaultRetryAttempts, func() error {
		r = test_tooling.InitTest(t, cfgProvider, nil)
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
	filtersEntity := service.NewDashboardFilter("", "", "")
	slog.Info("Exporting all dashboards")
	uploadedFiles, err := apiClient.UploadDashboards(filtersEntity)
	assert.NoError(t, err)
	assert.Equal(t, len(uploadedFiles), IgnoreDashboardCount)
	slog.Info("Listing all dashboards")
	boards := apiClient.ListDashboards(filtersEntity)
	assert.Equal(t, len(boards), IgnoreDashboardCount) // Includes the Ignored folder
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
	apiClient = test_tooling.CreateSimpleClientWithConfig(t, cfgProvider, r.Container)
	apiClient.DownloadDashboards(filtersEntity)
	assert.Nil(t, err)
	boards = apiClient.ListDashboards(filtersEntity)
	assert.Equal(t, len(boards), IgnoreDashboardCount) // includes the ignored folder
}

// Download relies on Listing behavior so we only need to check that the dashboard listing works properly
func TestDashListFilters(t *testing.T) {
	testCase := []struct {
		name          string
		ignore        bool
		expectedCount int
		disabled      bool
	}{
		{
			ignore:        true,
			name:          "ignore Enabled Test",
			expectedCount: IgnoreDashboardCount,
		},
		{
			ignore:        false,
			name:          "ignore Disabled Test",
			expectedCount: DashboardCount,
		},
	}
	for _, tc := range testCase {
		t.Log("Running test", tc.name)
		if tc.disabled {
			continue
		}
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

		apiClient := r.ApiClient
		encodeTags := func(tags ...string) string {
			raw, err := json.Marshal(tags)
			assert.NoError(t, err, "unable to encode tags")
			return string(raw)
		}
		uploadedFiles, err := apiClient.UploadDashboards(nil)
		assert.NoError(t, err)
		assert.Equal(t, len(uploadedFiles), tc.expectedCount)
		// folder test
		filtersEntity := service.NewDashboardFilter("linux%2Fgnu/Ot*", "", "")
		boards := apiClient.ListDashboards(filtersEntity)
		assert.Equal(t, len(boards), 4)
		//
		filtersEntity = service.NewDashboardFilter("", "", encodeTags("flow"))
		boards = apiClient.ListDashboards(filtersEntity)
		assert.Equal(t, len(boards), 8)
		// Dash filter
		filtersEntity = service.NewDashboardFilter("", "individual-flows-per-country", "")
		boards = apiClient.ListDashboards(filtersEntity)
		assert.Equal(t, len(boards), 1)
		func() {
			err := r.CleanUp()
			if err != nil {
				slog.Warn("Unable to clean up after test", "test", t.Name())
			}
		}()
	}
}

func TestUploadDashboardsBehavior(t *testing.T) {
	config.InitGdgConfig(common.DefaultTestConfig)
	cfgProvider := func() *config.Configuration {
		cfg := config.Config()
		cfg.GetDefaultGrafanaConfig().GetDashboardSettings().IgnoreFilters = true
		return cfg
	}
	var r *test_tooling.InitContainerResult
	err := Retry(context.Background(), DefaultRetryAttempts, func() error {
		r = test_tooling.InitTest(t, cfgProvider, nil)
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
	// Tags filter
	filtersEntity := service.NewDashboardFilter("", "", encodeTags("flow"))
	uploadedFiles, err = apiClient.UploadDashboards(filtersEntity)
	assert.NoError(t, err)
	assert.Equal(t, len(uploadedFiles), 8)
	cleanupDash(len(uploadedFiles))
	// Dash filter
	filtersEntity = service.NewDashboardFilter("", "individual-flows-per-country", "")
	uploadedFiles, err = apiClient.UploadDashboards(filtersEntity)
	assert.NoError(t, err)
	assert.Equal(t, len(uploadedFiles), 1)
	cleanupDash(len(uploadedFiles))
	// folder test
	filtersEntity = service.NewDashboardFilter("linux%2Fgnu/Ot*", "", "")
	uploadedFiles, err = apiClient.UploadDashboards(filtersEntity)
	assert.NoError(t, err)
	assert.Equal(t, len(uploadedFiles), 4)
	cleanupDash(len(uploadedFiles))
	//
	wrapTest(func() {
		config.InitGdgConfig(common.DefaultTestConfig)
	})
	cfgProvider = func() *config.Configuration {
		cfg := config.Config()
		cfg.GetDefaultGrafanaConfig().GetDashboardSettings().IgnoreFilters = false
		return cfg
	}
	apiClient = test_tooling.CreateSimpleClientWithConfig(t, cfgProvider, r.Container)
	uploadedFiles, err = apiClient.UploadDashboards(nil)
	assert.NoError(t, err)
	// upload files doesn't match if lib elements are missing
	assert.True(t, DashboardCount-len(uploadedFiles) < 2.0)
	cleanupDash(DashboardCount)
}

func TestDashboardCRUDTags(t *testing.T) {
	config.InitGdgConfig(common.DefaultTestConfig)
	var r *test_tooling.InitContainerResult
	err := Retry(context.Background(), DefaultRetryAttempts, func() error {
		r = test_tooling.InitTest(t, service.DefaultConfigProvider, nil)
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

	data, err := json.Marshal([]string{"netsage"})
	assert.NoError(t, err)
	filtersEntity := service.NewDashboardFilter("", "", string(data))

	slog.Info("Uploading all dashboards, filtered by tags")
	uploadedFiles, err := apiClient.UploadDashboards(filtersEntity)
	assert.Equal(t, len(uploadedFiles), 13)
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
	uploadedFiles, err = apiClient.UploadDashboards(filtersEntity)
	assert.NoError(t, err)
	assert.Equal(t, len(uploadedFiles), 8)
	slog.Info("Listing all dashboards")
	boards = apiClient.ListDashboards(filtersEntity)
	assert.Equal(t, 8, len(boards))
	slog.Info("Removing all dashboards")
	deleteList = apiClient.DeleteAllDashboards(filtersEntity)
	assert.Equal(t, 8, len(deleteList))
	//
	os.Setenv("GDG_CONTEXTS__TESTING__IGNORE_FILTERS", "true")
	defer os.Unsetenv("")
	apiClient, _ = test_tooling.CreateSimpleClient(t, nil, r.Container)
	filterNone := service.NewDashboardFilter("", "", "")
	uploadedFiles, err = apiClient.UploadDashboards(filterNone)
	assert.NoError(t, err)
	assert.Equal(t, len(uploadedFiles), DashboardCount)
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
	config.InitGdgConfig(common.DefaultTestConfig)
	var r *test_tooling.InitContainerResult
	err := Retry(context.Background(), DefaultRetryAttempts, func() error {
		r = test_tooling.InitTest(t, service.DefaultConfigProvider, nil)
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
	emptyFilter := v2.NewBaseFilter()

	data, err := json.Marshal([]string{"flow", "netsage"})
	assert.NoError(t, err)
	filtersEntity := service.NewDashboardFilter("", "", string(data))

	slog.Info("Exporting all dashboards")
	_, err = apiClient.UploadDashboards(emptyFilter)
	assert.NoError(t, err)

	slog.Info("Listing all dashboards")
	boards := apiClient.ListDashboards(filtersEntity)

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
	boards = apiClient.ListDashboards(filtersEntity)
	assert.Equal(t, len(boards), 0)
}

func TestWildcardFilter(t *testing.T) {
	// Setup Filters
	config.InitGdgConfig(common.DefaultTestConfig)
	var r *test_tooling.InitContainerResult
	err := Retry(context.Background(), DefaultRetryAttempts, func() error {
		r = test_tooling.InitTest(t, service.DefaultConfigProvider, nil)
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
	emptyFilter := service.NewDashboardFilter("", "", "")

	data, err := json.Marshal([]string{"flow", "netsage"})
	assert.NoError(t, err)
	filtersEntity := service.NewDashboardFilter("", "", string(data))

	// Enable Wildcard
	testingContext := config.Config().GetGDGConfig().GetContexts()[common.TestContextName]
	testingContext.GetDashboardSettings().IgnoreFilters = true
	assert.True(t, testingContext.GetDashboardSettings().IgnoreFilters)

	// Testing Exporting with Wildcard
	_, err = apiClient.UploadDashboards(emptyFilter)
	assert.NoError(t, err)
	boards := apiClient.ListDashboards(emptyFilter)

	_, err = apiClient.UploadDashboards(filtersEntity)
	assert.NoError(t, err)
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
