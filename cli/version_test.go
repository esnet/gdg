package cli

import (
	"io"
	"strings"
	"testing"

	"github.com/esnet/gdg/cli/domain"
	"github.com/esnet/gdg/internal/ports"
	"github.com/esnet/gdg/pkg/test_tooling"
	"github.com/esnet/gdg/pkg/test_tooling/path"
	domain2 "github.com/esnet/gdg/pkg/version"

	"github.com/esnet/gdg/internal/ports/mocks"
	"github.com/stretchr/testify/assert"
)

func TestDefaultConfigCommand(t *testing.T) {
	assert.NoError(t, path.FixTestDir("cli", ".."))
	rootSvc := NewRootService()
	execMe := func(mock *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		err := Execute(rootSvc, []string{"default-config"}, optionMockSvc())
		return err
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	assert.True(t, strings.Contains(outStr, "storage_engine"))
	assert.True(t, strings.Contains(outStr, "global"))
	assert.True(t, strings.Contains(outStr, "context_name:"))
	assert.True(t, strings.Contains(outStr, "contexts:"))
}

func TestVersionCommand(t *testing.T) {
	assert.NoError(t, path.FixTestDir("cli", ".."))
	rootSvc := NewRootService()
	execMe := func(mock *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		err := Execute(rootSvc, []string{"version"}, optionMockSvc())
		return err
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	assert.True(t, strings.Contains(outStr, "Build Date:"))
	assert.True(t, strings.Contains(outStr, "Git Commit:"))
	assert.True(t, strings.Contains(outStr, "Version:"))
	assert.True(t, strings.Contains(outStr, domain2.Version))
	assert.True(t, strings.Contains(outStr, "Date:"))
	assert.True(t, strings.Contains(outStr, "Go Version:"))
	assert.True(t, strings.Contains(outStr, "OS / Arch:"))
}

func TestVersionErrCommand(t *testing.T) {
	assert.NoError(t, path.FixTestDir("cli", ".."))
	testSvc := new(mocks.GrafanaService)
	getMockSvc := func() ports.GrafanaService {
		return testSvc
	}

	optionMockSvc := func() domain.RootOption {
		return func(response *domain.RootCommand) {
			response.SetUpTest(getMockSvc())
		}
	}
	r, w, cleanup := test_tooling.InterceptStdout()
	defer cleanup()
	rootSvc := NewRootService()
	err := Execute(rootSvc, []string{"dumb", "dumb"}, optionMockSvc())
	assert.NotNil(t, err)
	assert.NoError(t, w.Close())

	assert.Equal(t, err.Error(), `command error: unknown command "dumb" for "gdg"`)
	out, _ := io.ReadAll(r)
	output := string(out)
	assert.True(t, strings.Contains(output, "gdg [command] --help\" for more information about a command"))
}
