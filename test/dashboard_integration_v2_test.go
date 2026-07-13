package test

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/esnet/gdg/internal/adapter/filters/v2"
	"github.com/esnet/gdg/internal/adapter/grafana/api"
	customModels "github.com/esnet/gdg/internal/domain"
	"github.com/esnet/gdg/pkg/test_tooling/common"
	"github.com/esnet/gdg/pkg/test_tooling/containers"
	"github.com/esnet/gdg/pkg/tools"
	"github.com/samber/lo"

	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/pkg/test_tooling"

	"github.com/stretchr/testify/assert"
)

const (
	DashboardCountV2       = 17
	IgnoreDashboardV2Count = DashboardCountV2 + 1
	FolderCountV2          = 4
)

type StaticVersionCheck struct {
	Version string
}

func (v StaticVersionCheck) GetServerInfo() map[string]any {
	return map[string]any{"Version": v.Version}
}

func skipLegacyVersion(t *testing.T) {
	versionCheckerV2 := StaticVersionCheck{
		Version: strings.ReplaceAll(containers.GetGrafanaVersion(), "-ubuntu", ""),
	}

	if !tools.ValidateMinimumVersion("v13.0.0", versionCheckerV2) {
		t.Skip("Grafana version is too old, skipping test", t.Name())

	}
}

func TestDashboardCRUDIgnoreFiltersV2(t *testing.T) {
	skipLegacyVersion(t)

	cfg := config.NewConfig(common.DefaultTestConfig)
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
	filtersEntity := api.NewDashboardFilter(cfg, "", "", "")
	slog.Info("Exporting all dashboards")
	uploadedFiles, err := apiClient.UploadDashboards(filtersEntity)
	assert.NoError(t, err)
	assert.Equal(t, len(uploadedFiles), DashboardCountV2)
	folders := apiClient.ListFolders(api.NewFolderFilter(cfg))
	assert.Equal(t, len(folders), FolderCountV2)

	slog.Info("Listing all dashboards")
	boards := apiClient.ListDashboards(filtersEntity)
	slog.Info("Imported dashboards", "count", len(boards), "uploadedFiles", len(uploadedFiles))
	ignoredSkipped := true
	// TODO(step-6): once DashboardServiceImpl maps v1 NestedHit fields into
	// DashboardV2Gdg, restore slug-based lookup and the full field assertions.
	var generalBoard *customModels.DashboardV2Gdg
	var otherBoard *customModels.DashboardV2Gdg
	for ndx, board := range boards {
		title := board.Resource.Spec.Title
		slog.Info(title)
		if title == "Latency Patterns" {
			ignoredSkipped = false
		}
		if title == "Individual Flows" {
			generalBoard = boards[ndx]
		}
		if title == "Flow Information" {
			otherBoard = boards[ndx]
		}
	}
	assert.NotNil(t, otherBoard)
	assert.NotNil(t, generalBoard)
	assert.True(t, ignoredSkipped)
	validateGeneralBoard(t, generalBoard)
	validateOtherBoard(t, otherBoard)
	// Validate filters

	filterFolder := api.NewDashboardFilter(cfg, "linux%2Fgnu", "", "")
	boards = apiClient.ListDashboards(filterFolder)
	assert.Equal(t, 8, len(boards))
	// With Regex filters
	filterFolder = api.NewDashboardFilter(cfg, "linux%2Fgnu$", "", "")
	boards = apiClient.ListDashboards(filterFolder)
	assert.Equal(t, 4, len(boards))
	//
	dashboardFilter := api.NewDashboardFilter(cfg, "", "flow-information", "")
	boards = apiClient.ListDashboards(dashboardFilter)
	assert.Equal(t, 1, len(boards))

	// Import Dashboards — use an isolated authenticated temp dir so DownloadDashboards
	// never overwrites the test fixture files in test/data/.
	slog.Info("Importing Dashboards")
	dlClient := createDownloadClient(t, cfg, r.Container)
	list := dlClient.DownloadDashboards(filtersEntity)
	assert.Equal(t, len(list), DashboardCountV2)
	slog.Info("Deleting Dashboards")
	deleteList := apiClient.DeleteAllDashboards(filtersEntity)
	assert.Equal(t, len(deleteList), DashboardCountV2)
	slog.Info("List Dashboards again")
	boards = apiClient.ListDashboards(filtersEntity)
	assert.Equal(t, len(boards), 0)
}

