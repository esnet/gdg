// config_crud_test.go tests the exported, non-TUI functions in config_crud.go.
// These functions operate on GDGAppConfiguration directly and can be exercised
// without a terminal, using the same temp-config pattern as config_s3_crud_test.go.
package config_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/config/config_domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newCrudTestApp writes a minimal multi-context gdg.yml to a temp directory and
// loads it via NewConfig, mirroring the pattern used in config_s3_crud_test.go.
func newCrudTestApp(t *testing.T, extraContexts ...string) (*config_domain.GDGAppConfiguration, string) {
	t.Helper()

	outputDir := t.TempDir()
	cfgDir := t.TempDir()
	cfgPath := filepath.Join(cfgDir, "gdg-crud-test.yml")

	// Build extra context YAML blocks.
	extra := ""
	for _, name := range extraContexts {
		extra += fmt.Sprintf(`
  %s:
    url: http://localhost:3000
    output_path: %s
    watched:
      - General
`, name, outputDir)
	}

	yaml := fmt.Sprintf(`
context_name: primary
contexts:
  primary:
    url: http://grafana-primary:3000
    output_path: %s
    watched:
      - General
  secondary:
    url: http://grafana-secondary:3000
    output_path: %s
    watched:
      - General
%s
storage_engine: {}
`, outputDir, outputDir, extra)

	require.NoError(t, os.WriteFile(cfgPath, []byte(yaml), 0o600))
	app := config.NewConfig(cfgPath)
	return app, outputDir
}

// ── CopyContext ───────────────────────────────────────────────────────────────

func TestCopyContext_CreatesNewContext(t *testing.T) {
	app, _ := newCrudTestApp(t)

	config.CopyContext(app, "primary", "cloned")

	contexts := app.GetContexts()
	assert.Contains(t, contexts, "cloned", "copied context should exist in the map")
	assert.Contains(t, contexts, "primary", "source context should still exist")
}

func TestCopyContext_UpdatesActiveContext(t *testing.T) {
	app, _ := newCrudTestApp(t)

	config.CopyContext(app, "primary", "newctx")

	// After copying, the active context name should be set to the destination.
	assert.Equal(t, "newctx", app.ContextName)
}

func TestCopyContext_CopiedURLMatches(t *testing.T) {
	app, _ := newCrudTestApp(t)

	originalURL := app.GetContexts()["primary"].URL
	config.CopyContext(app, "primary", "mirror")

	copiedURL := app.GetContexts()["mirror"].URL
	assert.Equal(t, originalURL, copiedURL, "copied context should inherit the source URL")
}

func TestCopyContext_PersistsToDisk(t *testing.T) {
	app, _ := newCrudTestApp(t)

	config.CopyContext(app, "primary", "saved")

	// Reload config from the same file and check the copy is present.
	cfgFile := app.GetViperConfig().ConfigFileUsed()
	reloaded := config.NewConfig(cfgFile)
	assert.Contains(t, reloaded.GetContexts(), "saved",
		"copied context should persist across a config reload")
}

func TestCopyContext_IsDeepCopy(t *testing.T) {
	app, _ := newCrudTestApp(t)

	config.CopyContext(app, "primary", "deep")

	// Mutating the copy must not affect the source.
	app.GetContexts()["deep"].URL = "http://mutated"
	assert.NotEqual(t, app.GetContexts()["primary"].URL, app.GetContexts()["deep"].URL,
		"mutating the copy should not affect the source context")
}

// ── DeleteContext ─────────────────────────────────────────────────────────────

func TestDeleteContext_RemovesContext(t *testing.T) {
	app, _ := newCrudTestApp(t)

	config.DeleteContext(app, "secondary")

	assert.NotContains(t, app.GetContexts(), "secondary",
		"deleted context should be removed from the map")
	assert.Contains(t, app.GetContexts(), "primary",
		"surviving context should still be present")
}

func TestDeleteContext_CaseInsensitive(t *testing.T) {
	app, _ := newCrudTestApp(t)

	// DeleteContext lower-cases the name before lookup.
	config.DeleteContext(app, "SECONDARY")

	assert.NotContains(t, app.GetContexts(), "secondary")
}

func TestDeleteContext_SwitchesActiveContext(t *testing.T) {
	app, _ := newCrudTestApp(t)

	// Delete the currently active context; the active context should switch to another one.
	config.DeleteContext(app, "primary")

	remaining := app.GetContexts()
	assert.NotContains(t, remaining, "primary")
	// The new active context name must be one of the survivors.
	assert.Contains(t, remaining, app.ContextName,
		"active context should be updated to a surviving context")
}

func TestDeleteContext_RemovesAuthFileIfPresent(t *testing.T) {
	app, outputDir := newCrudTestApp(t)

	// Create a fake auth file that DeleteContext should remove.
	secureDir := filepath.Join(outputDir, "secure")
	require.NoError(t, os.MkdirAll(secureDir, 0o750))
	authFile := filepath.Join(secureDir, "auth_secondary.yaml")
	require.NoError(t, os.WriteFile(authFile, []byte("token: fake\n"), 0o600))

	config.DeleteContext(app, "secondary")

	_, statErr := os.Stat(authFile)
	assert.True(t, os.IsNotExist(statErr),
		"auth file should be removed when the context is deleted")
}

func TestDeleteContext_PersistsToDisk(t *testing.T) {
	app, _ := newCrudTestApp(t)

	config.DeleteContext(app, "secondary")

	cfgFile := app.GetViperConfig().ConfigFileUsed()
	reloaded := config.NewConfig(cfgFile)
	assert.NotContains(t, reloaded.GetContexts(), "secondary",
		"deleted context should not reappear after a config reload")
	assert.Contains(t, reloaded.GetContexts(), "primary",
		"surviving context should still be present after reload")
}

// ── ClearContexts ─────────────────────────────────────────────────────────────

func TestClearContexts_ResetsToExampleOnly(t *testing.T) {
	app, _ := newCrudTestApp(t)

	config.ClearContexts(app)

	contexts := app.GetContexts()
	assert.Len(t, contexts, 1, "only the 'example' context should remain")
	assert.Contains(t, contexts, "example", "the reset context must be named 'example'")
}

func TestClearContexts_SetsActiveContextToExample(t *testing.T) {
	app, _ := newCrudTestApp(t)

	config.ClearContexts(app)

	assert.Equal(t, "example", app.ContextName)
}

func TestClearContexts_RemovesAllPreviousContexts(t *testing.T) {
	app, _ := newCrudTestApp(t, "extra1", "extra2")

	config.ClearContexts(app)

	contexts := app.GetContexts()
	assert.NotContains(t, contexts, "primary")
	assert.NotContains(t, contexts, "secondary")
	assert.NotContains(t, contexts, "extra1")
	assert.NotContains(t, contexts, "extra2")
}

func TestClearContexts_PersistsToDisk(t *testing.T) {
	app, _ := newCrudTestApp(t)

	config.ClearContexts(app)

	cfgFile := app.GetViperConfig().ConfigFileUsed()
	reloaded := config.NewConfig(cfgFile)
	assert.Len(t, reloaded.GetContexts(), 1)
	assert.Contains(t, reloaded.GetContexts(), "example")
}
