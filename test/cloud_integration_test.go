package test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/esnet/gdg/pkg/test_tooling/common"

	"github.com/esnet/gdg/internal/config"

	"github.com/esnet/gdg/internal/service"
	"github.com/esnet/gdg/pkg/test_tooling"
	"github.com/esnet/gdg/pkg/test_tooling/path"
	"github.com/stretchr/testify/assert"
	_ "gocloud.dev/blob/memblob"
)

func TestCloudDataSourceCRUD(t *testing.T) {
	t.Log("Running Cloud Tests")
	assert.NoError(t, path.FixTestDir("test", ".."))
	assert.NoError(t, os.Setenv("GDG_CONTEXT_NAME", common.TestContextName))
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
	// Wipe all data from grafana
	dsFilter := service.NewConnectionFilter("")
	apiClient.DeleteAllConnections(dsFilter)

	apiClient.UploadConnections(dsFilter)
	dsList := apiClient.ListConnections(dsFilter)
	assert.True(t, len(dsList) > 0)
	_, cancel, apiClient, err := test_tooling.SetupCloudFunctionOpt(
		test_tooling.SetCloudType("custom"),
		test_tooling.SetBucketName(common.TestBucketName))
	assert.NoError(t, err)
	defer cancel()

	slog.Info("Importing DataSources")
	dsStringList := apiClient.DownloadConnections(dsFilter) // Saving to S3
	assert.Equal(t, len(dsList), len(dsStringList))
	slog.Info("Deleting DataSources")
	deleteDSList := apiClient.DeleteAllConnections(dsFilter) // Cleaning up Grafana
	assert.Equal(t, len(deleteDSList), len(dsStringList))
	dsList = apiClient.ListConnections(dsFilter)
	assert.Equal(t, len(dsList), 0)
	// Load Data from S3
	apiClient.UploadConnections(dsFilter) // Load data from S3
	dsList = apiClient.ListConnections(dsFilter)
	assert.Equal(t, len(dsList), len(dsStringList))
	apiClient.DeleteAllConnections(dsFilter) // Cleaning up Grafana
}

// TestDashboardCloudCrud will load testing_data to Grafana from local context.  Switch to CLoud,
// Save all data to Cloud, wipe grafana and reload data back into grafana and validate
func TestDashboardCloudCRUD(t *testing.T) {
	assert.NoError(t, os.Setenv("GDG_CONTEXT_NAME", common.TestContextName))
	assert.NoError(t, path.FixTestDir("test", ".."))
	var err error
	config.InitGdgConfig(common.DefaultTestConfig)
	var r *test_tooling.InitContainerResult
	err = Retry(context.Background(), DefaultRetryAttempts, func() error {
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
	// defer cleanup, "Failed to cleanup test containers for %s", t.Name())
	// Wipe all data from grafana
	dashFilter := service.NewDashboardFilter("", "", "")
	apiClient.DeleteAllDashboards(dashFilter)
	// Load data into grafana
	apiClient.UploadDashboards(dashFilter)
	boards := apiClient.ListDashboards(dashFilter)
	assert.True(t, len(boards) > 0)
	var cancel context.CancelFunc

	_, cancel, apiClient, err = test_tooling.SetupCloudFunctionOpt(
		test_tooling.SetCloudType("custom"),
		test_tooling.SetBucketName(common.TestBucketName))
	assert.NoError(t, err)
	defer cancel()

	// At this point all operations are reading/writing from Minio
	slog.Info("Importing Dashboards")
	list := apiClient.DownloadDashboards(dashFilter) // Saving to S3
	assert.Equal(t, len(list), len(boards))
	slog.Info("Deleting Dashboards") // Clearing Grafana
	deleteList := apiClient.DeleteAllDashboards(dashFilter)
	assert.Equal(t, len(list), len(deleteList))
	boards = apiClient.ListDashboards(dashFilter)
	assert.Equal(t, len(boards), 0)
	// Load Data from S3
	apiClient.UploadDashboards(dashFilter)        // ReLoad data from S3 backup
	boards = apiClient.ListDashboards(dashFilter) // Read data
	assert.Equal(t, len(list), len(boards))       // verify
	apiClient.DeleteAllDashboards(dashFilter)
}

func TestDashboardCloudLeadingSlashCRUD(t *testing.T) {
	assert.NoError(t, os.Setenv("GDG_CONTEXT_NAME", common.TestContextName))
	assert.NoError(t, path.FixTestDir("test", ".."))
	var (
		err    error
		cancel context.CancelFunc
		r      *test_tooling.InitContainerResult
	)
	test_tooling.WrapTest(func() {
		config.InitGdgConfig(common.DefaultTestConfig)
	})
	err = Retry(context.Background(), DefaultRetryAttempts, func() error {
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
	// defer cleanup, "Failed to clean up test containers for %s", t.Name())
	// Wipe all data from grafana
	dashFilter := service.NewDashboardFilter("", "", "")
	apiClient.DeleteAllDashboards(dashFilter)
	// Load data into grafana
	_, err = apiClient.UploadDashboards(dashFilter)
	assert.NoError(t, err)
	boards := apiClient.ListDashboards(dashFilter)
	assert.True(t, len(boards) > 0)

	// Tests all type of combination that can potential break things for cloud + test output config
	testcases := []struct {
		disabled bool
		name     string
		prefix   string
		output   string
		id       int
	}{
		{
			name:   "base default test",
			prefix: "dummy",
			output: "test/data",
		},
		{
			name:   "no prefix",
			prefix: "",
			output: "test/data",
		},
		{
			name:   "no prefix, slash output",
			prefix: "",
			output: "/test/data",
		},
		{
			name:   "/prefix and no output",
			prefix: "/dummy",
			output: "",
			id:     5,
		},
		{
			name:   "/prefix and no slash output",
			prefix: "/dummy",
			output: "test/data",
		},
		{
			name:   "/prefix and /output",
			prefix: "/dummy",
			output: "/test/data",
		},
		{
			name:   "/prefix and no output",
			prefix: "/dummy",
			output: "",
		},
	}

	for _, tc := range testcases {
		if tc.disabled {
			slog.Info("Skipping test, disabled", "name", tc.name)
			continue
		}
		slog.Warn("Running testcase", "name", tc.name)
		test_tooling.MaintainConfigAuth(common.DefaultTestConfig)
		_, cancel, apiClient, err = test_tooling.SetupCloudFunctionOpt(
			test_tooling.SetCloudType("custom"),
			test_tooling.SetPrefix(tc.prefix),
			test_tooling.SetBucketName(common.TestBucketName))

		apiClient = test_tooling.CreateSimpleClientWithConfig(t, func() *config.Configuration {
			cfg := config.Config()
			cfg.GetDefaultGrafanaConfig().OutputPath = tc.output
			return cfg
		}, r.Container)
		assert.NoError(t, err)

		// At this point all operations are reading/writing from Minio
		slog.Info("importing Dashboards")
		list := apiClient.DownloadDashboards(dashFilter) // Saving to S3
		assert.Equal(t, len(list), len(boards))
		slog.Info("deleting Dashboards") // Clearing Grafana
		deleteList := apiClient.DeleteAllDashboards(dashFilter)
		assert.Equal(t, len(list), len(deleteList))
		boards = apiClient.ListDashboards(dashFilter)
		assert.Equal(t, len(boards), 0)
		// Load Data from S3
		_, err = apiClient.UploadDashboards(dashFilter) // ReLoad data from S3 backup
		assert.NoError(t, err)

		boards = apiClient.ListDashboards(dashFilter) // Read data
		assert.Equal(t, len(list), len(boards))       // verify

		cancel()
	}
}
