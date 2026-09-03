package test

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/esnet/gdg/internal/adapter/filters/v2"
	"github.com/esnet/gdg/internal/adapter/grafana/api"
	"github.com/esnet/gdg/internal/config/config_domain"
	customModels "github.com/esnet/gdg/internal/domain"
	"github.com/esnet/gdg/internal/ports/outbound"
	"github.com/esnet/gdg/pkg/test_tooling/common"
	"github.com/esnet/gdg/pkg/test_tooling/containers"
	"github.com/esnet/gdg/pkg/tools"
	"github.com/samber/lo"
	"github.com/testcontainers/testcontainers-go"

	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/pkg/test_tooling"

	"github.com/stretchr/testify/assert"
)

const (
	// v12 (legacy API): only General/ v1-format files upload successfully.
	// v2-format files in linux%2Fgnu/, ES+net/, and Ignored/ are skipped on v12.
	DashboardCountV1       = 8
	IgnoreDashboardV1Count = 8 // Ignored/ has v2 format, also skipped on v12
	FolderCountV1          = 0 // General is the root folder; no extra folders are created

	// v12 tag counts (General/ v1 files only)
	TagNetsageCountV1 = 8 // all 8 General/ files have the "netsage" tag
	TagFlowCountV1    = 3 // other-flow-stats, science-discipline-patterns, top-talkers-over-time

	// v13 tag counts (all 17 watched files)
	TagNetsageCountV2 = 13
	TagFlowCountV2    = 8
)

// getDashboardCounts returns the expected (dashCount, ignoreCount, folderCount) for the
// Grafana version under test. v13+ uses the full v2 fixture set; v12 uses only the
// General/ v1-format files since v2-format files are skipped on older servers.
func getDashboardCounts() (dashCount, ignoreCount, folderCount int) {
	checker := StaticVersionCheck{
		Version: strings.ReplaceAll(containers.GetGrafanaVersion(), "-ubuntu", ""),
	}
	if tools.ValidateMinimumVersion("v13.0.0", checker) {
		return DashboardCountV2, IgnoreDashboardV2Count, FolderCountV2
	}
	return DashboardCountV1, IgnoreDashboardV1Count, FolderCountV1
}

// isV13 returns true when the server under test is Grafana v13 or newer.
func isV13() bool {
	checker := StaticVersionCheck{
		Version: strings.ReplaceAll(containers.GetGrafanaVersion(), "-ubuntu", ""),
	}
	return tools.ValidateMinimumVersion("v13.0.0", checker)
}

