package tools_test

import (
	"strings"
	"testing"

	"github.com/esnet/gdg/cli/domain"
	"github.com/esnet/gdg/pkg/test_tooling"

	"github.com/esnet/gdg/cli"
	"github.com/esnet/gdg/internal/ports/mocks"
	"github.com/stretchr/testify/assert"
)

func TestDevelSrvInfo(t *testing.T) {
	execMe := func(mock *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		expected := make(map[string]any)
		expected["Database"] = "db"
		expected["Commit"] = "commit"
		expected["Version"] = "version"

		mock.EXPECT().Login().Return()
		mock.EXPECT().GetServerInfo().Return(expected)
		rootSvc := cli.NewRootService()
		err := cli.Execute(rootSvc, []string{"tools", "devel", "srvinfo"}, optionMockSvc())
		return err
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	assert.True(t, strings.Contains(outStr, "Version="))
	assert.True(t, strings.Contains(outStr, "Database="))
	assert.True(t, strings.Contains(outStr, "Commit="))
}

func TestDevelSrvCompletion(t *testing.T) {
	rootSvc := cli.NewRootService()
	fn := func(args []string) func(mock *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		return func(mock *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
			err := cli.Execute(rootSvc, args, optionMockSvc())
			return err
		}
	}

	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, fn([]string{"tools", "devel", "completion", "fish"}))
	assert.True(t, strings.Contains(outStr, "fish"))
	assert.True(t, strings.Contains(outStr, "__completion_prepare_completions"))
	closeReader()
	outStr, closeReader = test_tooling.SetupAndExecuteMockingServices(t, fn([]string{"tools", "devel", "completion", "bash"}))
	assert.True(t, strings.Contains(outStr, "bash"))
	assert.True(t, strings.Contains(outStr, "flag_parsing_disabled"))
	closeReader()
	outStr, closeReader = test_tooling.SetupAndExecuteMockingServices(t, fn([]string{"tools", "devel", "completion", "zsh"}))
	assert.True(t, strings.Contains(outStr, "shellCompDirectiveKeepOrder"))
	closeReader()
}
