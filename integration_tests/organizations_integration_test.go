package integration_tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrgs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	apiClient, _ := initTest(t)
	orgs := apiClient.ListOrganizations()
	assert.Equal(t, len(orgs), 1)
	mainOrg := orgs[0]
	assert.Equal(t, mainOrg.ID, int64(1))
	assert.Equal(t, mainOrg.Name, "Main Org.")

}
