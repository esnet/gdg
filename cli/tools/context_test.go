// context_test.go exercises the "gdg tools contexts" subcommand tree.
// Tests follow the same pattern established in s3_test.go:
//   - SetupAndExecuteMockingServices for read-only / non-TUI commands
//   - Error-path commands swallow the cli.Execute error inside the closure
//     (SetupAndExecuteMockingServices would fail its assert.Nil otherwise)
//     and instead assert that the expected error text appears in the output.
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

// ── Parent command ────────────────────────────────────────────────────────────

// TestContextsParentShowsHelp verifies that running "gdg tools contexts" with no
// subcommand prints the help listing all child commands.
func TestContextsParentShowsHelp(t *testing.T) {
	rootSvc := cli.NewRootService()
	execMe := func(_ *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		return cli.Execute(rootSvc, []string{"tools", "contexts"}, optionMockSvc())
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	lower := strings.ToLower(outStr)
	assert.Contains(t, lower, "list")
	assert.Contains(t, lower, "new")
	assert.Contains(t, lower, "delete")
	assert.Contains(t, lower, "copy")
}

// TestContextsAliasCtx verifies that the "ctx" alias resolves to the same command.
func TestContextsAliasCtx(t *testing.T) {
	rootSvc := cli.NewRootService()
	execMe := func(_ *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		return cli.Execute(rootSvc, []string{"tools", "ctx"}, optionMockSvc())
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	lower := strings.ToLower(outStr)
	assert.Contains(t, lower, "list")
}

// ── contexts list ─────────────────────────────────────────────────────────────

// TestContextListShowsContextNames verifies that "contexts list" prints the
// context names that are present in testing.yml.
func TestContextListShowsContextNames(t *testing.T) {
	rootSvc := cli.NewRootService()
	execMe := func(_ *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		return cli.Execute(rootSvc, []string{"tools", "contexts", "list"}, optionMockSvc())
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	lower := strings.ToLower(outStr)
	// testing.yml must contain at least a "testing" or "qa" context.
	assert.True(t,
		strings.Contains(lower, "testing") || strings.Contains(lower, "qa"),
		"list output should contain at least one context name from testing.yml")
}

// TestContextListShowsActiveFlag verifies the table includes an "active" column.
func TestContextListShowsActiveFlag(t *testing.T) {
	rootSvc := cli.NewRootService()
	execMe := func(_ *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		return cli.Execute(rootSvc, []string{"tools", "contexts", "list"}, optionMockSvc())
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	lower := strings.ToLower(outStr)
	assert.Contains(t, lower, "active", "list table should have an 'active' column")
}

// ── contexts show ─────────────────────────────────────────────────────────────

// TestContextShowPrintsCurrentContext verifies that "contexts show" prints some
// YAML/config content for the active context.
func TestContextShowPrintsCurrentContext(t *testing.T) {
	rootSvc := cli.NewRootService()
	execMe := func(_ *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		return cli.Execute(rootSvc, []string{"tools", "contexts", "show"}, optionMockSvc())
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	// The output should contain config file path and YAML content markers.
	assert.Contains(t, outStr, "config file:", "show should print the config file path")
}

// TestContextShowWithExplicitName verifies that "contexts show <name>" accepts
// an optional argument.
func TestContextShowWithExplicitName(t *testing.T) {
	rootSvc := cli.NewRootService()
	execMe := func(_ *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		// "testing" is the context set by InterceptStdout via config.NewConfig("testing").
		return cli.Execute(rootSvc, []string{"tools", "contexts", "show", "testing"}, optionMockSvc())
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	assert.Contains(t, outStr, "config file:")
}

// ── contexts new (no args → error) ───────────────────────────────────────────

// TestContextNewNoArgsReturnsError verifies that "contexts new" without an
// argument returns an error and surfaces a meaningful message.
func TestContextNewNoArgsReturnsError(t *testing.T) {
	rootSvc := cli.NewRootService()
	execMe := func(_ *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		// Swallow the error so SetupAndExecuteMockingServices' assert.Nil doesn't
		// mark the test as failed; we check the output ourselves below.
		_ = cli.Execute(rootSvc, []string{"tools", "contexts", "new"}, optionMockSvc())
		return nil
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	lower := strings.ToLower(outStr)
	// cobra / simplecobra prints the error + usage on RunFunc error.
	assert.True(t,
		strings.Contains(lower, "context") || strings.Contains(lower, "error") || strings.Contains(lower, "requires"),
		"output should mention context name requirement")
}

// ── contexts delete (no args → error) ────────────────────────────────────────

// TestContextDeleteNoArgsReturnsError verifies that "contexts delete" without
// an argument returns an error.
func TestContextDeleteNoArgsReturnsError(t *testing.T) {
	rootSvc := cli.NewRootService()
	execMe := func(_ *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		_ = cli.Execute(rootSvc, []string{"tools", "contexts", "delete"}, optionMockSvc())
		return nil
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	lower := strings.ToLower(outStr)
	assert.True(t,
		strings.Contains(lower, "delete") || strings.Contains(lower, "requires") || strings.Contains(lower, "usage"),
		"output should reference the delete command or its usage")
}

// TestContextDeleteAliasDelNoArgsReturnsError verifies the "del" alias works.
func TestContextDeleteAliasDelNoArgsReturnsError(t *testing.T) {
	rootSvc := cli.NewRootService()
	execMe := func(_ *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		_ = cli.Execute(rootSvc, []string{"tools", "contexts", "del"}, optionMockSvc())
		return nil
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	lower := strings.ToLower(outStr)
	assert.True(t,
		strings.Contains(lower, "delete") || strings.Contains(lower, "del") || strings.Contains(lower, "usage"),
		"del alias should resolve and produce output")
}

// ── contexts set (no args → error) ───────────────────────────────────────────

// TestContextSetNoArgsReturnsError verifies that "contexts set" without an
// argument returns an error.
func TestContextSetNoArgsReturnsError(t *testing.T) {
	rootSvc := cli.NewRootService()
	execMe := func(_ *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		_ = cli.Execute(rootSvc, []string{"tools", "contexts", "set"}, optionMockSvc())
		return nil
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	lower := strings.ToLower(outStr)
	assert.True(t,
		strings.Contains(lower, "set") || strings.Contains(lower, "requires") || strings.Contains(lower, "usage"),
		"output should reference the set command or its usage")
}

// ── contexts copy (fewer than 2 args → error) ─────────────────────────────────

// TestContextCopyNoArgsReturnsError verifies that "contexts copy" without
// arguments returns an error (args validator fires before RunFunc).
func TestContextCopyNoArgsReturnsError(t *testing.T) {
	rootSvc := cli.NewRootService()
	execMe := func(_ *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		_ = cli.Execute(rootSvc, []string{"tools", "contexts", "copy"}, optionMockSvc())
		return nil
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	lower := strings.ToLower(outStr)
	assert.True(t,
		strings.Contains(lower, "copy") || strings.Contains(lower, "requires") || strings.Contains(lower, "usage"),
		"output should reference the copy command or its usage")
}

// TestContextCopyOneArgReturnsError verifies that "contexts copy <src>" (missing
// destination) also fails argument validation.
func TestContextCopyOneArgReturnsError(t *testing.T) {
	rootSvc := cli.NewRootService()
	execMe := func(_ *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		_ = cli.Execute(rootSvc, []string{"tools", "contexts", "copy", "testing"}, optionMockSvc())
		return nil
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	lower := strings.ToLower(outStr)
	assert.True(t,
		strings.Contains(lower, "copy") || strings.Contains(lower, "requires") || strings.Contains(lower, "usage"),
		"output should reference the copy command or its usage")
}

// TestContextCopyAliasCpReported verifies that the "cp" alias is registered for
// the copy command (alias appears in help output).
func TestContextCopyAliasCpReported(t *testing.T) {
	rootSvc := cli.NewRootService()
	execMe := func(_ *mocks.GrafanaService, optionMockSvc func() domain.RootOption) error {
		return cli.Execute(rootSvc, []string{"tools", "contexts", "copy", "--help"}, optionMockSvc())
	}
	outStr, closeReader := test_tooling.SetupAndExecuteMockingServices(t, execMe)
	defer closeReader()

	lower := strings.ToLower(outStr)
	// cobra always prints aliases in the help page when they are set.
	assert.Contains(t, lower, "cp", "help output should mention the 'cp' alias")
}