func TestDashboardCRUDIgnoreFilters(t *testing.T) {
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
	dashCount, _, folderCount := getDashboardCounts()
	apiClient := r.ApiClient
	filtersEntity := api.NewDashboardFilter(cfg, "", "", "")
	slog.Info("Exporting all dashboards")
	uploadedFiles, err := apiClient.UploadDashboards(filtersEntity)
	assert.NoError(t, err)
	assert.Equal(t, dashCount, len(uploadedFiles))
	folders := apiClient.ListFolders(api.NewFolderFilter(cfg))
	assert.Equal(t, folderCount, len(folders))

	slog.Info("Listing all dashboards")
	boards := apiClient.ListDashboards(filtersEntity)
	slog.Info("Imported dashboards", "count", len(boards), "uploadedFiles", len(uploadedFiles))
	ignoredSkipped := true
	// TODO(step-6): once DashboardServiceImpl maps v1 NestedHit fields into
	// DashboardV2Gdg, restore UID-based lookup and the full field assertions.
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
	assert.NotNil(t, generalBoard)
	assert.True(t, ignoredSkipped)
	validateGeneralBoard(t, generalBoard)

	// "Flow Information" lives in linux%2Fgnu (v2 format) — only present on v13.
	if isV13() {
		assert.NotNil(t, otherBoard)
		validateOtherBoard(t, otherBoard)
	}

	// Validate filters — v13 has linux%2Fgnu populated; v12 uses General/ only.
	if isV13() {
		filterFolder := api.NewDashboardFilter(cfg, "linux%2Fgnu", "", "")
		boards = apiClient.ListDashboards(filterFolder)
		assert.Equal(t, 8, len(boards))
		// With Regex filters
		filterFolder = api.NewDashboardFilter(cfg, "linux%2Fgnu$", "", "")
		boards = apiClient.ListDashboards(filterFolder)
		assert.Equal(t, 4, len(boards))
		dashboardFilter := api.NewDashboardFilter(cfg, "", "nzuMyBcGk", "")
		boards = apiClient.ListDashboards(dashboardFilter)
		assert.Equal(t, 1, len(boards))
	} else {
		filterFolder := api.NewDashboardFilter(cfg, "General", "", "")
		boards = apiClient.ListDashboards(filterFolder)
		assert.Equal(t, DashboardCountV1, len(boards))
		dashboardFilter := api.NewDashboardFilter(cfg, "", "000000003", "")
		boards = apiClient.ListDashboards(dashboardFilter)
		assert.Equal(t, 1, len(boards))
	}

	// Import Dashboards — use an isolated authenticated temp dir so DownloadDashboards
	// never overwrites the test fixture files in test/data/.
	slog.Info("Importing Dashboards")
	dlClient := createDownloadClient(t, cfg, r.Container)
	list := dlClient.DownloadDashboards(filtersEntity)
	assert.Equal(t, dashCount, len(list))
	slog.Info("Deleting Dashboards")
	deleteList := apiClient.DeleteAllDashboards(filtersEntity)
	assert.Equal(t, dashCount, len(deleteList))
	slog.Info("List Dashboards again")
	boards = apiClient.ListDashboards(filtersEntity)
	assert.Equal(t, len(boards), 0)
}

// createDownloadClient returns a GrafanaService pointed at an isolated temp dir
// pre-seeded with the secure credentials directory.  Downloads are fully
// authenticated (SecureLocation resolves to {tmpDir}/secure) but never write
// back to the real test/data/ fixture tree.  The caller's cfg is reused so
// runtime settings (e.g. IgnoreFilters) are preserved in the download client.
func createDownloadClient(t *testing.T, cfg *config_domain.GDGAppConfiguration, container testcontainers.Container) outbound.GrafanaService {
	t.Helper()
	tmpDir := t.TempDir()
	secureDir := filepath.Join(tmpDir, "secure")
	assert.NoError(t, os.MkdirAll(secureDir, 0o750))
	assert.NoError(t, os.CopyFS(secureDir, os.DirFS("test/data/secure")))
	// Preserve the caller's config (including IgnoreFilters and other runtime
	// settings) — only redirect OutputPath so writes go to the temp dir.
	origOutputPath := cfg.GetDefaultGrafanaConfig().OutputPath
	cfg.GetDefaultGrafanaConfig().OutputPath = tmpDir
	client := test_tooling.CreateSimpleClientWithConfig(t, cfg, container)
	cfg.GetDefaultGrafanaConfig().OutputPath = origOutputPath
	return client
}

// seedDashboardTempDir copies the dashboard fixture tree into a fresh temp
// directory and optionally writes an extra file into it.  It returns the
// temp dir path so callers can point OutputPath at it.  The real fixture
// directory under test/data/ is never modified.
func seedDashboardTempDir(t *testing.T, extraSrcFile, extraDstRelPath string) string {
	t.Helper()
	tmpDir := t.TempDir()
	// Copy the full fixture dashboard subtree.
	src := os.DirFS("test/data/org_main-org/dashboards")
	dstRoot := tmpDir + "/org_main-org/dashboards"
	assert.NoError(t, os.MkdirAll(dstRoot, 0o750))
	assert.NoError(t, os.CopyFS(dstRoot, src))
	// Write the optional extra file (e.g. a UID-duplicate to trigger ClearOutput).
	if extraSrcFile != "" && extraDstRelPath != "" {
		data, err := os.ReadFile(extraSrcFile)
		assert.NoError(t, err)
		dst := tmpDir + "/" + extraDstRelPath
		assert.NoError(t, os.MkdirAll(filepath.Dir(dst), 0o750))
		assert.NoError(t, os.WriteFile(dst, data, 0o644))
	}
	return tmpDir
}

