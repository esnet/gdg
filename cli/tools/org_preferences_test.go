// org_preferences_test.go exercises "gdg tools organizations preferences get"
// and "gdg tools organizations preferences set".
package tools_test

import (
	"testing"

	"github.com/esnet/gdg/cli"
	"github.com/esnet/gdg/cli/domain"
	"github.com/esnet/gdg/internal/ports/outbound/mocks"
	"github.com/esnet/gdg/pkg/test_tooling"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/stretchr/testify/assert"
)

func TestOrgPreferencesGetSuccess(t *testing.T) {
	rootSvc := cli.NewRootService()
	execMe := func(mock *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		mock.EXPECT().GetOrgPreferences().Return(&models.PreferencesSpec{
			HomeDashboardUID: "abc123",
			Theme:            "dark",
			WeekStart:        "monday",
		}, nil)
		return cli.Execute(rootSvc, []string{"tools", "organizations", "preferences", "get"}, optionMockSvc())
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	assert.Contains(t, outStr, "abc123")
	assert.Contains(t, outStr, "dark")
	assert.Contains(t, outStr, "monday")
}

func TestOrgPreferencesSetSuccess(t *testing.T) {
	rootSvc := cli.NewRootService()
	execMe := func(mock *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		mock.EXPECT().GetOrgPreferences().Return(&models.PreferencesSpec{}, nil)
		mock.EXPECT().UploadOrgPreferences("main-org", &models.PreferencesSpec{Theme: "light"}).Return(nil)
		return cli.Execute(rootSvc, []string{
			"tools", "organizations", "preferences", "set",
			"--orgName", "main-org", "--theme", "light",
		}, optionMockSvc())
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	assert.Contains(t, outStr, "Preferences update for organization")
}

// TestOrgPreferencesParentAlias verifies the "pref" alias resolves and shows
// help listing the get/set subcommands. (The upload-failure branch of "set"
// calls log.Fatal, which would kill the test process, so it isn't exercised
// here.)
func TestOrgPreferencesParentAlias(t *testing.T) {
	rootSvc := cli.NewRootService()
	execMe := func(_ *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		return cli.Execute(rootSvc, []string{"tools", "organizations", "pref"}, optionMockSvc())
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	assert.Contains(t, outStr, "get")
	assert.Contains(t, outStr, "set")
}
