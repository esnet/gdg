package cmd

import (
	"fmt"
	"github.com/esnet/gdg/cmd/support"
	"github.com/esnet/gdg/internal/service"
	"github.com/esnet/gdg/internal/service/mocks"
	"github.com/esnet/gdg/internal/version"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"strings"
	"testing"
)

func interceptStdout() (*os.File, *os.File, func()) {
	backupStd := os.Stdout
	backupErr := os.Stderr
	r, w, _ := os.Pipe()
	//Restore streams
	cleanup := func() {
		os.Stdout = backupStd
		os.Stderr = backupErr
	}
	os.Stdout = w
	os.Stderr = w

	return r, w, cleanup

}

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
	r, w, cleanup := interceptStdout()
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
