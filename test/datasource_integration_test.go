package test

import (
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service"
	"github.com/esnet/grafana-swagger-api-golang/goclient/models"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestDataSourceCRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	apiClient, _ := initTest(t, nil)
	filtersEntity := service.NewDataSourceFilter("")
	log.Info("Exporting all datasources")
	apiClient.ExportDataSources(filtersEntity)
	log.Info("Listing all datasources")
	dataSources := apiClient.ListDataSources(filtersEntity)
	assert.Equal(t, len(dataSources), 3)
	var dsItem *models.DataSourceListItemDTO
	for _, ds := range dataSources {
		if ds.Name == "netsage" {
			dsItem = &ds
			break
		}
	}
	assert.NotNil(t, dsItem)
	validateDataSource(t, *dsItem)
	//Import Dashboards
	log.Info("Importing datasources")
	list := apiClient.ImportDataSources(filtersEntity)
	assert.Equal(t, len(list), len(dataSources))
	log.Info("Deleting datasources")
	deleteList := apiClient.DeleteAllDataSources(filtersEntity)
	assert.Equal(t, len(deleteList), len(dataSources))
	log.Info("List datasources again")
	dataSources = apiClient.ListDataSources(filtersEntity)
	assert.Equal(t, len(dataSources), 0)
}

// TestDataSourceFilter ensures the regex matching and datasource type filters work as expected
func TestDataSourceFilter(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	initTest(t, nil)

	testingContext := config.Config().Contexts()["testing"]
	testingContext.GetDataSourceSettings().FilterRules = []config.MatchingRule{
		{
			Field: "name",
			Regex: "DEV-*|-Dev-*",
		},
		{
			Field:     "type",
			Inclusive: true,
			Regex:     "elasticsearch|globalnoc-tsds-datasource",
		},
	}
	testingContext = config.Config().Contexts()["testing"]

	apiClient := service.NewApiService("dummy")

	filtersEntity := service.NewDataSourceFilter("")
	log.Info("Exporting all datasources")
	apiClient.ExportDataSources(filtersEntity)
	log.Info("Listing all datasources")
	dataSources := apiClient.ListDataSources(filtersEntity)
	assert.Equal(t, len(dataSources), 2)
	var dsItem *models.DataSourceListItemDTO
	for _, ds := range dataSources {
		if ds.Name == "netsage" {
			dsItem = &ds
			break
		}
	}
	assert.NotNil(t, dsItem)
	validateDataSource(t, *dsItem)
	//Import Dashboards
	log.Info("Importing datasources")
	list := apiClient.ImportDataSources(filtersEntity)
	assert.Equal(t, len(list), len(dataSources))
	log.Info("Deleting datasources")
	deleteList := apiClient.DeleteAllDataSources(filtersEntity)
	assert.Equal(t, len(deleteList), len(dataSources))
	log.Info("List datasources again")
	dataSources = apiClient.ListDataSources(filtersEntity)
	assert.Equal(t, len(dataSources), 0)
}

func validateDataSource(t *testing.T, dsItem models.DataSourceListItemDTO) {
	assert.Equal(t, int64(1), dsItem.OrgID)
	assert.Equal(t, "netsage", dsItem.Name)
	assert.Equal(t, "elasticsearch", dsItem.Type)
	assert.Equal(t, models.DsAccess("proxy"), dsItem.Access)
	assert.Equal(t, "https://netsage-elk1.grnoc.iu.edu/esproxy2/", dsItem.URL)
	assert.True(t, dsItem.BasicAuth)
	assert.True(t, dsItem.IsDefault)

}
