package test

import (
	"golang.org/x/exp/slices"
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
