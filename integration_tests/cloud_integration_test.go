package integration_tests

import (
	"github.com/esnet/gdg/api"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

//TestDashboardCloudCrud will load testing_data to Grafana from local context.  Switch to CLoud,
// Save all data to Cloud, wipe grafana and reload data back into grafana and validate
func TestDashboardS3CRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	apiClient, _ := initTest(t)

	//Wipe all data from grafana
	dsFilter := api.DatasourceFilter{}
	dashFilter := api.NewDashboardFilter()
	apiClient.DeleteAllDashboards(dashFilter)
	apiClient.DeleteAllDataSources(dsFilter)
	//Load data into grafana
	apiClient.ExportDashboards(dashFilter)
	boards := apiClient.ListDashboards(dashFilter)
	assert.True(t, len(boards) > 0)
	//
	apiClient.ExportDataSources(dsFilter)
	dsList := apiClient.ListDataSources(dsFilter)
	assert.True(t, len(dsList) > 0)
	//Switch to Cloud version
	SetupCloudFunction(apiClient, "s3")
	//At this point all operations are reading/writing from Minio
	log.Info("Importing Dashboards")
	list := apiClient.ImportDashboards(dashFilter) //Saving to S3
	assert.Equal(t, len(list), len(boards))
	log.Info("Deleting Dashboards") // Clearing Grafana
	deleteList := apiClient.DeleteAllDashboards(dashFilter)
	assert.Equal(t, len(list), len(deleteList))
	boards = apiClient.ListDashboards(dashFilter)
	assert.Equal(t, len(boards), 0)
	//Load Data from S3
	apiClient.ExportDashboards(dashFilter)        //ReLoad data from S3 backup
	boards = apiClient.ListDashboards(dashFilter) //Read data
	assert.Equal(t, len(list), len(boards))       //verify
	apiClient.DeleteAllDashboards(dashFilter)
	//
	log.Info("Importing DataSources")
	dsStringList := apiClient.ImportDataSources(dsFilter) //Saving to S3
	assert.Equal(t, len(dsList), len(dsStringList))
	log.Info("Deleting DataSources")
	deleteDSList := apiClient.DeleteAllDataSources(dsFilter) // Cleaning up Grafana
	assert.Equal(t, len(deleteDSList), len(dsStringList))
	dsList = apiClient.ListDataSources(dsFilter)
	assert.Equal(t, len(dsList), 0)
	//Load Data from S3
	apiClient.ExportDataSources(dsFilter) //Load data from S3
	dsList = apiClient.ListDataSources(dsFilter)
	assert.Equal(t, len(dsList), len(dsStringList))
	apiClient.DeleteAllDataSources(dsFilter) // Cleaning up Grafana
}

func TestDashboardSFTPCRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	apiClient, _ := initTest(t)

	//Wipe all data from grafana
	dsFilter := api.DatasourceFilter{}
	dashFilter := api.NewDashboardFilter()
	apiClient.DeleteAllDashboards(dashFilter)
	apiClient.DeleteAllDataSources(dsFilter)
	//Load data into grafana
	apiClient.ExportDashboards(dashFilter)
	boards := apiClient.ListDashboards(dashFilter)
	assert.True(t, len(boards) > 0)
	//
	apiClient.ExportDataSources(dsFilter)
	dsList := apiClient.ListDataSources(dsFilter)
	assert.True(t, len(dsList) > 0)
	//Switch to Cloud version
	SetupCloudFunction(apiClient, "sftp")
	//At this point all operations are reading/writing from Minio
	log.Info("Importing Dashboards")
	list := apiClient.ImportDashboards(dashFilter) //Saving to S3
	assert.Equal(t, len(list), len(boards))
	log.Info("Deleting Dashboards") // Clearing Grafana
	deleteList := apiClient.DeleteAllDashboards(dashFilter)
	assert.Equal(t, len(list), len(deleteList))
	boards = apiClient.ListDashboards(dashFilter)
	assert.Equal(t, len(boards), 0)
	//Load Data from S3
	apiClient.ExportDashboards(dashFilter)        //ReLoad data from S3 backup
	boards = apiClient.ListDashboards(dashFilter) //Read data
	assert.Equal(t, len(list), len(boards))       //verify
	apiClient.DeleteAllDashboards(dashFilter)
	//
	log.Info("Importing DataSources")
	dsStringList := apiClient.ImportDataSources(dsFilter) //Saving to S3
	assert.Equal(t, len(dsList), len(dsStringList))
	log.Info("Deleting DataSources")
	deleteDSList := apiClient.DeleteAllDataSources(dsFilter) // Cleaning up Grafana
	assert.Equal(t, len(deleteDSList), len(dsStringList))
	dsList = apiClient.ListDataSources(dsFilter)
	assert.Equal(t, len(dsList), 0)
	//Load Data from S3
	apiClient.ExportDataSources(dsFilter) //Load data from S3
	dsList = apiClient.ListDataSources(dsFilter)
	assert.Equal(t, len(dsList), len(dsStringList))
	apiClient.DeleteAllDataSources(dsFilter) // Cleaning up Grafana
}
