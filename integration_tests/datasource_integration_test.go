package integration_tests

import (
	"testing"

	"github.com/grafana-tools/sdk"
	"github.com/netsage-project/gdg/api"
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
	assert.Equal(t, dsItem.OrgID, uint(1))
	assert.Equal(t, dsItem.Name, "netsage")
	assert.Equal(t, dsItem.Type, "elasticsearch")
	assert.Equal(t, dsItem.Access, "proxy")
	assert.Equal(t, dsItem.URL, "https://netsage-elk1.grnoc.iu.edu/esproxy2/")
	assert.True(t, *dsItem.BasicAuth)
	assert.True(t, dsItem.IsDefault)

}
