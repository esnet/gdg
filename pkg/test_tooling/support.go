package test_tooling

import (
	"context"
	"io"
	"log/slog"
	"os"
	"testing"

	"github.com/esnet/gdg/cli/domain"
	applog "github.com/esnet/gdg/internal/adapter/logger"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/ports"
	"github.com/esnet/gdg/internal/ports/mocks"
	"github.com/stretchr/testify/assert"
)

// SetupAndExecuteMockingServices  will create a mock for varous required entities allowing to test the CLI flag parsing
// process: function that setups mocks and invokes the Execute command
func SetupAndExecuteMockingServices(t *testing.T, process func(mock *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error) (string, func()) {
	testSvc := new(mocks.GrafanaService)
	testSvc.EXPECT().Login().Return()
	getMockSvc := func() ports.GrafanaService {
		return testSvc
	}

	optionMockSvc := func() domain.RootOption {
		return func(response *domain.RootCommand) {
			response.SetUpTest(getMockSvc())
		}
	}

	r, w, cleanup := InterceptStdout()
	err := process(testSvc, optionMockSvc)
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
func InterceptStdout() (*os.File, *os.File, context.CancelFunc) {
	backupStd := os.Stdout
	backupErr := os.Stderr
	r, w, e := os.Pipe()
	if e != nil {
		panic(e)
	}
	// Restore streams
	config.NewConfig("testing")
	cleanup := func() {
		os.Stdout = backupStd
		os.Stderr = backupErr
		applog.InitializeAppLogger(os.Stdout, os.Stderr, false)
	}
	os.Stdout = w
	os.Stderr = w
	applog.InitializeAppLogger(os.Stdout, os.Stderr, false)

	return r, w, cleanup
}
