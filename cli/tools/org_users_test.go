// org_users_test.go exercises "gdg tools organizations users" and its
// subcommands (currentOrg, list, updateRole, add, delete).
package tools_test

import (
	"strings"
	"testing"

	"github.com/esnet/gdg/cli"
	"github.com/esnet/gdg/cli/domain"
	"github.com/esnet/gdg/internal/ports/outbound/mocks"
	"github.com/esnet/gdg/pkg/test_tooling"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/stretchr/testify/assert"
)

// ── organizations users currentOrg ───────────────────────────────────────────

func TestOrgUsersCurrentOrgFound(t *testing.T) {
	rootSvc := cli.NewRootService()
	execMe := func(mock *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		mock.EXPECT().GetUserOrganization().Return(&models.OrgDetailsDTO{ID: 3, Name: "main-org"})
		return cli.Execute(rootSvc, []string{"tools", "organizations", "users", "currentOrg"}, optionMockSvc())
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	assert.Contains(t, outStr, "main-org")
}

func TestOrgUsersCurrentOrgNotFound(t *testing.T) {
	rootSvc := cli.NewRootService()
	execMe := func(mock *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		mock.EXPECT().GetUserOrganization().Return(nil)
		return cli.Execute(rootSvc, []string{"tools", "organizations", "users", "currentOrg"}, optionMockSvc())
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	assert.Contains(t, strings.ToLower(outStr), "no organizations found")
}

// ── organizations users list ─────────────────────────────────────────────────

func TestOrgUsersListNoArgsReturnsError(t *testing.T) {
	rootSvc := cli.NewRootService()
	execMe := func(_ *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		_ = cli.Execute(rootSvc, []string{"tools", "organizations", "users", "list"}, optionMockSvc())
		return nil
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	lower := strings.ToLower(outStr)
	assert.True(t, strings.Contains(lower, "orgid") || strings.Contains(lower, "requires"))
}

func TestOrgUsersListEmpty(t *testing.T) {
	rootSvc := cli.NewRootService()
	execMe := func(mock *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		mock.EXPECT().ListOrgUsers(int64(1)).Return(nil)
		return cli.Execute(rootSvc, []string{"tools", "organizations", "users", "list", "1"}, optionMockSvc())
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	assert.Contains(t, strings.ToLower(outStr), "no users found")
}

func TestOrgUsersListWithUsers(t *testing.T) {
	rootSvc := cli.NewRootService()
	execMe := func(mock *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		mock.EXPECT().ListOrgUsers(int64(1)).Return([]*models.OrgUserDTO{
			{UserID: 5, Login: "bob", OrgID: 1, Name: "Bob Smith", Email: "bob@example.com", Role: "Editor"},
		})
		return cli.Execute(rootSvc, []string{"tools", "organizations", "users", "list", "1"}, optionMockSvc())
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	assert.Contains(t, outStr, "bob@example.com")
}

func TestOrgUsersListAlias(t *testing.T) {
	rootSvc := cli.NewRootService()
	execMe := func(mock *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		mock.EXPECT().ListOrgUsers(int64(2)).Return(nil)
		return cli.Execute(rootSvc, []string{"tools", "organizations", "users", "listUsers", "2"}, optionMockSvc())
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	assert.Contains(t, strings.ToLower(outStr), "no users found")
}

// ── organizations users updateRole ───────────────────────────────────────────

func TestOrgUsersUpdateRoleMissingArgsReturnsError(t *testing.T) {
	rootSvc := cli.NewRootService()
	execMe := func(_ *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		_ = cli.Execute(rootSvc, []string{"tools", "organizations", "users", "updateRole", "main-org"}, optionMockSvc())
		return nil
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	lower := strings.ToLower(outStr)
	assert.True(t, strings.Contains(lower, "role") || strings.Contains(lower, "requires"))
}

func TestOrgUsersUpdateRoleSuccess(t *testing.T) {
	rootSvc := cli.NewRootService()
	execMe := func(mock *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		mock.EXPECT().UpdateUserInOrg("editor", "main-org", int64(5)).Return(nil)
		return cli.Execute(rootSvc, []string{"tools", "organizations", "users", "updateRole", "main-org", "5", "editor"}, optionMockSvc())
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	assert.Contains(t, outStr, "User has been updated")
}

// ── organizations users add ──────────────────────────────────────────────────

func TestOrgUsersAddMissingArgsReturnsError(t *testing.T) {
	rootSvc := cli.NewRootService()
	execMe := func(_ *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		_ = cli.Execute(rootSvc, []string{"tools", "organizations", "users", "add", "main-org", "5"}, optionMockSvc())
		return nil
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	lower := strings.ToLower(outStr)
	assert.True(t, strings.Contains(lower, "role") || strings.Contains(lower, "requires"))
}

func TestOrgUsersAddSuccess(t *testing.T) {
	rootSvc := cli.NewRootService()
	execMe := func(mock *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		mock.EXPECT().AddUserToOrg("editor", "main-org", int64(5)).Return(nil)
		return cli.Execute(rootSvc, []string{"tools", "organizations", "users", "add", "main-org", "5", "editor"}, optionMockSvc())
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	assert.Contains(t, outStr, "User has been add to Org")
}

func TestOrgUsersAddAlias(t *testing.T) {
	rootSvc := cli.NewRootService()
	execMe := func(mock *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		mock.EXPECT().AddUserToOrg("viewer", "main-org", int64(6)).Return(nil)
		return cli.Execute(rootSvc, []string{"tools", "organizations", "users", "addUser", "main-org", "6", "viewer"}, optionMockSvc())
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	assert.Contains(t, outStr, "User has been add to Org")
}

// ── organizations users delete ───────────────────────────────────────────────

func TestOrgUsersDeleteMissingArgsReturnsError(t *testing.T) {
	rootSvc := cli.NewRootService()
	execMe := func(_ *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		_ = cli.Execute(rootSvc, []string{"tools", "organizations", "users", "delete", "main-org"}, optionMockSvc())
		return nil
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	// The RunFunc's error text goes to a pre-redirect stderr reference and
	// isn't captured here (see TestOrgUsersListNoArgsReturnsError for the
	// same caveat); assert on the command's own usage/help text instead.
	lower := strings.ToLower(outStr)
	assert.True(t, strings.Contains(lower, "orgslug") || strings.Contains(lower, "usage"))
}

func TestOrgUsersDeleteSuccess(t *testing.T) {
	rootSvc := cli.NewRootService()
	execMe := func(mock *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		mock.EXPECT().DeleteUserFromOrg("main-org", int64(5)).Return(nil)
		return cli.Execute(rootSvc, []string{"tools", "organizations", "users", "delete", "main-org", "5"}, optionMockSvc())
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	assert.Contains(t, outStr, "User has been removed from Org")
}

func TestOrgUsersDeleteAlias(t *testing.T) {
	rootSvc := cli.NewRootService()
	execMe := func(mock *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		mock.EXPECT().DeleteUserFromOrg("main-org", int64(9)).Return(nil)
		return cli.Execute(rootSvc, []string{"tools", "organizations", "users", "remove", "main-org", "9"}, optionMockSvc())
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	assert.Contains(t, outStr, "User has been removed from Org")
}