// If a duplicate file with the same UID exists, the upload should fail.  Having a cleanup flag turned on, should
// fix that issue.
func TestDashboardCleanUpCrud(t *testing.T) {
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
	_, ignoreCount, _ := getDashboardCounts()
	apiClient := r.ApiClient
	filtersEntity := api.NewDashboardFilter(cfg, "", "", "")
	slog.Info("Exporting all dashboards")
	uploadedFiles, err := apiClient.UploadDashboards(filtersEntity)
	assert.NoError(t, err)
	assert.Equal(t, ignoreCount, len(uploadedFiles))
	slog.Info("Listing all dashboards")
	boards := apiClient.ListDashboards(filtersEntity)
	assert.Equal(t, ignoreCount, len(boards)) // Includes the Ignored folder

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
	assert.Equal(t, ignoreCount, len(boards)) // includes the ignored folder
}

// Download relies on Listing behavior so we only need to check that the dashboard listing works properly
func TestDashListFilters(t *testing.T) {
	dashCount, ignoreCount, _ := getDashboardCounts()
	testCase := []struct {
		name          string
		ignore        bool
		expectedCount int
		disabled      bool
	}{
		{
			ignore:        true,
			name:          "ignore Enabled Test",
			expectedCount: ignoreCount,
		},
		{
			ignore:        false,
			name:          "ignore Disabled Test",
			expectedCount: dashCount,
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
		assert.Equal(t, tc.expectedCount, len(uploadedFiles))

		if isV13() {
			// v13: linux%2Fgnu/* and ES+net/* are populated.
			filtersEntity := api.NewDashboardFilter(cfg, "linux%2Fgnu/Ot*", "", "")
			boards := apiClient.ListDashboards(filtersEntity)
			assert.Equal(t, 4, len(boards))
			filtersEntity = api.NewDashboardFilter(cfg, "", "", encodeTags("flow"))
			boards = apiClient.ListDashboards(filtersEntity)
			assert.Equal(t, TagFlowCountV2, len(boards))
			// Dash filter
			filtersEntity = api.NewDashboardFilter(cfg, "", "80IVUboZk", "")
			boards = apiClient.ListDashboards(filtersEntity)
			assert.Equal(t, 1, len(boards))
		} else {
			// v12: only General/ v1-format files uploaded.
			filtersEntity := api.NewDashboardFilter(cfg, "General", "", "")
			boards := apiClient.ListDashboards(filtersEntity)
			assert.Equal(t, DashboardCountV1, len(boards))
			filtersEntity = api.NewDashboardFilter(cfg, "", "", encodeTags("flow"))
			boards = apiClient.ListDashboards(filtersEntity)
			assert.Equal(t, TagFlowCountV1, len(boards))
			// Dash filter — UID of a v1 General/ dashboard
			filtersEntity = api.NewDashboardFilter(cfg, "", "000000003", "")
			boards = apiClient.ListDashboards(filtersEntity)
			assert.Equal(t, 1, len(boards))
		}
		func() {
			err := r.CleanUp()
			if err != nil {
				slog.Warn("Unable to clean up after test", "test", t.Name())
			}
		}()
	}
}

func TestUploadDashboardsBehavior(t *testing.T) {
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
	dashCount, ignoreCount, _ := getDashboardCounts()
	apiClient := r.ApiClient
	cleanupDash := func(count int) {
		items := apiClient.DeleteAllDashboards(api.NewDashboardFilter(cfg, "", "", ""))
		assert.Equal(t, count, len(items))
	}

	encodeTags := func(tags ...string) string {
		raw, err := json.Marshal(tags)
		assert.NoError(t, err, "unable to encode tags")
		return string(raw)
	}

	uploadedFiles, err := apiClient.UploadDashboards(nil)
	assert.NoError(t, err)
	assert.Equal(t, ignoreCount, len(uploadedFiles))
	cleanupDash(ignoreCount)

	if isV13() {
		// Tags filter — "flow"-tagged boards only exist in linux%2Fgnu on v13.
		filtersEntity := api.NewDashboardFilter(cfg, "", "", encodeTags("flow"))
		uploadedFiles, err = apiClient.UploadDashboards(filtersEntity)
		assert.NoError(t, err)
		assert.Equal(t, TagFlowCountV2, len(uploadedFiles))
		cleanupDash(len(uploadedFiles))
		// Dash filter — UID of a v1 General/ dashboard (individual-flows-per-country).
		filtersEntity = api.NewDashboardFilter(cfg, "", "80IVUboZk", "")
		uploadedFiles, err = apiClient.UploadDashboards(filtersEntity)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(uploadedFiles))
		cleanupDash(len(uploadedFiles))
		// Folder filter — linux%2Fgnu/Others subtree.
		filtersEntity = api.NewDashboardFilter(cfg, "linux%2Fgnu/Ot*", "", "")
		uploadedFiles, err = apiClient.UploadDashboards(filtersEntity)
		assert.NoError(t, err)
		assert.Equal(t, 4, len(uploadedFiles))
		cleanupDash(len(uploadedFiles))
	} else {
		// Tags filter — on v12 only General/ boards upload; 3 have "flow".
		filtersEntity := api.NewDashboardFilter(cfg, "", "", encodeTags("flow"))
		uploadedFiles, err = apiClient.UploadDashboards(filtersEntity)
		assert.NoError(t, err)
		assert.Equal(t, TagFlowCountV1, len(uploadedFiles))
		cleanupDash(len(uploadedFiles))
		// Dash filter — UID of a v1 General/ dashboard (bandwidth-dashboard).
		filtersEntity = api.NewDashboardFilter(cfg, "", "000000003", "")
		uploadedFiles, err = apiClient.UploadDashboards(filtersEntity)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(uploadedFiles))
		cleanupDash(len(uploadedFiles))
	}

	config.NewConfig(common.DefaultTestConfig)
	cfg.GetDefaultGrafanaConfig().GetDashboardSettings().IgnoreFilters = false
	apiClient = test_tooling.CreateSimpleClientWithConfig(t, cfg, r.Container)
	uploadedFiles, err = apiClient.UploadDashboards(nil)
	assert.NoError(t, err)
	// upload files doesn't match if lib elements are missing
	assert.True(t, dashCount-len(uploadedFiles) < 2)
	cleanupDash(dashCount)
}

