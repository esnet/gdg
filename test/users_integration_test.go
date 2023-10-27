package test

import (
	"github.com/esnet/gdg/internal/service"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUsers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	if os.Getenv("TEST_TOKEN_CONFIG") == "1" {
		t.Skip("Skipping Token configuration, Team and User CRUD requires Basic Auth")
	}
	apiClient, _ := initTest(t, nil)
	apiClient.DeleteAllUsers(service.NewUserFilter("")) //clear any previous state
	users := apiClient.ListUsers(service.NewUserFilter(""))
	assert.Equal(t, len(users), 1)
	adminUser := users[0]
	assert.Equal(t, adminUser.ID, int64(1))
	assert.Equal(t, adminUser.Login, "admin")
	assert.Equal(t, adminUser.IsAdmin, true)

}
