package test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/esnet/gdg/internal/tools/ptr"
	"github.com/esnet/gdg/pkg/test_tooling/common"

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
	cfg := config.InitGdgConfig(common.DefaultTestConfig)
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
	filter := service.NewTeamFilter("")
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
		if ptr.ValueOrDefault(team.Name, "") == "engineers" {
			engineerTeam = teams[ndx]
		} else if ptr.ValueOrDefault(team.Name, "") == "musicians" {
			musicianTeam = teams[ndx]
		}
	}
	assert.NotNil(t, engineerTeam)
	assert.Equal(t, ptr.ValueOrDefault(engineerTeam.Name, ""), "engineers")
	engineers := teamsMap[engineerTeam]
	assert.Equal(t, len(engineers), 2)
	assert.Equal(t, engineers[1].Login, "tux")
	assert.Equal(t, ptr.ValueOrDefault(musicianTeam.Name, ""), "musicians")
	// Import Teams
	slog.Info("Importing teams")
	list := apiClient.DownloadTeams(filter)
	assert.Equal(t, len(list), len(teams))
	for _, members := range list {
		assert.True(t, len(members) > 0)
	}

	// CleanUp
	_, err = apiClient.DeleteTeam(filter)
	assert.Nil(t, err)
	// Remove Users
	apiClient.DeleteAllUsers(service.NewUserFilter(""))
}
