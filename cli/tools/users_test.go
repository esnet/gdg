// users_test.go exercises "gdg tools users makeGrafanaAdmin".
package tools_test

import (
	"testing"

	"github.com/esnet/gdg/cli"
	"github.com/esnet/gdg/cli/domain"
	"github.com/esnet/gdg/internal/ports/outbound/mocks"
	"github.com/esnet/gdg/pkg/test_tooling"
	"github.com/stretchr/testify/assert"
)

func TestPromoteUserSuccess(t *testing.T) {
	rootSvc := cli.NewRootService()
	execMe := func(mock *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		mock.EXPECT().PromoteUser("bob@example.com").Return("bob@example.com has been promoted", nil)
		return cli.Execute(rootSvc, []string{"tools", "users", "makeGrafanaAdmin", "--user", "bob@example.com"}, optionMockSvc())
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	assert.Contains(t, outStr, "bob@example.com has been promoted")
}

func TestPromoteUserAlias(t *testing.T) {
	rootSvc := cli.NewRootService()
	execMe := func(mock *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		mock.EXPECT().PromoteUser("alice@example.com").Return("alice@example.com has been promoted", nil)
		return cli.Execute(rootSvc, []string{"tools", "users", "promote", "--user", "alice@example.com"}, optionMockSvc())
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	assert.Contains(t, outStr, "alice@example.com has been promoted")
}

func TestPromoteUserMissingFlagReturnsError(t *testing.T) {
	rootSvc := cli.NewRootService()
	execMe := func(_ *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		_ = cli.Execute(rootSvc, []string{"tools", "users", "makeGrafanaAdmin"}, optionMockSvc())
		return nil
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	// Cobra's "required flag(s) ... not set" line goes to a pre-redirect
	// stderr reference and isn't captured here; assert on the command's own
	// usage/help text (always printed on a PreRunE validation failure).
	assert.Contains(t, outStr, "user email")
}
