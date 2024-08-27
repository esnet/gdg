package test

import (
	"github.com/esnet/gdg/internal/service"
	"github.com/esnet/gdg/pkg/test_tooling"
	"github.com/gosimple/slug"
	"github.com/grafana/grafana-openapi-client-go/models"
	"golang.org/x/exp/slices"
	"os"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrganizationCrud(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	if os.Getenv("TEST_TOKEN_CONFIG") == "1" {
		t.Skip("Skipping Token configuration, Organization CRUD requires Basic SecureData")
	}
	apiClient, _, _, cleanup := test_tooling.InitTest(t, nil, false)
	defer cleanup()
	orgs := apiClient.ListOrganizations(service.NewOrganizationFilter(), true)
	assert.Equal(t, len(orgs), 1)
	mainOrg := orgs[0]
	assert.Equal(t, mainOrg.Organization.ID, int64(1))
	assert.Equal(t, mainOrg.Organization.Name, "Main Org.")
	newOrgs := apiClient.UploadOrganizations(service.NewOrganizationFilter())
	assert.Equal(t, len(newOrgs), 2)
	assert.True(t, slices.Contains(newOrgs, "DumbDumb"))
	assert.True(t, slices.Contains(newOrgs, "Moo"))
	//Filter Org
	orgs = apiClient.ListOrganizations(service.NewOrganizationFilter("DumbDumb"), true)
	assert.Equal(t, len(orgs), 1)
	assert.Equal(t, orgs[0].Organization.Name, "DumbDumb")

}

func TestOrganizationUserMembership(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	if os.Getenv(test_tooling.EnableTokenTestsEnv) == "1" {
		t.Skip("Skipping Token configuration, Organization CRUD requires Basic SecureData")
	}
	apiClient, _, _, cleanup := test_tooling.InitTest(t, nil, false)
	defer cleanup()
	//Create Orgs in case they aren't already present.
	apiClient.UploadOrganizations(service.NewOrganizationFilter())
	orgs := apiClient.ListOrganizations(service.NewOrganizationFilter(), true)
	sort.Slice(orgs, func(a, b int) bool {
		return orgs[a].Organization.ID < orgs[b].Organization.ID
	})
	newOrg := orgs[2]
	//Create Users in case they aren't already present.
	apiClient.UploadUsers(service.NewUserFilter(""))
	// get users
	users := apiClient.ListUsers(service.NewUserFilter(""))
	assert.Equal(t, len(users), 3)
	var orgUser *models.UserSearchHitDTO
	for _, u := range users {
		if u.Login == "tux" {
			orgUser = u
			break
		}
	}
	assert.NotNil(t, orgUser)
	//Reset if any state exists.
	err := apiClient.DeleteUserFromOrg(slug.Make(newOrg.Organization.Name), orgUser.ID)
	assert.Nil(t, err)
	//Start CRUD test
	orgUsers := apiClient.ListOrgUsers(newOrg.Organization.ID)
	assert.Equal(t, len(orgUsers), 1)
	assert.Equal(t, orgUsers[0].Login, "admin")
	assert.Equal(t, orgUsers[0].Role, "Admin")

	err = apiClient.AddUserToOrg("Admin", slug.Make(newOrg.Organization.Name), orgUser.ID)
	assert.Nil(t, err)
	orgUsers = apiClient.ListOrgUsers(newOrg.Organization.ID)
	assert.Equal(t, len(orgUsers), 2)
	assert.Equal(t, orgUsers[1].Role, "Admin")
	err = apiClient.UpdateUserInOrg("Viewer", slug.Make(newOrg.Organization.Name), orgUser.ID)
	orgUsers = apiClient.ListOrgUsers(newOrg.Organization.ID)
	assert.Nil(t, err)
	assert.Equal(t, orgUsers[1].Role, "Viewer")
	err = apiClient.DeleteUserFromOrg(slug.Make(newOrg.Organization.Name), orgUser.ID)
	orgUsers = apiClient.ListOrgUsers(newOrg.Organization.ID)
	assert.Equal(t, len(orgUsers), 1)
	assert.Nil(t, err)
}

func TestOrganizationProperties(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	if os.Getenv(test_tooling.EnableTokenTestsEnv) == "1" {
		t.Skip("Skipping Token configuration, Organization CRUD requires Basic SecureData")
	}
	apiClient, _, _, cleanup := test_tooling.InitTest(t, nil, false)
	defer cleanup()
	apiClient.UploadDashboards(service.NewDashboardFilter("", "", ""))
	defer apiClient.DeleteAllDashboards(service.NewDashboardFilter("", "", ""))
	prefs, err := apiClient.GetOrgPreferences("Main Org.")
	assert.Nil(t, err)
	prefs.HomeDashboardUID = "000000003"
	prefs.Theme = "dark"
	prefs.WeekStart = "Saturday"
	err = apiClient.UploadOrgPreferences("Main Org.", prefs)
	assert.Nil(t, err)
	prefs, err = apiClient.GetOrgPreferences("Main Org.")
	assert.Nil(t, err)
	assert.NotNil(t, prefs)
	assert.Equal(t, prefs.Theme, "dark")
	assert.Equal(t, prefs.WeekStart, "Saturday")
	assert.Equal(t, prefs.HomeDashboardUID, "000000003")
}
