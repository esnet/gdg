package test

import (
	"github.com/esnet/gdg/internal/service"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	_ "gocloud.dev/blob/memblob"
	"os"
	"testing"
)

func TestCloudDataSourceCRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	os.Setenv("GDG_CONTEXT_NAME", "testing")
	apiClient, cfg := initTest(t, nil)

	//Wipe all data from grafana
	dsFilter := service.NewConnectionFilter("")
	apiClient.DeleteAllConnections(dsFilter)

	apiClient.UploadConnections(dsFilter)
	dsList := apiClient.ListConnections(dsFilter)
	assert.True(t, len(dsList) > 0)
	SetupCloudFunction(t, cfg, apiClient, []string{"minio", "testing"})
	//SetupCloudFunction(apiClient, []string{"mem", "testing"})

	log.Info("Importing DataSources")
	dsStringList := apiClient.DownloadConnections(dsFilter) //Saving to S3
	assert.Equal(t, len(dsList), len(dsStringList))
	log.Info("Deleting DataSources")
	deleteDSList := apiClient.DeleteAllConnections(dsFilter) // Cleaning up Grafana
	assert.Equal(t, len(deleteDSList), len(dsStringList))
	dsList = apiClient.ListConnections(dsFilter)
	assert.Equal(t, len(dsList), 0)
	//Load Data from S3
	apiClient.UploadConnections(dsFilter) //Load data from S3
	dsList = apiClient.ListConnections(dsFilter)
	assert.Equal(t, len(dsList), len(dsStringList))
	apiClient.DeleteAllConnections(dsFilter) // Cleaning up Grafana
}

// TestDashboardCloudCrud will load testing_data to Grafana from local context.  Switch to CLoud,
// Save all data to Cloud, wipe grafana and reload data back into grafana and validate
func TestDashboardCloudCRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	os.Setenv("GDG_CONTEXT_NAME", "testing")
	apiClient, cfg := initTest(t, nil)

	//Wipe all data from grafana
	dashFilter := service.NewDashboardFilter("", "", "")
	apiClient.DeleteAllDashboards(dashFilter)
	//Load data into grafana
	apiClient.UploadDashboards(dashFilter)
	boards := apiClient.ListDashboards(dashFilter)
	assert.True(t, len(boards) > 0)

	//SetupCloudFunction(apiClient, []string{"mem", "testing"})
	_, apiClient = SetupCloudFunction(t, cfg, apiClient, []string{"minio", "testing"})

	//At this point all operations are reading/writing from Minio
	log.Info("Importing Dashboards")
	list := apiClient.DownloadDashboards(dashFilter) //Saving to S3
	assert.Equal(t, len(list), len(boards))
	log.Info("Deleting Dashboards") // Clearing Grafana
	deleteList := apiClient.DeleteAllDashboards(dashFilter)
	assert.Equal(t, len(list), len(deleteList))
	boards = apiClient.ListDashboards(dashFilter)
	assert.Equal(t, len(boards), 0)
	//Load Data from S3
	apiClient.UploadDashboards(dashFilter)        //ReLoad data from S3 backup
	boards = apiClient.ListDashboards(dashFilter) //Read data
	assert.Equal(t, len(list), len(boards))       //verify
	apiClient.DeleteAllDashboards(dashFilter)
}