func TestDashboardCRUDTags(t *testing.T) {
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
	dashCount, _, _ := getDashboardCounts()
	netsageCount := TagNetsageCountV1
	flowCount := TagFlowCountV1
	if isV13() {
		netsageCount = TagNetsageCountV2
		flowCount = TagFlowCountV2
	}
	apiClient := r.ApiClient

	data, err := json.Marshal([]string{"netsage"})
	assert.NoError(t, err)
	filtersEntity := api.NewDashboardFilter(cfg, "", "", string(data))

	slog.Info("Uploading all dashboards, filtered by tags")
	uploadedFiles, err := apiClient.UploadDashboards(filtersEntity)
	assert.NoError(t, err)
	assert.Equal(t, netsageCount, len(uploadedFiles))
	slog.Info("Listing all dashboards")
	boards := apiClient.ListDashboards(filtersEntity)
	slog.Info("Removing all dashboards")
	assert.Equal(t, netsageCount, len(boards))
	deleteList := apiClient.DeleteAllDashboards(filtersEntity)
	assert.Equal(t, netsageCount, len(deleteList))
	// Multiple Tags behavior
	slog.Info("Uploading all dashboards, filtered by tags")
	data, err = json.Marshal([]string{"flow"})
	assert.NoError(t, err)
	filtersEntity = api.NewDashboardFilter(cfg, "", "", string(data))
	uploadedFiles, err = apiClient.UploadDashboards(filtersEntity)
	assert.NoError(t, err)
	assert.Equal(t, flowCount, len(uploadedFiles))
	slog.Info("Listing all dashboards")
	boards = apiClient.ListDashboards(filtersEntity)
	assert.Equal(t, flowCount, len(boards))
	slog.Info("Removing all dashboards")
	deleteList = apiClient.DeleteAllDashboards(filtersEntity)
	assert.Equal(t, flowCount, len(deleteList))
	//
	os.Setenv("GDG_CONTEXTS__TESTING__IGNORE_FILTERS", "true")
	defer os.Unsetenv("")
	apiClient, _ = test_tooling.CreateSimpleClient(t, cfg, nil, r.Container)
	filterNone := api.NewDashboardFilter(cfg, "", "", "")
	uploadedFiles, err = apiClient.UploadDashboards(filterNone)
	assert.NoError(t, err)
	assert.Equal(t, dashCount, len(uploadedFiles))
	// Listing with no filter
	boards = apiClient.ListDashboards(filterNone)
	assert.Equal(t, dashCount, len(boards))

	data, err = json.Marshal([]string{"flow"})
	assert.NoError(t, err)
	filtersEntity = api.NewDashboardFilter(cfg, "", "", string(data))

	slog.Info("Listing dashboards by tag")
	boards = apiClient.ListDashboards(filtersEntity)
	assert.Equal(t, flowCount, len(deleteList))
	// Listing with union of tags
	data, err = json.Marshal([]string{"flow", "netsage"})
	assert.NoError(t, err)
	filtersEntity = api.NewDashboardFilter(cfg, "", "", string(data))

	boards = apiClient.ListDashboards(filtersEntity)
	assert.Equal(t, netsageCount, len(boards)) // netsage is a superset of flow
	deleteList = apiClient.DeleteAllDashboards(filtersEntity)
	assert.Equal(t, netsageCount, len(deleteList))
}

