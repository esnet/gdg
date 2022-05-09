package integration_tests

import (
	"github.com/esnet/gdg/config"
	"testing"

	"github.com/esnet/gdg/api"
	"github.com/grafana-tools/sdk"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestDataSourceCRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	apiClient, _ := initTest(t)
	filters := api.NewDashboardFilter()
	log.Info("Exporting all datasources")
	apiClient.ExportDataSources(filters)
	log.Info("Listing all datasources")
	dataSources := apiClient.ListDataSources(filters)
	assert.Equal(t, len(dataSources), 3)
	var dsItem *sdk.Datasource
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
	list := apiClient.ImportDataSources(filters)
	assert.Equal(t, len(list), len(dataSources))
	log.Info("Deleting datasources")
	deleteList := apiClient.DeleteAllDataSources(filters)
	assert.Equal(t, len(deleteList), len(dataSources))
	log.Info("List datasources again")
	dataSources = apiClient.ListDataSources(filters)
	assert.Equal(t, len(dataSources), 0)
}

//TestDataSourceFilter ensures the regex matching and datasource type filters work as expected
func TestDataSourceFilter(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	initTest(t)

	testingContext := config.Config().Contexts()["testing"]
	testingContext.GetDataSourceSettings().Filters = &config.DataSourceFilters{
		NameExclusions:  "DEV-*|-Dev-*",
		DataSourceTypes: []string{"elasticsearch", "globalnoc-tsds-datasource"},
	}
	testingContext = config.Config().Contexts()["testing"]
	log.Info(testingContext.GetDataSourceSettings().Filters)

	apiClient := api.NewApiService("dummy")

	filters := api.NewDashboardFilter()
	log.Info("Exporting all datasources")
	apiClient.ExportDataSources(filters)
	log.Info("Listing all datasources")
	dataSources := apiClient.ListDataSources(filters)
	assert.Equal(t, len(dataSources), 2)
	var dsItem *sdk.Datasource
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
	list := apiClient.ImportDataSources(filters)
	assert.Equal(t, len(list), len(dataSources))
	log.Info("Deleting datasources")
	deleteList := apiClient.DeleteAllDataSources(filters)
	assert.Equal(t, len(deleteList), len(dataSources))
	log.Info("List datasources again")
	dataSources = apiClient.ListDataSources(filters)
	assert.Equal(t, len(dataSources), 0)
}

func validateDataSource(t *testing.T, dsItem sdk.Datasource) {
	assert.Equal(t, uint(1), dsItem.OrgID)
	assert.Equal(t, "netsage", dsItem.Name)
	assert.Equal(t, "elasticsearch", dsItem.Type)
	assert.Equal(t, "proxy", dsItem.Access)
	assert.Equal(t, "https://netsage-elk1.grnoc.iu.edu/esproxy2/", dsItem.URL)
	assert.True(t, *dsItem.BasicAuth)
	assert.True(t, dsItem.IsDefault)

}
