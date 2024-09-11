package test

import (
	"io"
	"log/slog"
	"os"
	"testing"

	"github.com/esnet/gdg/cli/support"
	"github.com/esnet/gdg/internal/config"
	applog "github.com/esnet/gdg/internal/log"
	"github.com/esnet/gdg/internal/service"
	"github.com/esnet/gdg/internal/service/mocks"
	"github.com/esnet/gdg/pkg/test_tooling/common"

	"github.com/stretchr/testify/assert"
)

// setupAndExecuteMockingServices  will create a mock for varous required entities allowing to test the CLI flag parsing
// process: function that setups mocks and invokes the Execute command
func setupAndExecuteMockingServices(t *testing.T, process func(mock *mocks.GrafanaService, data []byte, optionMockSvc func() support.RootOption) error) (string, func()) {
	testSvc := new(mocks.GrafanaService)
	getMockSvc := func() service.GrafanaService {
		return testSvc
	}

	optionMockSvc := func() support.RootOption {
		return func(response *support.RootCommand) {
			response.GrafanaSvc = getMockSvc
		}
	}

	r, w, cleanup := InterceptStdout()
	data, err := os.ReadFile("../../config/" + common.DefaultTestConfig)
	assert.Nil(t, err)

	err = process(testSvc, data, optionMockSvc)
	assert.Nil(t, err)
	defer cleanup()
	err = w.Close()
	if err != nil {
		slog.Warn("unable to close write stream")
	}
	clean := func() {
		defer r.Close()
	}
	out, _ := io.ReadAll(r)
	outStr := string(out)
	return outStr, clean
}

// InterceptStdout is a test helper function that will redirect all stdout in and out to a different file stream.
// It returns the stdout, stderr, and a function to be invoked to close the streams.
func InterceptStdout() (*os.File, *os.File, func()) {
	backupStd := os.Stdout
	backupErr := os.Stderr
	r, w, _ := os.Pipe()
	// Restore streams
	config.InitGdgConfig("testing", "")
	applog.InitializeAppLogger(w, w, false)
	cleanup := func() {
		os.Stdout = backupStd
		os.Stderr = backupErr
		applog.InitializeAppLogger(os.Stdout, os.Stderr, false)
	}
	os.Stdout = w
	os.Stderr = w

	return r, w, cleanup
}
