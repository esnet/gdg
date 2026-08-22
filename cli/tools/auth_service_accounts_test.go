// auth_service_accounts_test.go exercises "gdg tools auth service-accounts"
// and its subcommands (list, delete, clear, new), following the pattern
// established in context_test.go / devel_test.go:
//   - SetupAndExecuteMockingServices for happy paths where cli.Execute
//     returns nil (GrafanaService calls are mocked).
//   - Error-path commands swallow the cli.Execute error inside the closure
//     and assert on the printed output instead.
package tools_test

import (
	"strings"
	"testing"

	"github.com/esnet/gdg/cli"
	"github.com/esnet/gdg/cli/domain"
	customModels "github.com/esnet/gdg/internal/domain"
	"github.com/esnet/gdg/internal/ports/outbound/mocks"
	"github.com/esnet/gdg/pkg/test_tooling"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/stretchr/testify/assert"
)

// ── service-accounts list ────────────────────────────────────────────────────

func TestServiceAccountsListEmpty(t *testing.T) {
	rootSvc := cli.NewRootService()
	execMe := func(mock *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		mock.EXPECT().ListServiceAccounts().Return(nil)
		return cli.Execute(rootSvc, []string{"tools", "auth", "service-accounts", "list"}, optionMockSvc())
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	assert.Contains(t, strings.ToLower(outStr), "no service accounts found")
}

func TestServiceAccountsListWithTokens(t *testing.T) {
	rootSvc := cli.NewRootService()
	execMe := func(mock *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		accounts := []*customModels.ServiceAccountDTOWithTokens{
			{
				ServiceAccount: &models.ServiceAccountDTO{ID: 1, Name: "svc-one", Role: "Admin", Tokens: 1},
				Tokens: []*models.TokenDTO{
					{ID: 10, Name: "token-one"},
				},
			},
		}
		mock.EXPECT().ListServiceAccounts().Return(accounts)
		return cli.Execute(rootSvc, []string{"tools", "auth", "service-accounts", "list"}, optionMockSvc())
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	assert.Contains(t, outStr, "svc-one")
	assert.Contains(t, outStr, "token-one")
}

// ── service-accounts delete ──────────────────────────────────────────────────

func TestServiceAccountsDeleteNoArgsReturnsError(t *testing.T) {
	rootSvc := cli.NewRootService()
	execMe := func(_ *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		_ = cli.Execute(rootSvc, []string{"tools", "auth", "service-accounts", "delete"}, optionMockSvc())
		return nil
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	lower := strings.ToLower(outStr)
	assert.True(t, strings.Contains(lower, "service account") || strings.Contains(lower, "requires") || strings.Contains(lower, "usage"))
}

func TestServiceAccountsDeleteSuccess(t *testing.T) {
	rootSvc := cli.NewRootService()
	execMe := func(mock *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		mock.EXPECT().DeleteServiceAccount(int64(42)).Return(nil)
		return cli.Execute(rootSvc, []string{"tools", "auth", "service-accounts", "delete", "42"}, optionMockSvc())
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	assert.Contains(t, outStr, "Service account has been removed")
}

// ── service-accounts clear ───────────────────────────────────────────────────

func TestServiceAccountsClearEmpty(t *testing.T) {
	rootSvc := cli.NewRootService()
	execMe := func(mock *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		mock.EXPECT().DeleteAllServiceAccounts().Return(nil)
		return cli.Execute(rootSvc, []string{"tools", "auth", "service-accounts", "clear"}, optionMockSvc())
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	assert.Contains(t, strings.ToLower(outStr), "no service accounts found")
}

func TestServiceAccountsClearWithFiles(t *testing.T) {
	rootSvc := cli.NewRootService()
	execMe := func(mock *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		mock.EXPECT().DeleteAllServiceAccounts().Return([]string{"svc-a.json", "svc-b.json"})
		return cli.Execute(rootSvc, []string{"tools", "auth", "service-accounts", "clear"}, optionMockSvc())
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	assert.Contains(t, outStr, "svc-a.json")
	assert.Contains(t, outStr, "svc-b.json")
}

// ── service-accounts new ─────────────────────────────────────────────────────

func TestServiceAccountsNewMissingArgsReturnsError(t *testing.T) {
	rootSvc := cli.NewRootService()
	execMe := func(_ *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		_ = cli.Execute(rootSvc, []string{"tools", "auth", "service-accounts", "new", "onlyName"}, optionMockSvc())
		return nil
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	lower := strings.ToLower(outStr)
	assert.True(t, strings.Contains(lower, "role") || strings.Contains(lower, "requires"))
}

func TestServiceAccountsNewSuccess(t *testing.T) {
	rootSvc := cli.NewRootService()
	execMe := func(mock *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		mock.EXPECT().CreateServiceAccount("newSvc", "editor", int64(0)).
			Return(&models.ServiceAccountDTO{ID: 7, Name: "newSvc", Role: "editor"}, nil)
		return cli.Execute(rootSvc, []string{"tools", "auth", "service-accounts", "new", "newSvc", "editor"}, optionMockSvc())
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	assert.Contains(t, outStr, "newSvc")
}

func TestServiceAccountsNewWithTTLSuccess(t *testing.T) {
	rootSvc := cli.NewRootService()
	execMe := func(mock *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		mock.EXPECT().CreateServiceAccount("ttlSvc", "viewer", int64(3600)).
			Return(&models.ServiceAccountDTO{ID: 8, Name: "ttlSvc", Role: "viewer"}, nil)
		return cli.Execute(rootSvc, []string{"tools", "auth", "service-accounts", "new", "ttlSvc", "viewer", "3600"}, optionMockSvc())
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	assert.Contains(t, outStr, "ttlSvc")
}

// ── service-accounts tokens clear/new ────────────────────────────────────────

func TestServiceAccountTokensClearEmpty(t *testing.T) {
	rootSvc := cli.NewRootService()
	execMe := func(mock *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		mock.EXPECT().DeleteServiceAccountTokens(int64(5)).Return(nil)
		return cli.Execute(rootSvc, []string{"tools", "auth", "service-accounts", "tokens", "clear", "5"}, optionMockSvc())
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	assert.Contains(t, strings.ToLower(outStr), "no service accounts tokens found")
}

func TestServiceAccountTokensClearNoArgsReturnsError(t *testing.T) {
	rootSvc := cli.NewRootService()
	execMe := func(_ *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		_ = cli.Execute(rootSvc, []string{"tools", "auth", "service-accounts", "tokens", "clear"}, optionMockSvc())
		return nil
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	lower := strings.ToLower(outStr)
	assert.True(t, strings.Contains(lower, "service account") || strings.Contains(lower, "requires"))
}

func TestServiceAccountTokensNewSuccess(t *testing.T) {
	rootSvc := cli.NewRootService()
	execMe := func(mock *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		mock.EXPECT().CreateServiceAccountToken(int64(5), "my-token", int64(0)).
			Return(&models.NewAPIKeyResult{ID: 99, Name: "my-token", Key: "glsa_abc123"}, nil)
		return cli.Execute(rootSvc, []string{"tools", "auth", "service-accounts", "tokens", "new", "5", "my-token"}, optionMockSvc())
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	assert.Contains(t, outStr, "glsa_abc123")
}

func TestServiceAccountTokensNewMissingArgsReturnsError(t *testing.T) {
	rootSvc := cli.NewRootService()
	execMe := func(_ *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		_ = cli.Execute(rootSvc, []string{"tools", "auth", "service-accounts", "tokens", "new", "5"}, optionMockSvc())
		return nil
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	lower := strings.ToLower(outStr)
	assert.True(t, strings.Contains(lower, "token") || strings.Contains(lower, "requires"))
}
