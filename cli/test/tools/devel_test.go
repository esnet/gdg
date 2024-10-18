package tools

import (
	"strings"
	"testing"

	"github.com/esnet/gdg/pkg/test_tooling"

	"github.com/esnet/gdg/cli"
	"github.com/esnet/gdg/cli/support"
	"github.com/esnet/gdg/internal/service/mocks"
	"github.com/stretchr/testify/assert"
)

func TestDevelSrvInfo(t *testing.T) {
	execMe := func(mock *mocks.GrafanaService, data []byte, optionMockSvc func() support.RootOption) error {
		expected := make(map[string]interface{})
		expected["Database"] = "db"
		expected["Commit"] = "commit"
		expected["Version"] = "version"

		mock.EXPECT().GetServerInfo().Return(expected)
		err := cli.Execute(string(data), []string{"tools", "devel", "srvinfo"}, optionMockSvc())
		return err
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	assert.True(t, strings.Contains(outStr, "Version="))
	assert.True(t, strings.Contains(outStr, "Database="))
	assert.True(t, strings.Contains(outStr, "Commit="))
}

func TestDevelSrvCompletion(t *testing.T) {
	fn := func(args []string) func(mock *mocks.GrafanaService, data []byte, optionMockSvc func() support.RootOption) error {
		return func(mock *mocks.GrafanaService, data []byte, optionMockSvc func() support.RootOption) error {
			err := cli.Execute(string(data), args, optionMockSvc())
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
