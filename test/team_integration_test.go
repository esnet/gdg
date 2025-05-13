package test

import (
	"log/slog"
	"os"
	"testing"

	"github.com/esnet/gdg/internal/config"

	"github.com/esnet/gdg/internal/service"
	"github.com/esnet/gdg/pkg/test_tooling"
	"github.com/grafana/grafana-openapi-client-go/models"
	"golang.org/x/exp/maps"

	"github.com/stretchr/testify/assert"
)

func TestTeamCRUD(t *testing.T) {
	if os.Getenv(test_tooling.EnableTokenTestsEnv) == test_tooling.FeatureEnabled {
		t.Skip("Skipping Token configuration, Team and User CRUD requires Basic SecureData")
	}
	config.InitGdgConfig("testing")
	apiClient, _, cleanup := test_tooling.InitTest(t, service.DefaultConfigProvider, nil)
	filter := service.NewTeamFilter("")
	defer cleanup()
	slog.Info("Exporting current user list")
	apiClient.UploadUsers(service.NewUserFilter(""))
	users := apiClient.ListUsers(service.NewUserFilter(""))
	assert.Equal(t, len(users), 3)
	slog.Info("Exporting all teams")
	apiClient.UploadTeams(filter)
	slog.Info("Listing all Teams")
	teamsMap := apiClient.ListTeams(filter)
	teams := maps.Keys(teamsMap)
	assert.Equal(t, len(teams), 2)
	var engineerTeam *models.TeamDTO
	var musicianTeam *models.TeamDTO
	for ndx, team := range teams {
		if team.Name == "engineers" {
			engineerTeam = teams[ndx]
		} else if team.Name == "musicians" {
			musicianTeam = teams[ndx]
		}
	}
	assert.NotNil(t, engineerTeam)
	assert.Equal(t, engineerTeam.Name, "engineers")
	engineers := teamsMap[engineerTeam]
	assert.Equal(t, len(engineers), 2)
	assert.Equal(t, engineers[1].Login, "tux")
	assert.Equal(t, musicianTeam.Name, "musicians")
	// Import Teams
	slog.Info("Importing teams")
	list := apiClient.DownloadTeams(filter)
	assert.Equal(t, len(list), len(teams))
	// CleanUp
	_, err := apiClient.DeleteTeam(filter)
	assert.Nil(t, err)
	// Remove Users
	apiClient.DeleteAllUsers(service.NewUserFilter(""))
}
