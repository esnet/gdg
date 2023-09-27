package cmd

import (
	"fmt"
	"github.com/esnet/gdg/cmd/support"
	"github.com/esnet/gdg/cmd/test_tools"
	"github.com/esnet/gdg/internal/service"
	"github.com/esnet/gdg/internal/service/mocks"
	"github.com/esnet/gdg/internal/version"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"strings"
	"testing"
)

func TestVersionCommand(t *testing.T) {
	testSvc := new(mocks.GrafanaService)
	getMockSvc := func() service.GrafanaService {
		return testSvc
	}

	optionMockSvc := func() support.RootOption {
		return func(response *support.RootCommand) {
			response.GrafanaSvc = getMockSvc
		}
	}
	path, _ := os.Getwd()
	fmt.Println(path)
	r, w, cleanup := test_tools.InterceptStdout()
	data, err := os.ReadFile("../config/testing.yml")
	assert.Nil(t, err)

	err = Execute(string(data), []string{"version"}, optionMockSvc())
	assert.Nil(t, err)
	defer cleanup()
	w.Close()
	out, _ := io.ReadAll(r)
	outStr := string(out)
	assert.True(t, strings.Contains(outStr, "Build Date:"))
	assert.True(t, strings.Contains(outStr, "Git Commit:"))
	assert.True(t, strings.Contains(outStr, "Version:"))
	assert.True(t, strings.Contains(outStr, version.Version))
	assert.True(t, strings.Contains(outStr, "Date:"))
	assert.True(t, strings.Contains(outStr, "Go Version:"))
	assert.True(t, strings.Contains(outStr, "OS / Arch:"))
}

func TestVersionErrCommand(t *testing.T) {
	testSvc := new(mocks.GrafanaService)
	getMockSvc := func() service.GrafanaService {
		return testSvc
	}

	optionMockSvc := func() support.RootOption {
		return func(response *support.RootCommand) {
			response.GrafanaSvc = getMockSvc
		}
	}
	path, _ := os.Getwd()
	fmt.Println(path)
	data, err := os.ReadFile("../config/testing.yml")
	assert.Nil(t, err)

	err = Execute(string(data), []string{"dumb", "dumb"}, optionMockSvc())
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), `command error: unknown command "dumb" for "gdg"`)
}
