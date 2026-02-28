package test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/esnet/gdg/internal/adapter/grafana/api"
	"github.com/esnet/gdg/internal/adapter/plugins/secure/noop"
	"github.com/esnet/gdg/internal/config/config_domain"
	"github.com/esnet/gdg/internal/ports"
	"github.com/esnet/gdg/pkg/test_tooling/common"

	"github.com/esnet/gdg/internal/config"

	"github.com/esnet/gdg/pkg/test_tooling"
	"github.com/esnet/gdg/pkg/test_tooling/path"
	"github.com/stretchr/testify/assert"
	_ "gocloud.dev/blob/memblob"
)

func TestCloudDataSourceCRUD(t *testing.T) {
	t.Log("Running Cloud Tests")
	assert.NoError(t, path.FixTestDir("test", ".."))
	assert.NoError(t, os.Setenv(common.ContextNameEnv, common.TestContextName))
	assert.NoError(t, os.Unsetenv(common.ContextNameEnv))

	cfg := config.NewConfig(common.DefaultTestConfig)
	var r *test_tooling.InitContainerResult
	err := Retry(context.Background(), DefaultRetryAttempts, func() error {
		r = test_tooling.InitTest(t, cfg, nil)
		return r.Err
	})
	assert.NotNil(t, r)
	assert.NoError(t, err)
	defer func() {
		errCleanup := r.CleanUp()
		if errCleanup != nil {
			slog.Warn("Unable to clean up after test", "test", t.Name())
		}
	}()
	apiClient := r.ApiClient
	// Wipe all data from grafana
	dsFilter := api.NewConnectionFilter("")
	apiClient.DeleteAllConnections(dsFilter)

	apiClient.UploadConnections(dsFilter)
	dsList := apiClient.ListConnections(dsFilter)
	assert.True(t, len(dsList) > 0)
	_, cancel, apiClient, s, err := test_tooling.SetupCloudFunctionOpt(
		cfg,
		noop.NoOpEncoder{},
		test_tooling.SetCloudType("custom"),
		test_tooling.SetBucketName(common.TestBucketName))
	apiClient.(*api.DashNGoImpl).SetStorage(s)
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
	assert.NoError(t, os.Setenv(common.ContextNameEnv, common.TestContextName))
	assert.NoError(t, os.Unsetenv(common.ContextNameEnv))
	assert.NoError(t, path.FixTestDir("test", ".."))
	var (
		err error
		s   ports.Storage
	)
	cfg := config.NewConfig(common.DefaultTestConfig)
	var r *test_tooling.InitContainerResult
	err = Retry(context.Background(), DefaultRetryAttempts, func() error {
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
	// defer cleanup, "Failed to cleanup test containers for %s", t.Name())
	// Wipe all data from grafana
	dashFilter := api.NewDashboardFilter(cfg, "", "", "")
	apiClient.DeleteAllDashboards(dashFilter)
	// Load data into grafana
	apiClient.UploadDashboards(dashFilter)
	boards := apiClient.ListDashboards(dashFilter)
	assert.True(t, len(boards) > 0)
	var cancel context.CancelFunc

	_, cancel, apiClient, s, err = test_tooling.SetupCloudFunctionOpt(
		cfg,
		noop.NoOpEncoder{},
		test_tooling.SetCloudType("custom"),
		test_tooling.SetBucketName(common.TestBucketName))
	assert.NoError(t, err)
	apiClient.(*api.DashNGoImpl).SetStorage(s)
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
	_, err = apiClient.UploadDashboards(dashFilter) // ReLoad data from S3 backup
	assert.NoError(t, err)
	boards = apiClient.ListDashboards(dashFilter) // Read data
	assert.Equal(t, len(list), len(boards))       // verify
	apiClient.DeleteAllDashboards(dashFilter)
}

func TestDashboardCloudLeadingSlashCRUD(t *testing.T) {
	var (
		err    error
		cancel context.CancelFunc
		r      *test_tooling.InitContainerResult
	)
	assert.NoError(t, os.Setenv(common.ContextNameEnv, common.TestContextName))
	assert.NoError(t, os.Unsetenv(common.ContextNameEnv))
	assert.NoError(t, path.FixTestDir("test", ".."))

	cfg := config.NewConfig(common.DefaultTestConfig)
	r = test_tooling.InitTest(t, cfg, nil)
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
	dashFilter := api.NewDashboardFilter(cfg, "", "", "")
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
		useEnv   bool
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
			name:   "/prefix and no slash output",
			prefix: "/dummy",
			output: "test/data",
		},
		// The use case below will not be able to use secure auth location since the output path will be invalid.
		// The path need to be a location that can be found on the file system. The fallback is to rely on ENV settings
		// to configure the auth.
		{
			useEnv: true,
			name:   "no prefix, slash output",
			prefix: "",
			output: "/test/data",
		},
		{
			useEnv: true,
			name:   "/prefix and no output",
			prefix: "/dummy",
			output: "",
		},
		{
			useEnv: true,
			name:   "/prefix and /output",
			prefix: "/dummy",
			output: "/test/data",
		},
	}

	for _, tc := range testcases {
		if tc.disabled {
			t.Log("Skipping disabled test case", tc.name)
			continue
		}
		t.Log("Running test", tc.name)
		var storageSvc ports.Storage
		slog.Warn("Running testcase", "name", tc.name)
		if tc.useEnv {
			os.Setenv("AWS_ACCESS_KEY", "test")
			os.Setenv("AWS_SECRET_KEY", "secretsss")
		}
		_, cancel, apiClient, storageSvc, err = test_tooling.SetupCloudFunctionOpt(
			cfg,
			noop.NoOpEncoder{},
			test_tooling.SetCloudType("custom"),
			test_tooling.SetPrefix(tc.prefix),
			test_tooling.SetAccessKey("test"),
			test_tooling.SetSecretKey("secretsss"),
			test_tooling.SetBucketName(common.TestBucketName))

		grafanaConfig := cfg.GetDefaultGrafanaConfig()
		grafanaConfig.OutputPath = tc.output
		apiClient = test_tooling.CreateSimpleClientWithConfig(t, cfg, r.Container)
		apiClient.(*api.DashNGoImpl).SetStorage(storageSvc) // Override storage
		assert.NoError(t, err)

		// At this point all operations are reading/writing from Minio
		slog.Info("Downloading Dashboards")
		list := apiClient.DownloadDashboards(dashFilter) // Saving to S3
		assert.Equal(t, len(list), len(boards))
		slog.Info("Deleting Dashboards") // Clearing Grafana
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

// MaintainConfigAuth updates Grafana config auth, preserving existing secure data during init.
func MaintainConfigAuth(cfg *config_domain.GDGAppConfiguration, configVal string) {
	cfg.GetDefaultGrafanaConfig().Apply()
	var auth *config_domain.SecureModel
	auth = cfg.GetDefaultGrafanaConfig().TestGetSecureAuth()
	cfg = config.NewConfig(configVal)
	if auth != nil {
		cfg.GetDefaultGrafanaConfig().Apply(config_domain.WithSecureAuth(*auth))
	}
}
