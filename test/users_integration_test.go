package test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/pkg/test_tooling/common"
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
	userFilter := service.NewUserFilter("")
	cfg := config.InitGdgConfig(common.DefaultTestConfig)
	var r *test_tooling.InitContainerResult
	err := Retry(context.Background(), DefaultRetryAttempts, func() error {
		r = test_tooling.InitTest(t, cfg, nil)
		return r.Err
	})
	assert.NotNil(t, r)
	assert.NoError(t, err)
	defer func() {
		errCleanUp := r.CleanUp()
		if errCleanUp != nil {
			slog.Warn("Unable to clean up after test", "test", t.Name())
		}
	}()
	apiClient := r.ApiClient
	apiClient.DeleteAllUsers(userFilter) // clear any previous state
	users := apiClient.ListUsers(userFilter)
	assert.Equal(t, len(users), 1)
	adminUser := users[0]
	assert.Equal(t, adminUser.ID, int64(1))
	assert.Equal(t, adminUser.Login, "admin")
	assert.Equal(t, adminUser.IsAdmin, true)
	// Only upload users matching filter
	newUsers := apiClient.UploadUsers(service.NewUserFilter("foobar"))
	assert.Equal(t, len(newUsers), 1)
	assert.Equal(t, newUsers[0].Email, "s@s.com")
	// upload remaining user that do not already exist
	newUsers = apiClient.UploadUsers(userFilter)
	assert.Equal(t, len(newUsers), 1)
	assert.Equal(t, newUsers[0].Email, "bob@aol.com")
	users = apiClient.ListUsers(userFilter)
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
