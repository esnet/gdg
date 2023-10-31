package test

import (
	"fmt"
	"github.com/esnet/gdg/cmd"
	"github.com/esnet/gdg/cmd/support"
	"github.com/esnet/gdg/internal/service"
	"github.com/esnet/gdg/internal/service/mocks"
	"github.com/esnet/gdg/internal/version"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
)

func TestVersionCommand(t *testing.T) {
	var execMe = func(mock *mocks.GrafanaService, data []byte, optionMockSvc func() support.RootOption) error {
		err := cmd.Execute(string(data), []string{"version"}, optionMockSvc())
		return err
	}
	outStr, closeReader := setupAndExecuteMockingServices(t, execMe)
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
	data, err := os.ReadFile("../../config/testing.yml")
	assert.Nil(t, err)

	err = cmd.Execute(string(data), []string{"dumb", "dumb"}, optionMockSvc())
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), `command error: unknown command "dumb" for "gdg"`)
}