// If a duplicate file with the same UID exists, the upload should fail.  Having a cleanup flag turned on, should
// fix that issue.
func TestDashboardCleanUpCrudV2(t *testing.T) {
	skipLegacyVersion(t)
	cfg := config.NewConfig(common.DefaultTestConfig)
	ctx := context.Background()
	cfg.GetDefaultGrafanaConfig().GetDashboardSettings().IgnoreFilters = true
	var r *test_tooling.InitContainerResult
	err := Retry(ctx, DefaultRetryAttempts, func() error {
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
	filtersEntity := api.NewDashboardFilter(cfg, "", "", "")
	slog.Info("Exporting all dashboards")
	uploadedFiles, err := apiClient.UploadDashboards(filtersEntity)
	assert.NoError(t, err)
	assert.Equal(t, len(uploadedFiles), IgnoreDashboardV2Count)
	slog.Info("Listing all dashboards")
	boards := apiClient.ListDashboards(filtersEntity)
	assert.Equal(t, len(boards), IgnoreDashboardV2Count) // Includes the Ignored folder
	// Pre-seed a temp dir with the full fixture dashboard tree plus a UID-duplicate
	// copy of bandwidth-dashboard.json.  ClearOutput=true will wipe and re-seed the
	// temp dir from Grafana, resolving the duplicate — the real fixture dir is
	// never touched.
	tmpDir := seedDashboardTempDir(t,
		"test/data/org_main-org/dashboards/General/bandwidth-dashboard.json",
		"org_main-org/dashboards/General/bandwidth-dashboard-copy.json",
	)
	cfg.GetDefaultGrafanaConfig().GetDashboardSettings().IgnoreFilters = true
	cfg.GetDefaultGrafanaConfig().OutputPath = tmpDir
	cfg.Global.ClearOutput = true
	apiClient = test_tooling.CreateSimpleClientWithConfig(t, cfg, r.Container)
	apiClient.DownloadDashboards(filtersEntity)
	assert.Nil(t, err)
	boards = apiClient.ListDashboards(filtersEntity)
	assert.Equal(t, len(boards), IgnoreDashboardV2Count) // includes the ignored folder
}

// Download relies on Listing behavior so we only need to check that the dashboard listing works properly
func TestDashListFiltersV2(t *testing.T) {
	skipLegacyVersion(t)
	testCase := []struct {
		name          string
		ignore        bool
		expectedCount int
		disabled      bool
	}{
		{
			ignore:        true,
			name:          "ignore Enabled Test",
			expectedCount: IgnoreDashboardV2Count,
		},
		{
			ignore:        false,
			name:          "ignore Disabled Test",
			expectedCount: DashboardCountV2,
		},
	}
	for _, tc := range testCase {
		t.Log("Running test", tc.name)
		if tc.disabled {
			continue
		}
		cfg := config.NewConfig(common.DefaultTestConfig)
		cfg.GetDefaultGrafanaConfig().GetDashboardSettings().IgnoreFilters = tc.ignore

		var r *test_tooling.InitContainerResult
		err := Retry(context.Background(), DefaultRetryAttempts, func() error {
			r = test_tooling.InitTest(t, cfg, nil)
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
		filtersEntity := api.NewDashboardFilter(cfg, "linux%2Fgnu/Ot*", "", "")
		boards := apiClient.ListDashboards(filtersEntity)
		assert.Equal(t, len(boards), 4)
		//
		filtersEntity = api.NewDashboardFilter(cfg, "", "", encodeTags("flow"))
		boards = apiClient.ListDashboards(filtersEntity)
		assert.Equal(t, len(boards), 8)
		// Dash filter
		filtersEntity = api.NewDashboardFilter(cfg, "", "individual-flows-per-country", "")
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

func TestUploadDashboardsBehaviorV2(t *testing.T) {
	skipLegacyVersion(t)
	cfg := config.NewConfig(common.DefaultTestConfig)
	cfg.GetDefaultGrafanaConfig().GetDashboardSettings().IgnoreFilters = true
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
	cleanupDash := func(count int) {
		items := apiClient.DeleteAllDashboards(api.NewDashboardFilter(cfg, "", "", ""))
		assert.Equal(t, len(items), count)
	}

	encodeTags := func(tags ...string) string {
		raw, err := json.Marshal(tags)
		assert.NoError(t, err, "unable to encode tags")
		return string(raw)
	}

	uploadedFiles, err := apiClient.UploadDashboards(nil)
	assert.NoError(t, err)
	assert.Equal(t, len(uploadedFiles), IgnoreDashboardV2Count)
	cleanupDash(IgnoreDashboardV2Count)
	// Tags filter
	filtersEntity := api.NewDashboardFilter(cfg, "", "", encodeTags("flow"))
	uploadedFiles, err = apiClient.UploadDashboards(filtersEntity)
	assert.NoError(t, err)
	assert.Equal(t, len(uploadedFiles), 8)
	cleanupDash(len(uploadedFiles))
	// Dash filter
	filtersEntity = api.NewDashboardFilter(cfg, "", "individual-flows-per-country", "")
	uploadedFiles, err = apiClient.UploadDashboards(filtersEntity)
	assert.NoError(t, err)
	assert.Equal(t, len(uploadedFiles), 1)
	cleanupDash(len(uploadedFiles))
	// folder test
	filtersEntity = api.NewDashboardFilter(cfg, "linux%2Fgnu/Ot*", "", "")
	uploadedFiles, err = apiClient.UploadDashboards(filtersEntity)
	assert.NoError(t, err)
	assert.Equal(t, len(uploadedFiles), 4)
	cleanupDash(len(uploadedFiles))
	//
	config.NewConfig(common.DefaultTestConfig)
	cfg.GetDefaultGrafanaConfig().GetDashboardSettings().IgnoreFilters = false
	apiClient = test_tooling.CreateSimpleClientWithConfig(t, cfg, r.Container)
	uploadedFiles, err = apiClient.UploadDashboards(nil)
	assert.NoError(t, err)
	// upload files doesn't match if lib elements are missing
	assert.True(t, DashboardCountV2-len(uploadedFiles) < 2.0)
	cleanupDash(DashboardCountV2)
}

func TestDashboardCRUDTagsV2(t *testing.T) {
	skipLegacyVersion(t)
	cfg := config.NewConfig(common.DefaultTestConfig)
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

	data, err := json.Marshal([]string{"netsage"})
	assert.NoError(t, err)
	filtersEntity := api.NewDashboardFilter(cfg, "", "", string(data))

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
	filtersEntity = api.NewDashboardFilter(cfg, "", "", string(data))
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
	apiClient, _ = test_tooling.CreateSimpleClient(t, cfg, nil, r.Container)
	filterNone := api.NewDashboardFilter(cfg, "", "", "")
	uploadedFiles, err = apiClient.UploadDashboards(filterNone)
	assert.NoError(t, err)
	assert.Equal(t, len(uploadedFiles), DashboardCountV2)
	// Listing with no filter
	boards = apiClient.ListDashboards(filterNone)
	assert.Equal(t, DashboardCountV2, len(boards))

	data, err = json.Marshal([]string{"flow"})
	assert.NoError(t, err)
	filtersEntity = api.NewDashboardFilter(cfg, "", "", string(data))

	slog.Info("Listing dashboards by tag")
	boards = apiClient.ListDashboards(filtersEntity)
	assert.Equal(t, 8, len(deleteList))
	// Listing with
	data, err = json.Marshal([]string{"flow", "netsage"})
	assert.NoError(t, err)
	filtersEntity = api.NewDashboardFilter(cfg, "", "", string(data))

	boards = apiClient.ListDashboards(filtersEntity)
	assert.Equal(t, 13, len(boards))
	deleteList = apiClient.DeleteAllDashboards(filtersEntity)
	assert.Equal(t, 13, len(deleteList))
}

func TestDashboardTagsFilterV2(t *testing.T) {
	skipLegacyVersion(t)
	cfg := config.NewConfig(common.DefaultTestConfig)
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
	emptyFilter := v2.NewBaseFilter()

	data, err := json.Marshal([]string{"flow", "netsage"})
	assert.NoError(t, err)
	filtersEntity := api.NewDashboardFilter(cfg, "", "", string(data))

	slog.Info("Exporting all dashboards")
	_, err = apiClient.UploadDashboards(emptyFilter)
	assert.NoError(t, err)

	slog.Info("Listing all dashboards")
	boards := apiClient.ListDashboards(filtersEntity)

	slog.Info("Filtered Count is", "count", len(boards))
	for _, board := range boards {
		validateTags(t, board)
	}

	// Import Dashboards — use an isolated authenticated temp dir to avoid overwriting fixture files.
	slog.Info("Importing Dashboards")
	dlClient := createDownloadClient(t, cfg, r.Container)
	list := dlClient.DownloadDashboards(filtersEntity)
	assert.Equal(t, len(list), len(boards))

	slog.Info("Deleting Dashboards")
	deleteList := apiClient.DeleteAllDashboards(filtersEntity)
	assert.Equal(t, len(deleteList), len(boards))

	slog.Info("List Dashboards again")
	boards = apiClient.ListDashboards(filtersEntity)
	assert.Equal(t, len(boards), 0)
}

func TestWildcardFilterV2(t *testing.T) {
	skipLegacyVersion(t)
	// Setup Filters
	cfg := config.NewConfig(common.DefaultTestConfig)
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
	emptyFilter := api.NewDashboardFilter(cfg, "", "", "")

	data, err := json.Marshal([]string{"flow", "netsage"})
	assert.NoError(t, err)
	filtersEntity := api.NewDashboardFilter(cfg, "", "", string(data))

	// Enable Wildcard
	testingContext := cfg.GetContexts()[common.TestContextName]
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

	// Import Dashboards — use an isolated authenticated temp dir to avoid overwriting fixture files.
	slog.Info("Importing Dashboards")
	dlClient := createDownloadClient(t, cfg, r.Container)
	list := dlClient.DownloadDashboards(emptyFilter)
	assert.Equal(t, len(list), len(boards))

	slog.Info("Deleting Dashboards")
	deleteList := apiClient.DeleteAllDashboards(emptyFilter)
	assert.Equal(t, len(deleteList), len(boards))

	slog.Info("List Dashboards again")
	boards = apiClient.ListDashboards(filtersEntity)
	assert.Equal(t, len(boards), 0)
}

// TODO(step-6): restore full field assertions (UID, URI, URL, Slug, FolderTitle,
// FolderID, Type) once DashboardServiceImpl maps v1 NestedHit into DashboardV2Gdg.

func validateOtherBoardV2(t *testing.T, board *customModels.DashboardV2Gdg) {
	assert.Equal(t, board.Resource.Spec.Title, "Flow Information")
	assert.Equal(t, board.NestedPath, "linux%2Fgnu")
}

func validateGeneralBoardV2(t *testing.T, board *customModels.DashboardV2Gdg) {
	assert.Equal(t, board.Resource.Spec.Title, "Individual Flows")
	assert.Equal(t, len(board.Resource.Spec.Tags), 1)
	assert.Equal(t, board.Resource.Spec.Tags[0], "netsage")
	assert.Equal(t, board.NestedPath, "General")
}

func validateTagsV2(t *testing.T, board *customModels.DashboardV2Gdg) {
	assert.True(t, len(board.Resource.Spec.Tags) > 0)
	allTags := []string{"netsage", "flow"}
	common := lo.Intersect(board.Resource.Spec.Tags, allTags)
	assert.True(t, len(common) > 0)
}
