package test

import (
	"io"
	"strings"
	"testing"

	"github.com/esnet/gdg/pkg/test_tooling"
	"github.com/esnet/gdg/pkg/test_tooling/path"

	"github.com/esnet/gdg/cli"
	"github.com/esnet/gdg/cli/support"
	"github.com/esnet/gdg/internal/service"
	"github.com/esnet/gdg/internal/service/mocks"
	"github.com/esnet/gdg/internal/version"
	"github.com/esnet/gdg/pkg/test_tooling/common"
	"github.com/stretchr/testify/assert"
)

func TestVersionCommand(t *testing.T) {
	assert.NoError(t, path.FixTestDir("test", "../.."))
	execMe := func(mock *mocks.GrafanaService, data []byte, optionMockSvc func() support.RootOption) error {
		err := cli.Execute(string(data), []string{"version"}, optionMockSvc())
		return err
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	assert.True(t, strings.Contains(outStr, "Build Date:"))
	assert.True(t, strings.Contains(outStr, "Git Commit:"))
	assert.True(t, strings.Contains(outStr, "Version:"))
	assert.True(t, strings.Contains(outStr, version.Version))
	assert.True(t, strings.Contains(outStr, "Date:"))
	assert.True(t, strings.Contains(outStr, "Go Version:"))
	assert.True(t, strings.Contains(outStr, "OS / Arch:"))
}

func TestVersionErrCommand(t *testing.T) {
	assert.NoError(t, path.FixTestDir("test", "../.."))
	testSvc := new(mocks.GrafanaService)
	getMockSvc := func() service.GrafanaService {
		return testSvc
	}

	optionMockSvc := func() support.RootOption {
		return func(response *support.RootCommand) {
			response.GrafanaSvc = getMockSvc
		}
	}
	r, w, cleanup := test_tooling.InterceptStdout()
	defer cleanup()
	err := cli.Execute(common.DefaultTestConfig, []string{"dumb", "dumb"}, optionMockSvc())
	assert.NotNil(t, err)
	assert.NoError(t, w.Close())

	assert.Equal(t, err.Error(), `command error: unknown command "dumb" for "gdg"`)
	out, _ := io.ReadAll(r)
	output := string(out)
	assert.True(t, strings.Contains(output, "gdg [command] --help\" for more information about a command"))
}
