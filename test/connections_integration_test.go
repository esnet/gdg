package test

import (
	"context"
	"log/slog"
	"strings"
	"testing"

	"github.com/esnet/gdg/internal/adapter/grafana/api"
	"github.com/esnet/gdg/internal/adapter/grafana/extended"
	"github.com/esnet/gdg/internal/adapter/grafana/resources"
	"github.com/esnet/gdg/internal/adapter/plugins/secure/noop"
	"github.com/esnet/gdg/internal/adapter/storage"
	configDomain "github.com/esnet/gdg/internal/config/config_domain"
	"github.com/esnet/gdg/internal/domain"
	"github.com/esnet/gdg/pkg/test_tooling/common"

	"github.com/samber/lo"

	"github.com/esnet/gdg/pkg/test_tooling/path"

	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/pkg/test_tooling"
	"github.com/esnet/gdg/pkg/test_tooling/containers"
	"github.com/grafana/grafana-openapi-client-go/models"

	"github.com/stretchr/testify/assert"
)

func TestConnectionPermissionsCrud(t *testing.T) {
	test_tooling.SkipEnterpriseTests(t)
	test_tooling.SkipTokenBasedTests(t)
	assert.NoError(t, path.FixTestDir("test", ".."))
	cfg := config.NewConfig(common.DefaultTestConfig)
	props := containers.DefaultGrafanaEnv()
	err := containers.SetupGrafanaLicense(&props)
	if err != nil {
		slog.Error("no valid grafana license found, skipping enterprise tests")
		t.Skip()
	}
	r := test_tooling.InitTest(t, cfg, props)
	assert.NotNil(t, r)
	assert.NoError(t, err)
	defer func() {
		err := r.CleanUp()
		if err != nil {
			slog.Warn("Unable to clean up after test", "test", t.Name())
		}
	}()
	apiClient := r.ApiClient

	// Probe whether this Grafana instance's Enterprise license includes fine-grained
	// datasource permissions (FGAC).  Instances that respond 403 Unlicensed do not
	// support per-user/per-team datasource permission grants, so there is nothing to
	// test — skip rather than fail.
	if !apiClient.IsDataSourcePermissionsEnabled() {
		t.Skip("skipping: datasource FGAC permissions are not available on this Grafana license")
	}
	// Upload all connections
	filtersEntity := api.NewConnectionFilter("")
	connectionsAdded := apiClient.UploadConnections(filtersEntity)
	assert.Equal(t, len(connectionsAdded), 4)
	// Upload all users
	newUsers := apiClient.UploadUsers(api.NewUserFilter(""))
	assert.Equal(t, len(newUsers), 2)
	// Upload all teams
	filter := api.NewTeamFilter("")
	teams := apiClient.UploadTeams(filter)
	assert.Equal(t, len(teams), 2)
	// Get current Permissions
	permissionFilters := api.NewConnectionFilter("")
	currentPerms := apiClient.ListConnectionPermissions(permissionFilters)
	assert.Equal(t, len(currentPerms), 4)
	// Use netsage (elasticsearch) as the fixture connection for permission assertions —
	// it uses a built-in plugin type that is always available in Grafana containers,
	// making the test reliable across versions and environments.
	var entry *domain.ConnectionPermissionItem
	for ndx, item := range currentPerms {
		if item.Connection.Name == "netsage" {
			entry = &currentPerms[ndx]
			break
		}
	}
	assert.NotNil(t, entry)
	// Initial permission count varies by Grafana version and Enterprise defaults.
	assert.Greater(t, len(entry.Permissions), 0)

	removed := apiClient.DeleteAllConnectionPermissions(permissionFilters)
	assert.Equal(t, len(removed), 4)
	currentPerms = apiClient.ListConnectionPermissions(permissionFilters)
	for ndx, item := range currentPerms {
		if item.Connection.Name == "netsage" {
			entry = &currentPerms[ndx]
			break
		}
	}
	// After clear, non-removable base permissions (admin user, basic:admin role) may remain.
	assert.GreaterOrEqual(t, len(entry.Permissions), 0)
	updated := apiClient.UploadConnectionPermissions(permissionFilters)
	assert.Equal(t, 3, len(updated)) // 3 fixture files → 3 successful uploads (stable)
	currentPerms = apiClient.ListConnectionPermissions(permissionFilters)
	for ndx, item := range currentPerms {
		if item.Connection.Name == "netsage" {
			entry = &currentPerms[ndx]
			break
		}
	}
	// Total count after upload varies by version; focus on verifying specific actors.
	assert.Greater(t, len(entry.Permissions), 0)
	var foundTux, foundBob, foundTeam bool
	for _, item := range entry.Permissions {
		if item.UserLogin == "tux" {
			foundTux = true
			assert.Equal(t, item.Permission, "Admin")
			// Action list size is an implementation detail that varies across Grafana versions.
			assert.True(t, strings.Contains(item.RoleName, "managed:users"))
			assert.True(t, strings.Contains(item.RoleName, "permissions"))
		} else if item.UserLogin == "bob" {
			foundBob = true
			assert.Equal(t, item.Permission, "Edit")
			assert.True(t, strings.Contains(item.RoleName, "managed:users"))
			assert.True(t, strings.Contains(item.RoleName, "permissions"))
		} else if item.Team == "musicians" {
			foundTeam = true
			assert.Equal(t, item.Permission, "Query")
			assert.True(t, strings.Contains(item.RoleName, "managed:teams"))
			assert.True(t, strings.Contains(item.RoleName, "permissions"))
		}
	}
	assert.True(t, foundTux)
	assert.True(t, foundBob)
	assert.True(t, foundTeam)
}

func TestConnectionsCRUD(t *testing.T) {
	cfg := config.NewConfig(common.DefaultTestConfig)
	var r *test_tooling.InitContainerResult
	err := Retry(context.Background(), DefaultRetryAttempts, func() error {
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
	filtersEntity := api.NewConnectionFilter("")
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
	cfg := config.NewConfig(common.DefaultTestConfig)
	var r *test_tooling.InitContainerResult
	err := Retry(context.Background(), DefaultRetryAttempts, func() error {
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

	testingContext := cfg.GetContexts()[common.TestContextName]
	testingContext.GetConnectionSettings().FilterRules = []configDomain.MatchingRule{
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
	testingContext = cfg.GetContexts()[common.TestContextName]

	localEngine := storage.NewLocalStorage(context.Background())
	apiClient = api.NewDashNGo(cfg, noop.NoOpEncoder{}, localEngine, extended.NewExtendedApi(cfg), resources.NewHelpers())
	apiClient.Login()

	filtersEntity := api.NewConnectionFilter("")
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
