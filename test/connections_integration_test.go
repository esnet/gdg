package test

import (
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service"
	"github.com/esnet/grafana-swagger-api-golang/goclient/models"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConnectionsCRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	apiClient, _, cleanup := initTest(t, nil)
	defer cleanup()
	filtersEntity := service.NewConnectionFilter("")
	slog.Info("Exporting all connections")
	apiClient.UploadConnections(filtersEntity)
	slog.Info("Listing all connections")
	dataSources := apiClient.ListConnections(filtersEntity)
	assert.Equal(t, len(dataSources), 3)
	var dsItem *models.DataSourceListItemDTO
	for _, ds := range dataSources {
		if ds.Name == "netsage" {
			dsItem = &ds
			break
		}
	}
	assert.NotNil(t, dsItem)
	validateConnection(t, *dsItem)
	//Import Dashboards
	slog.Info("Importing connections")
	list := apiClient.DownloadConnections(filtersEntity)
	assert.Equal(t, len(list), len(dataSources))
	slog.Info("Deleting connections")
	deleteList := apiClient.DeleteAllConnections(filtersEntity)
	assert.Equal(t, len(deleteList), len(dataSources))
	slog.Info("List connections again")
	dataSources = apiClient.ListConnections(filtersEntity)
	assert.Equal(t, len(dataSources), 0)
}

// TestConnectionFilter ensures the regex matching and datasource type filters work as expected
func TestConnectionFilter(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	_, _, clean := initTest(t, nil)
	defer clean()

	testingContext := config.Config().GetAppConfig().GetContexts()["testing"]
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
	testingContext = config.Config().GetAppConfig().GetContexts()["testing"]
	_ = testingContext

	apiClient := service.NewApiService("dummy")

	filtersEntity := service.NewConnectionFilter("")
	slog.Info("Exporting all connections")
	apiClient.UploadConnections(filtersEntity)
	slog.Info("Listing all connections")
	dataSources := apiClient.ListConnections(filtersEntity)
	assert.Equal(t, len(dataSources), 2)
	var dsItem *models.DataSourceListItemDTO
	for _, ds := range dataSources {
		if ds.Name == "netsage" {
			dsItem = &ds
			break
		}
	}
	assert.NotNil(t, dsItem)
	validateConnection(t, *dsItem)
	//Import Dashboards
	slog.Info("Importing connections")
	list := apiClient.DownloadConnections(filtersEntity)
	assert.Equal(t, len(list), len(dataSources))
	slog.Info("Deleting connections")
	deleteList := apiClient.DeleteAllConnections(filtersEntity)
	assert.Equal(t, len(deleteList), len(dataSources))
	slog.Info("List connections again")
	dataSources = apiClient.ListConnections(filtersEntity)
	assert.Equal(t, len(dataSources), 0)
}

func validateConnection(t *testing.T, dsItem models.DataSourceListItemDTO) {
	assert.Equal(t, int64(1), dsItem.OrgID)
	assert.Equal(t, "netsage", dsItem.Name)
	assert.Equal(t, "elasticsearch", dsItem.Type)
	assert.Equal(t, models.DsAccess("proxy"), dsItem.Access)
	assert.Equal(t, "https://netsage-elk1.grnoc.iu.edu/esproxy2/", dsItem.URL)
	assert.True(t, dsItem.BasicAuth)
	assert.True(t, dsItem.IsDefault)

}
