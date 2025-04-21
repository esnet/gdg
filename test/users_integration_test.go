package test

import (
	"os"
	"testing"

	"github.com/esnet/gdg/internal/config"

	"github.com/samber/lo"

	"github.com/esnet/gdg/internal/service"
	"github.com/esnet/gdg/pkg/test_tooling"
	"github.com/grafana/grafana-openapi-client-go/models"

	"github.com/stretchr/testify/assert"
)

func TestUsers(t *testing.T) {
	if os.Getenv(test_tooling.EnableTokenTestsEnv) == test_tooling.FeatureEnabled {
		t.Skip("Skipping Token configuration, Team and User CRUD requires Basic SecureData")
	}
	config.InitGdgConfig("testing")
	apiClient, _, cleanup := test_tooling.InitTest(t, service.DefaultConfigProvider, nil)
	defer cleanup()
	apiClient.DeleteAllUsers(service.NewUserFilter("")) // clear any previous state
	users := apiClient.ListUsers(service.NewUserFilter(""))
	assert.Equal(t, len(users), 1)
	adminUser := users[0]
	assert.Equal(t, adminUser.ID, int64(1))
	assert.Equal(t, adminUser.Login, "admin")
	assert.Equal(t, adminUser.IsAdmin, true)
	newUsers := apiClient.UploadUsers(service.NewUserFilter(""))
	assert.Equal(t, len(newUsers), 2)
	users = apiClient.ListUsers(service.NewUserFilter(""))
	assert.Equal(t, len(users), 3)
	var user *models.UserSearchHitDTO
	user = lo.FirstOrEmpty(lo.Filter(users, func(userItem *models.UserSearchHitDTO, index int) bool {
		return userItem.Name == "supertux"
	}))
	assert.NotNil(t, user)
	assert.Equal(t, user.Login, "tux")
	assert.Equal(t, user.Email, "s@s.com")
	assert.Equal(t, user.LastSeenAtAge, "10 years")
}
