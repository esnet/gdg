package test

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/esnet/gdg/pkg/test_tooling/common"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/pkg/test_tooling"
	"github.com/stretchr/testify/assert"
)

func TestServiceAccountCrud(t *testing.T) {
	cfg := config.InitGdgConfig(common.DefaultTestConfig)
	if os.Getenv(test_tooling.EnableTokenTestsEnv) == "1" {
		t.Skip("Skipping Token configuration, Organization CRUD requires Basic SecureData")
	}
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

	name := gofakeit.Name()
	account, err := apiClient.CreateServiceAccount(name, "admin", 0)
	assert.NoError(t, err)
	assert.Equal(t, account.Name, name)
	// another svc
	_, err = apiClient.CreateServiceAccount(gofakeit.Name(), "admin", 0)
	serviceAccounts := apiClient.ListServiceAccounts()
	assert.Equal(t, len(serviceAccounts), 2)
	tokenName := gofakeit.Name()
	token, err := apiClient.CreateServiceAccountToken(account.ID, tokenName, 0)

	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(token.Key, "glsa_"))
	tokens, err := apiClient.ListServiceAccountsTokens(account.ID)
	assert.NoError(t, err)
	assert.Equal(t, len(tokens), 1)
	assert.Equal(t, tokens[0].Name, tokenName)
	assert.NoError(t, apiClient.DeleteServiceAccount(account.ID))
	serviceAccounts = apiClient.ListServiceAccounts()
	assert.Equal(t, len(serviceAccounts), 1)
	names := apiClient.DeleteAllServiceAccounts()
	assert.Equal(t, len(names), 1)
	serviceAccounts = apiClient.ListServiceAccounts()
	assert.Equal(t, len(serviceAccounts), 0)
}
