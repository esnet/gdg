package test

import (
	"log/slog"
	"testing"

	"github.com/esnet/gdg/internal/adapter/grafana/api"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/domain"
	"github.com/esnet/gdg/pkg/test_tooling/common"

	"github.com/esnet/gdg/pkg/test_tooling"
	"github.com/esnet/gdg/pkg/test_tooling/containers"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestDashboardPermissionsCrud(t *testing.T) {
	test_tooling.SkipTokenBasedTests(t)
	test_tooling.SkipEnterpriseTests(t)
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
		cleanUpErr := r.CleanUp()
		if cleanUpErr != nil {
			slog.Warn("Unable to clean up after test", "test", t.Name())
		}
	}()
	apiClient := r.ApiClient
	// Upload all dashboards
	_, err = apiClient.UploadDashboards(api.NewDashboardFilter(cfg, "", "", ""))
	assert.NoError(t, err)
	// Upload all users
	newUsers := apiClient.UploadUsers(api.NewUserFilter(""))
	assert.Equal(t, len(newUsers), 2)
	// Upload all teams
	filter := api.NewTeamFilter("")
	teams := apiClient.UploadTeams(filter)
	assert.Equal(t, len(teams), 2)
	// Get current Permissions
	dashFilter := api.NewDashboardFilter(cfg, "", "", "")
	currentPerms, err := apiClient.ListDashboardPermissions(dashFilter)
	assert.Equal(t, len(currentPerms), DashboardCount)
	entry := new(lo.FirstOrEmpty(lo.Filter(currentPerms, func(item domain.DashboardAndPermissions, index int) bool {
		return item.Dashboard.Title == "Bandwidth Dashboard"
	})))
	assert.NotNil(t, entry)
	assert.Equal(t, len(entry.Permissions), 3)

	assert.NoError(t, apiClient.ClearDashboardPermissions(dashFilter))
	currentPerms, err = apiClient.ListDashboardPermissions(dashFilter)
	assert.NoError(t, err)
	assert.Equal(t, len(currentPerms), DashboardCount)
	assert.Equal(t, len(currentPerms[0].Permissions), 0)
	addPerms, err := apiClient.UploadDashboardPermissions(dashFilter)
	assert.NoError(t, err)
	assert.Equal(t, len(addPerms), DashboardCount)
	currentPerms, err = apiClient.ListDashboardPermissions(dashFilter)
	entry = nil
	entry = new(lo.FirstOrEmpty(lo.Filter(currentPerms, func(item domain.DashboardAndPermissions, index int) bool {
		return item.Dashboard.Title == "Bandwidth Dashboard"
	})))
	assert.NotNil(t, entry)
	assert.Equal(t, 5, len(entry.Permissions))
	var bobPerm *models.DashboardACLInfoDTO
	var teamMusic *models.DashboardACLInfoDTO
	for ndx, entryPerm := range entry.Permissions {
		if entryPerm.Team == "musicians" {
			teamMusic = entry.Permissions[ndx]
		}
		if entryPerm.UserLogin == "bob" {
			bobPerm = entry.Permissions[ndx]
		}
	}
	assert.NotNil(t, bobPerm)
	assert.NotNil(t, teamMusic)
	// validate bob
	assert.Equal(t, bobPerm.PermissionName, "Edit")
	assert.Equal(t, bobPerm.UserEmail, "bob@aol.com")
	assert.Equal(t, bobPerm.UserID, int64(2))
	assert.Equal(t, bobPerm.Permission, models.PermissionType(2))
	// validate team permission
	assert.Equal(t, teamMusic.PermissionName, "Admin")
	assert.Equal(t, teamMusic.TeamID, int64(2))
	assert.Equal(t, teamMusic.Permission, models.PermissionType(4))
}
