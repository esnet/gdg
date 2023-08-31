package test

import (
	"github.com/esnet/gdg/internal/service"
	"github.com/esnet/grafana-swagger-api-golang/goclient/models"
	"golang.org/x/exp/slices"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrgsCrud(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	apiClient, _ := initTest(t, nil)
	orgs := apiClient.ListOrganizations()
	assert.Equal(t, len(orgs), 1)
	mainOrg := orgs[0]
	assert.Equal(t, mainOrg.ID, int64(1))
	assert.Equal(t, mainOrg.Name, "Main Org.")
	newOrgs := apiClient.UploadOrganizations()
	assert.Equal(t, len(newOrgs), 2)
	assert.True(t, slices.Contains(newOrgs, "DumbDumb"))
	assert.True(t, slices.Contains(newOrgs, "Moo"))

}

func TestOrgUserMembership(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	apiClient, _ := initTest(t, nil)
	//Create Orgs in case they aren't already present.
	apiClient.UploadOrganizations()
	orgs := apiClient.ListOrganizations()
	sort.Slice(orgs, func(a, b int) bool {
		return orgs[a].ID < orgs[b].ID
	})
	newOrg := orgs[2]
	//Create Users in case they aren't already present.
	apiClient.UploadUsers(service.NewUserFilter(""))
	// get users
	users := apiClient.ListUsers(service.NewUserFilter(""))
	assert.Equal(t, len(users), 2)
	var orgUser *models.UserSearchHitDTO
	for _, u := range users {
		if u.Login == "tux" {
			orgUser = u
			break
		}
	}
	//Reset if any state exists.
	err := apiClient.DeleteUserFromOrg(orgUser.ID, newOrg.ID)
	//Start CRUD test
	orgUsers := apiClient.ListOrgUsers(newOrg.ID)
	assert.Equal(t, len(orgUsers), 1)
	assert.Equal(t, orgUsers[0].Login, "admin")
	assert.Equal(t, orgUsers[0].Role, "Admin")

	err = apiClient.AddUserToOrg("Admin", orgUser.ID, newOrg.ID)
	assert.Nil(t, err)
	orgUsers = apiClient.ListOrgUsers(newOrg.ID)
	assert.Equal(t, len(orgUsers), 2)
	assert.Equal(t, orgUsers[1].Role, "Admin")
	err = apiClient.UpdateUserInOrg("Viewer", orgUser.ID, newOrg.ID)
	orgUsers = apiClient.ListOrgUsers(newOrg.ID)
	assert.Nil(t, err)
	assert.Equal(t, orgUsers[1].Role, "Viewer")
	err = apiClient.DeleteUserFromOrg(orgUser.ID, newOrg.ID)
	orgUsers = apiClient.ListOrgUsers(newOrg.ID)
	assert.Equal(t, len(orgUsers), 1)
	assert.Nil(t, err)
}
