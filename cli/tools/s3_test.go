package tools_test

import (
	"strings"
	"testing"

	"github.com/esnet/gdg/cli"
	"github.com/esnet/gdg/cli/domain"
	"github.com/esnet/gdg/internal/ports/mocks"
	"github.com/esnet/gdg/pkg/test_tooling"
	"github.com/stretchr/testify/assert"
)

// TestS3ListShowsEngines exercises "gdg tools contexts s3 list" against the
// testing.yml config (which defines a "test" storage engine) and verifies that
// the table output contains the expected column headers and engine data.
func TestS3ListShowsEngines(t *testing.T) {
	rootSvc := cli.NewRootService()
	execMe := func(_ *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		return cli.Execute(rootSvc, []string{"tools", "contexts", "s3", "list"}, optionMockSvc())
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	lower := strings.ToLower(outStr)

	// Column headers (go-pretty StyleLight uppercases them, so compare lowercase)
	assert.Contains(t, lower, "label")
	assert.Contains(t, lower, "cloud_type")
	assert.Contains(t, lower, "endpoint")
	assert.Contains(t, lower, "bucket")

	// Data from testing.yml storage_engine.test
	assert.Contains(t, outStr, "test")
	assert.Contains(t, outStr, "http://localhost:9000")
}

// TestS3ListViaStorageAlias verifies that the "storage" alias on the s3 command
// and the "ctx" alias on the contexts command both resolve correctly.
func TestS3ListViaStorageAlias(t *testing.T) {
	rootSvc := cli.NewRootService()
	execMe := func(_ *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		return cli.Execute(rootSvc, []string{"tools", "ctx", "storage", "list"}, optionMockSvc())
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	// Aliases should produce the same table as the canonical command
	assert.Contains(t, strings.ToLower(outStr), "label")
	assert.Contains(t, outStr, "http://localhost:9000")
}

// TestS3ParentShowsHelp verifies that running "gdg tools contexts s3" with no
// subcommand prints the help listing all three child commands.
func TestS3ParentShowsHelp(t *testing.T) {
	rootSvc := cli.NewRootService()
	execMe := func(_ *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		return cli.Execute(rootSvc, []string{"tools", "contexts", "s3"}, optionMockSvc())
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	lower := strings.ToLower(outStr)
	assert.Contains(t, lower, "new")
	assert.Contains(t, lower, "list")
	assert.Contains(t, lower, "delete")
}

// TestS3DeleteNoArgsShowsHelp verifies that running "gdg tools contexts s3 delete"
// without a label argument prints the command's usage help rather than returning an error.
func TestS3DeleteNoArgsShowsHelp(t *testing.T) {
	rootSvc := cli.NewRootService()
	execMe := func(_ *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		return cli.Execute(rootSvc, []string{"tools", "contexts", "s3", "delete"}, optionMockSvc())
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	lower := strings.ToLower(outStr)
	// Cobra help output always includes "usage" and the command name
	assert.Contains(t, lower, "usage")
	assert.Contains(t, lower, "delete")
}
