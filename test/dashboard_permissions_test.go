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

// permDashboardCount returns the number of dashboards expected to be returned
// by ListDashboardPermissions for the server version under test.
//
//   - v12: uses listDashboardsV1 (/api/search) → 8 (only General/ v1 files upload)
//   - v13: uses listDashboardsV2 (App Platform API) → 17 (all watched dashboards)
//
// Both paths return one DashboardAndPermissions entry per dashboard found.
func permDashboardCount() int {
	dashCount, _, _ := getDashboardCounts()
	return dashCount
}

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
	// Get current Permissions.
	// permDashboardCount(): 17 on v13 (App Platform listing), 8 on v12 (/api/search).
	permCount := permDashboardCount()
	dashFilter := api.NewDashboardFilter(cfg, "", "", "")
	currentPerms, err := apiClient.ListDashboardPermissions(dashFilter)
	assert.Equal(t, permCount, len(currentPerms))
	entry := new(lo.FirstOrEmpty(lo.Filter(currentPerms, func(item domain.DashboardAndPermissions, index int) bool {
		return item.Dashboard.Title == "Bandwidth Dashboard"
	})))
	assert.NotNil(t, entry)
	// Default permission count varies by version (v12: 3 ACL entries, v13: varies with RBAC).
	assert.Greater(t, len(entry.Permissions), 0)

	assert.NoError(t, apiClient.ClearDashboardPermissions(dashFilter))
	currentPerms, err = apiClient.ListDashboardPermissions(dashFilter)
	assert.NoError(t, err)
	assert.Equal(t, permCount, len(currentPerms))
	// After clear, managed permissions should be removed. Non-managed (inherited) permissions
	// may remain on v13 RBAC; the key check is that bob and musicians are gone.
	if isV13() {
		// Find Bandwidth Dashboard entry and confirm no bob or musicians permission.
		bwEntry := lo.FirstOrEmpty(lo.Filter(currentPerms, func(item domain.DashboardAndPermissions, _ int) bool {
			return item.Dashboard.Title == "Bandwidth Dashboard"
		}))
		bobGone := lo.NoneBy(bwEntry.Permissions, func(p *models.ResourcePermissionDTO) bool { return p.UserLogin == "bob" })
		musiciansGone := lo.NoneBy(bwEntry.Permissions, func(p *models.ResourcePermissionDTO) bool { return p.Team == "musicians" })
		assert.True(t, bobGone, "bob's permission should be cleared")
		assert.True(t, musiciansGone, "musicians team permission should be cleared")
	} else {
		assert.Equal(t, 0, len(currentPerms[0].Permissions))
	}
	addPerms, err := apiClient.UploadDashboardPermissions(dashFilter)
	assert.NoError(t, err)
	assert.Equal(t, permCount, len(addPerms))
	currentPerms, err = apiClient.ListDashboardPermissions(dashFilter)
	entry = nil
	entry = new(lo.FirstOrEmpty(lo.Filter(currentPerms, func(item domain.DashboardAndPermissions, index int) bool {
		return item.Dashboard.Title == "Bandwidth Dashboard"
	})))
	assert.NotNil(t, entry)
	// Permission count varies by version: v12 ACL returns 5 flat entries; v13 RBAC
	// may return a different count depending on inherited and managed permissions.
	assert.Greater(t, len(entry.Permissions), 0)
	var bobPerm *models.ResourcePermissionDTO
	var teamMusic *models.ResourcePermissionDTO
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
	// Validate using ResourcePermissionDTO fields (unified type for v12 and v13).
	// v12: DashboardACLInfoDTO is mapped to ResourcePermissionDTO before returning,
	//      so Permission is the string "Edit"/"Admin" (from PermissionName),
	//      BuiltInRole is the role string (from Role), UserID and TeamID are preserved.
	assert.Equal(t, "Edit", bobPerm.Permission)
	assert.Equal(t, int64(2), bobPerm.UserID)
	assert.Equal(t, "Admin", teamMusic.Permission)
	assert.Equal(t, int64(2), teamMusic.TeamID)
}
