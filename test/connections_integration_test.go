package test

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/esnet/gdg/pkg/test_tooling/common"

	"github.com/samber/lo"

	"github.com/esnet/gdg/pkg/test_tooling/path"

	"github.com/esnet/gdg/internal/storage"

	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service"
	"github.com/esnet/gdg/internal/types"
	"github.com/esnet/gdg/pkg/test_tooling"
	"github.com/esnet/gdg/pkg/test_tooling/containers"
	"github.com/safaci2000/grafana-openapi-client-go/models"

	"github.com/stretchr/testify/assert"
)

// There's some issues with these tests, temporarily disabling this
func TestConnectionPermissionsCrud(t *testing.T) {
	t.Skip() // Buggy test right now, disabled
	assert.NoError(t, path.FixTestDir("test", ".."))
	if os.Getenv(test_tooling.EnableTokenTestsEnv) == test_tooling.FeatureEnabled {
		t.Skip("Skipping Token configuration, Team and User CRUD requires Basic SecureData")
	}
	props := containers.DefaultGrafanaEnv()
	err := containers.SetupGrafanaLicense(&props)
	if err != nil {
		slog.Error("no valid grafana license found, skipping enterprise tests")
		t.Skip()
	}
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
	// Upload all connections
	filtersEntity := service.NewConnectionFilter("")
	connectionsAdded := apiClient.UploadConnections(filtersEntity)
	assert.Equal(t, len(connectionsAdded), 3)
	// Upload all users
	newUsers := apiClient.UploadUsers(service.NewUserFilter(""))
	assert.Equal(t, len(newUsers), 2)
	// Upload all teams
	filter := service.NewTeamFilter("")
	teams := apiClient.UploadTeams(filter)
	assert.Equal(t, len(teams), 2)
	// Get current Permissions
	permissionFilters := service.NewConnectionFilter("")
	currentPerms := apiClient.ListConnectionPermissions(permissionFilters)
	assert.Equal(t, len(currentPerms), 3)
	var entry *types.ConnectionPermissionItem
	for ndx, item := range currentPerms {
		if item.Connection.Name == "Google Sheets" {
			entry = &currentPerms[ndx]
			break
		}
	}
	assert.NotNil(t, entry)
	assert.Equal(t, len(entry.Permissions), 4)

	removed := apiClient.DeleteAllConnectionPermissions(permissionFilters)
	assert.Equal(t, len(removed), 3)
	currentPerms = apiClient.ListConnectionPermissions(permissionFilters)
	for ndx, item := range currentPerms {
		if item.Connection.Name == "Google Sheets" {
			entry = &currentPerms[ndx]
			break
		}
	}
	assert.Equal(t, 2, len(entry.Permissions))
	updated := apiClient.UploadConnectionPermissions(permissionFilters)
	assert.Equal(t, 3, len(updated))
	currentPerms = apiClient.ListConnectionPermissions(permissionFilters)
	for ndx, item := range currentPerms {
		if item.Connection.Name == "Google Sheets" {
			entry = &currentPerms[ndx]
			break
		}
	}
	assert.Equal(t, len(entry.Permissions), 7)
	currentPerms = apiClient.ListConnectionPermissions(permissionFilters)
	var foundTux, foundBob, foundTeam bool
	for _, item := range entry.Permissions {
		if item.UserLogin == "tux" {
			foundTux = true
			assert.Equal(t, item.Permission, "Admin")
			assert.Equal(t, len(item.Actions), 8)
			assert.True(t, strings.Contains(item.RoleName, "managed:users"))
			assert.True(t, strings.Contains(item.RoleName, "permissions"))
		} else if item.UserLogin == "bob" {
			foundBob = true
			assert.Equal(t, item.Permission, "Edit")
			assert.Equal(t, len(item.Actions), 4)
			assert.True(t, strings.Contains(item.RoleName, "managed:users"))
			assert.True(t, strings.Contains(item.RoleName, "permissions"))
		} else if item.Team == "musicians" {
			foundTeam = true
			assert.Equal(t, item.Permission, "Query")
			assert.Equal(t, len(item.Actions), 2)
			assert.True(t, strings.Contains(item.RoleName, "managed:teams"))
			assert.True(t, strings.Contains(item.RoleName, "permissions"))
		}
	}
	assert.True(t, foundTux)
	assert.True(t, foundBob)
	assert.True(t, foundTeam)
}

func TestConnectionsCRUD(t *testing.T) {
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
	filtersEntity := service.NewConnectionFilter("")
	slog.Info("Exporting all connections")
	apiClient.UploadConnections(filtersEntity)
	slog.Info("Listing all connections")
	dataSources := apiClient.ListConnections(filtersEntity)
	assert.Equal(t, len(dataSources), 4)
	dsItem := lo.FirstOrEmpty(lo.Filter(dataSources, func(item models.DataSourceListItemDTO, index int) bool {
		return item.Name == "netsage"
	}))
	assert.NotNil(t, dsItem)
	validateConnection(t, dsItem)
	// Import Dashboards
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
	assert.NoError(t, path.FixTestDir("test", ".."))
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

	testingContext := config.Config().GetGDGConfig().GetContexts()[common.TestContextName]
	testingContext.GetConnectionSettings().FilterRules = []config.MatchingRule{
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
	testingContext = config.Config().GetGDGConfig().GetContexts()[common.TestContextName]

	localEngine := storage.NewLocalStorage(context.Background())
	apiClient = service.NewTestApiService(localEngine, func() *config.Configuration {
		return config.Config()
	})

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
	// Import Dashboards
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
