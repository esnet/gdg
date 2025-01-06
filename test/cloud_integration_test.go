package test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/esnet/gdg/internal/service"
	"github.com/esnet/gdg/pkg/test_tooling"
	"github.com/esnet/gdg/pkg/test_tooling/path"
	"github.com/stretchr/testify/assert"
	_ "gocloud.dev/blob/memblob"
)

func TestCloudDataSourceCRUD(t *testing.T) {
	assert.NoError(t, path.FixTestDir("test", ".."))
	t.Log("Running Cloud Tests")
	apiClient, _, _, cleanup := test_tooling.InitTestLegacy(t, nil, nil)
	defer func() {
		cleanErr := cleanup()
		if cleanErr != nil {
			slog.Error("unable to clean up after test", slog.Any("err", cleanErr))
		}
	}()
	// Wipe all data from grafana
	dsFilter := service.NewConnectionFilter("")
	apiClient.DeleteAllConnections(dsFilter)

	apiClient.UploadConnections(dsFilter)
	dsList := apiClient.ListConnections(dsFilter)
	assert.True(t, len(dsList) > 0)
	_, cancel, apiClient, err := test_tooling.SetupCloudFunction([]string{"s3", "testing"})
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
	assert.NoError(t, os.Setenv("GDG_CONTEXT_NAME", "testing"))
	assert.NoError(t, path.FixTestDir("test", ".."))
	var err error
	apiClient, _, _, cleanup := test_tooling.InitTestLegacy(t, nil, nil)
	defer cleanup()
	// defer cleanup, "Failed to cleanup test containers for %s", t.Name())
	// Wipe all data from grafana
	dashFilter := service.NewDashboardFilter("", "", "")
	apiClient.DeleteAllDashboards(dashFilter)
	// Load data into grafana
	apiClient.UploadDashboards(dashFilter)
	boards := apiClient.ListDashboards(dashFilter)
	assert.True(t, len(boards) > 0)
	var cancel context.CancelFunc

	_, cancel, apiClient, err = test_tooling.SetupCloudFunction([]string{"s3", "testing"})
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
