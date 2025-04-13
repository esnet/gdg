package test

import (
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service"
	"github.com/esnet/gdg/pkg/test_tooling"
	"github.com/stretchr/testify/assert"
)

func TestServiceAccountCrud(t *testing.T) {
	config.InitGdgConfig("testing")
	if os.Getenv(test_tooling.EnableTokenTestsEnv) == "1" {
		t.Skip("Skipping Token configuration, Organization CRUD requires Basic SecureData")
	}
	apiClient, _, cleanup := test_tooling.InitTest(t, service.DefaultConfigProvider, nil)
	defer func() {
		err := cleanup()
		if err != nil {
			slog.Warn("Unable to clean up after service account testtests")
		}
	}()

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