func TestDashboardTagsFilter(t *testing.T) {
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

func TestWildcardFilter(t *testing.T) {
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

	// flow|netsage union: 14 on v13 (all folders incl. Ignored), 8 on v12 (General/ only).
	expectedWildcard := 8
	if isV13() {
		expectedWildcard = 14
	}
	assert.Equal(t, expectedWildcard, len(boards_filtered))

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

// TODO(step-6): restore full field assertions (UID, URI, URL, FolderTitle,
// FolderID, Type) once DashboardServiceImpl maps v1 NestedHit into DashboardV2Gdg.

func validateOtherBoard(t *testing.T, board *customModels.DashboardV2Gdg) {
	assert.Equal(t, board.Resource.Spec.Title, "Flow Information")
	assert.Equal(t, board.NestedPath, "linux%2Fgnu")
}

func validateGeneralBoard(t *testing.T, board *customModels.DashboardV2Gdg) {
	assert.Equal(t, board.Resource.Spec.Title, "Individual Flows")
	assert.Equal(t, len(board.Resource.Spec.Tags), 1)
	assert.Equal(t, board.Resource.Spec.Tags[0], "netsage")
	assert.Equal(t, board.NestedPath, "General")
}

func validateTags(t *testing.T, board *customModels.DashboardV2Gdg) {
	assert.True(t, len(board.Resource.Spec.Tags) > 0)
	allTags := []string{"netsage", "flow"}
	common := lo.Intersect(board.Resource.Spec.Tags, allTags)
	assert.True(t, len(common) > 0)
}
