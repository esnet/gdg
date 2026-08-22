// organizations_test.go exercises "gdg tools organizations set" and
// "gdg tools organizations tokenOrg".
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

// ── organizations set ────────────────────────────────────────────────────────

func TestOrgSetNoFlagsReturnsError(t *testing.T) {
	rootSvc := cli.NewRootService()
	execMe := func(_ *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		_ = cli.Execute(rootSvc, []string{"tools", "organizations", "set"}, optionMockSvc())
		return nil
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	lower := strings.ToLower(outStr)
	assert.True(t, strings.Contains(lower, "orgname") || strings.Contains(lower, "orgslugname"))
}

func TestOrgSetByNameSuccess(t *testing.T) {
	rootSvc := cli.NewRootService()
	execMe := func(mock *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		mock.EXPECT().SetOrganizationByName("main-org", false).Return(nil)
		mock.EXPECT().InitOrganizations().Return()
		mock.EXPECT().GetUserOrganization().Return(&models.OrgDetailsDTO{ID: 1, Name: "main-org"})
		return cli.Execute(rootSvc, []string{"tools", "organizations", "set", "--orgName", "main-org"}, optionMockSvc())
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	assert.Contains(t, outStr, "main-org")
}

func TestOrgSetBySlugSuccess(t *testing.T) {
	rootSvc := cli.NewRootService()
	execMe := func(mock *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		mock.EXPECT().SetOrganizationByName("main-org-slug", true).Return(nil)
		mock.EXPECT().InitOrganizations().Return()
		mock.EXPECT().GetUserOrganization().Return(&models.OrgDetailsDTO{ID: 2, Name: "main-org-slug"})
		return cli.Execute(rootSvc, []string{"tools", "organizations", "set", "--orgSlugName", "main-org-slug"}, optionMockSvc())
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	assert.Contains(t, outStr, "main-org-slug")
}

// ── organizations tokenOrg ───────────────────────────────────────────────────

func TestOrgTokenOrgFound(t *testing.T) {
	rootSvc := cli.NewRootService()
	execMe := func(mock *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		mock.EXPECT().GetTokenOrganization().Return(&models.OrgDetailsDTO{ID: 4, Name: "token-org"})
		return cli.Execute(rootSvc, []string{"tools", "organizations", "tokenOrg"}, optionMockSvc())
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	assert.Contains(t, outStr, "token-org")
}

func TestOrgTokenOrgNotFound(t *testing.T) {
	rootSvc := cli.NewRootService()
	execMe := func(mock *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		mock.EXPECT().GetTokenOrganization().Return(nil)
		return cli.Execute(rootSvc, []string{"tools", "organizations", "tokenOrg"}, optionMockSvc())
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	assert.Contains(t, strings.ToLower(outStr), "no tokens were found")
}

// ── organizations parent help / alias ────────────────────────────────────────

func TestOrgAliasOrg(t *testing.T) {
	rootSvc := cli.NewRootService()
	execMe := func(_ *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		return cli.Execute(rootSvc, []string{"tools", "org"}, optionMockSvc())
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	lower := strings.ToLower(outStr)
	assert.Contains(t, lower, "set")
}
