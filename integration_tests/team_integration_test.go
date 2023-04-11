package integration_tests

import (
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestTeamCRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	apiClient, _ := initTest(t)
	log.Info("Exporting all teams")
	apiClient.ExportTeams()
	log.Info("Listing all Teams")
	teams := apiClient.ListTeams()
	assert.Equal(t, len(teams), 2)
	var firstDsItem = teams[0]
	assert.Equal(t, firstDsItem.Name, "engineers")
	var secondDsItem = teams[1]
	assert.Equal(t, secondDsItem.Name, "musicians")
	//Import Teams
	log.Info("Importing teams")
	list := apiClient.ImportTeams()
	assert.Equal(t, len(list), len(teams))
	// Add and List Team Members
	log.Info("Add team members")
	_, err := apiClient.AddTeamMember("engineers", "admin")
	assert.NoError(t, err)
	_, err = apiClient.AddTeamMember("musicians", "admin")
	assert.NoError(t, err)
	log.Info("Checking team members")
	listEngineers := apiClient.ListTeamMembers("engineers")
	listMusicians := apiClient.ListTeamMembers("musicians")
	assert.Equal(t, len(listEngineers), 1)
	assert.Equal(t, len(listMusicians), 1)
	// Delete teams
	log.Info("Deleting Teams")
	_, err = apiClient.DeleteTeam("engineers")
	assert.NoError(t, err)
	_, err = apiClient.DeleteTeam("musicians")
	assert.NoError(t, err)
	log.Info("List Teams again")
	teams = apiClient.ListTeams()
	assert.Equal(t, len(teams), 0)
}
