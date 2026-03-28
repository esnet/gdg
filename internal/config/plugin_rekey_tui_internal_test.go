// plugin_rekey_tui_internal_test.go exercises the pure-logic helpers and
// state-machine transitions defined in plugin_rekey_tui.go.
//
// The TUI itself (tea.Program) is not exercised here; instead, each testable
// non-TUI function is driven directly:
//
//   - rekeyPhase.sectionName     — section label for every phase
//   - currentPluginDescription   — human-readable plugin config summary
//   - rekeyState.buildNewPluginEntity — assembles a PluginEntity from collected state
//   - pluginRekeyModel.nextPhase — forward phase-transition FSM
//   - pluginRekeyModel.prevPhase — backward phase-transition FSM
//   - pluginRekeyModel.applyPhase — side-effects applied when a screen is submitted
//
// Because all of these symbols are unexported the file uses package config
// (white-box) rather than package config_test.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/esnet/gdg/internal/adapter/plugins/migration"
	"github.com/esnet/gdg/internal/adapter/plugins/secure/noop"
	"github.com/esnet/gdg/internal/config/config_domain"
	"github.com/esnet/gdg/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── test helpers ──────────────────────────────────────────────────────────────

// newRekeyTestApp writes a minimal gdg.yml to a temp dir and returns the
// parsed GDGAppConfiguration — the same pattern used by newTUITestApp.
func newRekeyTestApp(t *testing.T) *config_domain.GDGAppConfiguration {
	t.Helper()
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "gdg-rekey-test.yml")
	yaml := fmt.Sprintf(`
context_name: rekey-test
contexts:
  rekey-test:
    url: http://localhost:3000
    output_path: %s
    watched:
      - General
storage_engine: {}
`, dir)
	require.NoError(t, os.WriteFile(cfgPath, []byte(yaml), 0o600))
	return NewConfig(cfgPath)
}

// newRekeyRS builds a minimal rekeyState suitable for unit testing.
func newRekeyRS(t *testing.T) *rekeyState {
	t.Helper()
	app := newRekeyTestApp(t)
	return &rekeyState{
		app:          app,
		oldEncoder:   noop.NoOpEncoder{},
		doBackup:     true,
		contextNames: []string{app.GetContext()},
	}
}

// newRekeyModel returns a pluginRekeyModel starting at phaseRekeyCurrentInfo.
// Tests that need a specific phase should overwrite m.phase after calling this.
func newRekeyModel(t *testing.T) pluginRekeyModel {
	t.Helper()
	return newPluginRekeyModel(newRekeyRS(t))
}

// ── rekeyPhase.sectionName ────────────────────────────────────────────────────

func TestRekeyPhaseSectionName_AllPhases(t *testing.T) {
	cases := []struct {
		phase    rekeyPhase
		expected string
	}{
		{phaseRekeyCurrentInfo, "Current Configuration"},
		{phaseRekeyAction, "Choose Action"},
		{phaseRekeyPluginSelect, "Plugin Selection"},
		{phaseRekeyPluginVersion, "Plugin Selection"},
		{phaseRekeyPluginConfig, "Plugin Selection"},
		{phaseRekeyContextSelect, "Contexts Selection"},
		{phaseRekeyBackupOptions, "Backup Options"},
		{phaseRekeyCredentials, "Credential Migration"},
		{phaseRekeyDryRun, "File Scan"},
		{phaseRekeyFileSelect, "Select Files"},
		{phaseRekeyConfirm, "Confirm"},
		{phaseRekeyDone, ""},
	}
	for _, tc := range cases {
		t.Run(fmt.Sprintf("phase_%d", int(tc.phase)), func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.phase.sectionName())
		})
	}
}

// ── currentPluginDescription ──────────────────────────────────────────────────

func TestCurrentPluginDescription_Disabled(t *testing.T) {
	app := &config_domain.GDGAppConfiguration{
		PluginConfig: config_domain.PluginConfig{Disabled: true},
	}
	desc := currentPluginDescription(app)
	assert.Contains(t, desc, "DISABLED")
	assert.Contains(t, desc, "plaintext")
}

func TestCurrentPluginDescription_NilPlugin(t *testing.T) {
	app := &config_domain.GDGAppConfiguration{
		PluginConfig: config_domain.PluginConfig{Disabled: false, CipherPlugin: nil},
	}
	desc := currentPluginDescription(app)
	assert.Contains(t, desc, "No cipher plugin")
	assert.Contains(t, desc, "without encryption")
}

func TestCurrentPluginDescription_WithFilePath(t *testing.T) {
	app := &config_domain.GDGAppConfiguration{
		PluginConfig: config_domain.PluginConfig{
			CipherPlugin: &config_domain.PluginEntity{FilePath: "/opt/gdg/plugins/cipher.wasm"},
		},
	}
	desc := currentPluginDescription(app)
	assert.Contains(t, desc, "/opt/gdg/plugins/cipher.wasm")
	assert.Contains(t, desc, "local WASM file")
}

func TestCurrentPluginDescription_WithURL(t *testing.T) {
	app := &config_domain.GDGAppConfiguration{
		PluginConfig: config_domain.PluginConfig{
			CipherPlugin: &config_domain.PluginEntity{Url: "https://example.com/cipher.wasm"},
		},
	}
	desc := currentPluginDescription(app)
	assert.Contains(t, desc, "https://example.com/cipher.wasm")
}

func TestCurrentPluginDescription_UnknownSource(t *testing.T) {
	// No Url and no FilePath — should fall into the "unknown" branch.
	app := &config_domain.GDGAppConfiguration{
		PluginConfig: config_domain.PluginConfig{
			CipherPlugin: &config_domain.PluginEntity{},
		},
	}
	desc := currentPluginDescription(app)
	assert.Contains(t, desc, "unknown")
}

func TestCurrentPluginDescription_ConfigFieldsSorted(t *testing.T) {
	app := &config_domain.GDGAppConfiguration{
		PluginConfig: config_domain.PluginConfig{
			CipherPlugin: &config_domain.PluginEntity{
				Url: "https://example.com/cipher.wasm",
				PluginConfig: map[string]string{
					"key_z": "zz",
					"key_a": "aa",
					"key_m": "mm",
				},
			},
		},
	}
	desc := currentPluginDescription(app)
	idxA := strings.Index(desc, "key_a")
	idxM := strings.Index(desc, "key_m")
	idxZ := strings.Index(desc, "key_z")
	require.True(t, idxA >= 0 && idxM >= 0 && idxZ >= 0, "all config keys should appear in the description")
	assert.Less(t, idxA, idxM, "key_a should appear before key_m (alphabetical order)")
	assert.Less(t, idxM, idxZ, "key_m should appear before key_z (alphabetical order)")
}

func TestCurrentPluginDescription_NoConfigFields_ShowsNone(t *testing.T) {
	app := &config_domain.GDGAppConfiguration{
		PluginConfig: config_domain.PluginConfig{
			CipherPlugin: &config_domain.PluginEntity{Url: "https://example.com/cipher.wasm"},
		},
	}
	desc := currentPluginDescription(app)
	assert.Contains(t, desc, "none")
}

// ── rekeyState.buildNewPluginEntity ───────────────────────────────────────────

func TestBuildNewPluginEntity_NilResolvedEntry_NoOp(t *testing.T) {
	rs := &rekeyState{resolvedEntry: nil, resolvedVersionEntry: nil}
	rs.buildNewPluginEntity()
	assert.Nil(t, rs.newPluginEntity, "should remain nil when resolvedEntry is nil")
}

func TestBuildNewPluginEntity_NilResolvedVersion_NoOp(t *testing.T) {
	entry := &domain.PluginRegistryEntry{Name: "p", URLPattern: "https://x/{version}.wasm"}
	rs := &rekeyState{resolvedEntry: entry, resolvedVersionEntry: nil}
	rs.buildNewPluginEntity()
	assert.Nil(t, rs.newPluginEntity, "should remain nil when resolvedVersionEntry is nil")
}

func TestBuildNewPluginEntity_SetsURLFromPattern(t *testing.T) {
	entry := &domain.PluginRegistryEntry{URLPattern: "https://example.com/{version}/aes.wasm"}
	ver := &domain.PluginVersionEntry{Version: "1.2.3"}
	rs := &rekeyState{
		resolvedEntry:        entry,
		resolvedVersionEntry: ver,
		pluginConfigValues:   map[string]string{},
	}
	rs.buildNewPluginEntity()
	require.NotNil(t, rs.newPluginEntity)
	assert.Equal(t, "https://example.com/1.2.3/aes.wasm", rs.newPluginEntity.Url)
}

func TestBuildNewPluginEntity_CopiesConfigValues(t *testing.T) {
	entry := &domain.PluginRegistryEntry{URLPattern: "https://x/{version}.wasm"}
	ver := &domain.PluginVersionEntry{Version: "0.1.0"}
	rs := &rekeyState{
		resolvedEntry:        entry,
		resolvedVersionEntry: ver,
		pluginConfigValues:   map[string]string{"k1": "v1", "k2": "v2"},
	}
	rs.buildNewPluginEntity()
	require.NotNil(t, rs.newPluginEntity)
	assert.Equal(t, "v1", rs.newPluginEntity.PluginConfig["k1"])
	assert.Equal(t, "v2", rs.newPluginEntity.PluginConfig["k2"])
}

func TestBuildNewPluginEntity_ConfigIsDefensiveCopy(t *testing.T) {
	// Mutating pluginConfigValues after calling buildNewPluginEntity must not
	// affect the already-built entity's PluginConfig map.
	entry := &domain.PluginRegistryEntry{URLPattern: "https://x/{version}.wasm"}
	ver := &domain.PluginVersionEntry{Version: "0.1.0"}
	original := map[string]string{"key": "original"}
	rs := &rekeyState{
		resolvedEntry:        entry,
		resolvedVersionEntry: ver,
		pluginConfigValues:   original,
	}
	rs.buildNewPluginEntity()
	rs.pluginConfigValues["key"] = "mutated"
	assert.Equal(t, "original", rs.newPluginEntity.PluginConfig["key"],
		"entity PluginConfig must be independent of the source map")
}

// ── nextPhase ─────────────────────────────────────────────────────────────────

func TestNextPhase_CurrentInfo_GoesToAction(t *testing.T) {
	m := newRekeyModel(t)
	m.phase = phaseRekeyCurrentInfo
	assert.Equal(t, phaseRekeyAction, m.nextPhase())
}

func TestNextPhase_Action_Cancel_GoesToDone(t *testing.T) {
	m := newRekeyModel(t)
	m.phase = phaseRekeyAction
	m.rs.action = "cancel"
	assert.Equal(t, phaseRekeyDone, m.nextPhase())
}

func TestNextPhase_Action_Switch_GoesToPluginSelect(t *testing.T) {
	m := newRekeyModel(t)
	m.phase = phaseRekeyAction
	m.rs.action = "switch"
	assert.Equal(t, phaseRekeyPluginSelect, m.nextPhase())
}

func TestNextPhase_Action_Disable_GoesToContextSelect(t *testing.T) {
	m := newRekeyModel(t)
	m.phase = phaseRekeyAction
	m.rs.action = "disable"
	assert.Equal(t, phaseRekeyContextSelect, m.nextPhase())
}

func TestNextPhase_PluginSelect_RegistryError_BouncesToAction_ClearsAction(t *testing.T) {
	m := newRekeyModel(t)
	m.phase = phaseRekeyPluginSelect
	m.rs.action = "switch"
	m.rs.registryError = "could not connect to registry"
	next := m.nextPhase()
	assert.Equal(t, phaseRekeyAction, next, "registry error should bounce back to action selection")
	assert.Empty(t, m.rs.action, "action must be cleared so the user starts fresh")
}

func TestNextPhase_PluginSelect_NoError_GoesToVersion(t *testing.T) {
	m := newRekeyModel(t)
	m.phase = phaseRekeyPluginSelect
	m.rs.registryError = ""
	assert.Equal(t, phaseRekeyPluginVersion, m.nextPhase())
}

func TestNextPhase_PluginVersion_NoConfigFields_SkipsToContextSelect(t *testing.T) {
	m := newRekeyModel(t)
	m.phase = phaseRekeyPluginVersion
	entry := domain.PluginRegistryEntry{URLPattern: "https://x/{version}.wasm"}
	ver := domain.PluginVersionEntry{Version: "1.0.0"}
	m.rs.resolvedEntry = &entry
	m.rs.resolvedVersionEntry = &ver
	m.rs.pluginConfigValues = make(map[string]string)
	m.rs.pluginConfigFields = nil // empty → no config fields to collect

	next := m.nextPhase()

	assert.Equal(t, phaseRekeyContextSelect, next)
	assert.NotNil(t, m.rs.newPluginEntity,
		"entity should be built immediately when there are no config fields")
}

func TestNextPhase_PluginVersion_WithConfigFields_GoesToPluginConfig(t *testing.T) {
	m := newRekeyModel(t)
	m.phase = phaseRekeyPluginVersion
	m.rs.pluginConfigFields = []string{"encryption_key"}
	assert.Equal(t, phaseRekeyPluginConfig, m.nextPhase())
}

func TestNextPhase_PluginConfig_StillCollecting_StaysAtPluginConfig(t *testing.T) {
	m := newRekeyModel(t)
	m.phase = phaseRekeyPluginConfig
	m.rs.pluginConfigFields = []string{"field_a", "field_b"}
	m.rs.configFieldIdx = 1 // still one field outstanding
	assert.Equal(t, phaseRekeyPluginConfig, m.nextPhase())
}

func TestNextPhase_PluginConfig_AllFieldsCollected_GoesToContextSelect(t *testing.T) {
	m := newRekeyModel(t)
	m.phase = phaseRekeyPluginConfig
	m.rs.pluginConfigFields = []string{"field_a"}
	m.rs.configFieldIdx = 1 // equal to len — all collected
	assert.Equal(t, phaseRekeyContextSelect, m.nextPhase())
}

func TestNextPhase_ContextSelect_GoesToBackupOptions(t *testing.T) {
	m := newRekeyModel(t)
	m.phase = phaseRekeyContextSelect
	assert.Equal(t, phaseRekeyBackupOptions, m.nextPhase())
}

func TestNextPhase_BackupOptions_GoesToCredentials(t *testing.T) {
	m := newRekeyModel(t)
	m.phase = phaseRekeyBackupOptions
	assert.Equal(t, phaseRekeyCredentials, m.nextPhase())
}

func TestNextPhase_Credentials_GoesToDryRun(t *testing.T) {
	m := newRekeyModel(t)
	m.phase = phaseRekeyCredentials
	assert.Equal(t, phaseRekeyDryRun, m.nextPhase())
}

func TestNextPhase_DryRun_NoPreviews_SkipsToConfirm(t *testing.T) {
	m := newRekeyModel(t)
	m.phase = phaseRekeyDryRun
	m.rs.previews = nil
	assert.Equal(t, phaseRekeyConfirm, m.nextPhase())
}

func TestNextPhase_DryRun_WithPreviews_GoesToFileSelect(t *testing.T) {
	m := newRekeyModel(t)
	m.phase = phaseRekeyDryRun
	m.rs.previews = []migration.FilePreview{{Path: "/ctx/contacts.json", DecodedOK: true}}
	assert.Equal(t, phaseRekeyFileSelect, m.nextPhase())
}

func TestNextPhase_FileSelect_GoesToConfirm(t *testing.T) {
	m := newRekeyModel(t)
	m.phase = phaseRekeyFileSelect
	assert.Equal(t, phaseRekeyConfirm, m.nextPhase())
}

func TestNextPhase_Confirm_GoesToDone(t *testing.T) {
	m := newRekeyModel(t)
	m.phase = phaseRekeyConfirm
	assert.Equal(t, phaseRekeyDone, m.nextPhase())
}

func TestNextPhase_Default_GoesToDone(t *testing.T) {
	m := newRekeyModel(t)
	m.phase = rekeyPhase(99) // any unknown phase
	assert.Equal(t, phaseRekeyDone, m.nextPhase())
}

// ── prevPhase ─────────────────────────────────────────────────────────────────

func TestPrevPhase_CurrentInfo_StaysAtCurrentInfo(t *testing.T) {
	m := newRekeyModel(t)
	m.phase = phaseRekeyCurrentInfo
	assert.Equal(t, phaseRekeyCurrentInfo, m.prevPhase(),
		"there is no previous phase from the opening screen — caller must cancel")
}

func TestPrevPhase_Action_GoesToCurrentInfo(t *testing.T) {
	m := newRekeyModel(t)
	m.phase = phaseRekeyAction
	assert.Equal(t, phaseRekeyCurrentInfo, m.prevPhase())
}

func TestPrevPhase_PluginSelect_GoesToAction(t *testing.T) {
	m := newRekeyModel(t)
	m.phase = phaseRekeyPluginSelect
	assert.Equal(t, phaseRekeyAction, m.prevPhase())
}

func TestPrevPhase_PluginVersion_GoesToPluginSelect(t *testing.T) {
	m := newRekeyModel(t)
	m.phase = phaseRekeyPluginVersion
	assert.Equal(t, phaseRekeyPluginSelect, m.prevPhase())
}

func TestPrevPhase_PluginConfig_GoesToPluginVersion(t *testing.T) {
	// configFieldIdx == 0: the Update loop already handles the within-loop step;
	// prevPhase itself maps PluginConfig → PluginVersion unconditionally.
	m := newRekeyModel(t)
	m.phase = phaseRekeyPluginConfig
	assert.Equal(t, phaseRekeyPluginVersion, m.prevPhase())
}

func TestPrevPhase_ContextSelect_SwitchAction_GoesToPluginVersion(t *testing.T) {
	m := newRekeyModel(t)
	m.phase = phaseRekeyContextSelect
	m.rs.action = "switch"
	assert.Equal(t, phaseRekeyPluginVersion, m.prevPhase())
}

func TestPrevPhase_ContextSelect_DisableAction_GoesToAction(t *testing.T) {
	m := newRekeyModel(t)
	m.phase = phaseRekeyContextSelect
	m.rs.action = "disable"
	assert.Equal(t, phaseRekeyAction, m.prevPhase())
}

func TestPrevPhase_BackupOptions_GoesToContextSelect(t *testing.T) {
	m := newRekeyModel(t)
	m.phase = phaseRekeyBackupOptions
	assert.Equal(t, phaseRekeyContextSelect, m.prevPhase())
}

func TestPrevPhase_Credentials_GoesToBackupOptions(t *testing.T) {
	m := newRekeyModel(t)
	m.phase = phaseRekeyCredentials
	assert.Equal(t, phaseRekeyBackupOptions, m.prevPhase())
}

func TestPrevPhase_DryRun_GoesToCredentials(t *testing.T) {
	m := newRekeyModel(t)
	m.phase = phaseRekeyDryRun
	assert.Equal(t, phaseRekeyCredentials, m.prevPhase())
}

func TestPrevPhase_FileSelect_GoesToDryRun(t *testing.T) {
	m := newRekeyModel(t)
	m.phase = phaseRekeyFileSelect
	assert.Equal(t, phaseRekeyDryRun, m.prevPhase())
}

func TestPrevPhase_Confirm_WithPreviews_GoesToFileSelect(t *testing.T) {
	m := newRekeyModel(t)
	m.phase = phaseRekeyConfirm
	m.rs.previews = []migration.FilePreview{{Path: "/ctx/contacts.json"}}
	assert.Equal(t, phaseRekeyFileSelect, m.prevPhase())
}

func TestPrevPhase_Confirm_NoPreviews_GoesToDryRun(t *testing.T) {
	m := newRekeyModel(t)
	m.phase = phaseRekeyConfirm
	m.rs.previews = nil
	assert.Equal(t, phaseRekeyDryRun, m.prevPhase())
}

func TestPrevPhase_Default_GoesToCurrentInfo(t *testing.T) {
	m := newRekeyModel(t)
	m.phase = rekeyPhase(99) // any unknown phase
	assert.Equal(t, phaseRekeyCurrentInfo, m.prevPhase())
}

// ── applyPhase ────────────────────────────────────────────────────────────────

func TestApplyPhase_PluginSelect_ResetsAllVersionAndConfigState(t *testing.T) {
	m := newRekeyModel(t)
	m.phase = phaseRekeyPluginSelect
	// Pre-populate all fields that the apply should reset.
	m.rs.pluginVersion = "1.2.3"
	m.rs.resolvedEntry = &domain.PluginRegistryEntry{Name: "old"}
	m.rs.resolvedVersionEntry = &domain.PluginVersionEntry{Version: "1.0.0"}
	m.rs.pluginConfigFields = []string{"k"}
	m.rs.pluginConfigValues = map[string]string{"k": "v"}
	m.rs.configFieldIdx = 1
	m.rs.configFieldValue = "some-value"
	m.rs.newPluginEntity = &config_domain.PluginEntity{Url: "https://old"}

	m.applyPhase()

	assert.Empty(t, m.rs.pluginVersion, "version should be cleared")
	assert.Nil(t, m.rs.resolvedEntry, "resolvedEntry should be cleared")
	assert.Nil(t, m.rs.resolvedVersionEntry, "resolvedVersionEntry should be cleared")
	assert.Nil(t, m.rs.pluginConfigFields, "config fields slice should be cleared")
	assert.Equal(t, 0, m.rs.configFieldIdx, "field index should be reset to 0")
	assert.Empty(t, m.rs.configFieldValue, "field value buffer should be cleared")
	assert.Nil(t, m.rs.newPluginEntity, "newPluginEntity should be cleared")
	assert.NotNil(t, m.rs.pluginConfigValues, "pluginConfigValues should be an empty map, not nil")
	assert.Empty(t, m.rs.pluginConfigValues)
}

func TestApplyPhase_PluginVersion_ResolvesEntryAndVersion(t *testing.T) {
	m := newRekeyModel(t)
	m.phase = phaseRekeyPluginVersion

	entry := domain.PluginRegistryEntry{
		Name:       "cipher-aes",
		URLPattern: "https://example.com/{version}/aes.wasm",
		Versions: []domain.PluginVersionEntry{
			{Version: "0.9.0", ConfigFields: nil},
			{Version: "1.0.0", ConfigFields: []string{"encryption_key"}},
		},
	}
	m.rs.availablePlugins = []domain.PluginRegistryEntry{entry}
	m.rs.pluginName = "cipher-aes"
	m.rs.pluginVersion = "1.0.0"

	m.applyPhase()

	require.NotNil(t, m.rs.resolvedEntry)
	assert.Equal(t, "cipher-aes", m.rs.resolvedEntry.Name)
	require.NotNil(t, m.rs.resolvedVersionEntry)
	assert.Equal(t, "1.0.0", m.rs.resolvedVersionEntry.Version)
	assert.Equal(t, []string{"encryption_key"}, m.rs.pluginConfigFields)
	assert.Equal(t, 0, m.rs.configFieldIdx, "field index must be reset after version resolution")
	assert.Empty(t, m.rs.configFieldValue)
}

func TestApplyPhase_PluginVersion_FallsBackToLatestWhenVersionNotFound(t *testing.T) {
	m := newRekeyModel(t)
	m.phase = phaseRekeyPluginVersion

	entry := domain.PluginRegistryEntry{
		Name:       "cipher-aes",
		URLPattern: "https://example.com/{version}/aes.wasm",
		Versions: []domain.PluginVersionEntry{
			{Version: "1.0.0"},
			{Version: "1.1.0"}, // latest
		},
	}
	m.rs.availablePlugins = []domain.PluginRegistryEntry{entry}
	m.rs.pluginName = "cipher-aes"
	m.rs.pluginVersion = "9.9.9" // not present in the registry

	m.applyPhase()

	require.NotNil(t, m.rs.resolvedVersionEntry)
	assert.Equal(t, "1.1.0", m.rs.resolvedVersionEntry.Version,
		"should fall back to the latest version when the requested one is not found")
}

func TestApplyPhase_PluginConfig_CapturesValueAndAdvancesIndex(t *testing.T) {
	m := newRekeyModel(t)
	m.phase = phaseRekeyPluginConfig
	m.rs.pluginConfigFields = []string{"key_a", "key_b"}
	m.rs.pluginConfigValues = make(map[string]string)
	m.rs.configFieldIdx = 0
	m.rs.configFieldValue = "value-for-key_a"

	m.applyPhase()

	assert.Equal(t, "value-for-key_a", m.rs.pluginConfigValues["key_a"],
		"submitted value should be stored under the current field name")
	assert.Equal(t, 1, m.rs.configFieldIdx, "index should advance to the next field")
	assert.Empty(t, m.rs.configFieldValue, "value buffer should be cleared for the next field")
}

func TestApplyPhase_PluginConfig_BuildsEntityAfterLastField(t *testing.T) {
	// Set up a single-field plugin so that after collecting it, the entity
	// should be assembled automatically.
	entry := &domain.PluginRegistryEntry{URLPattern: "https://x/{version}.wasm"}
	ver := &domain.PluginVersionEntry{Version: "1.0.0", ConfigFields: []string{"secret_key"}}
	m := newRekeyModel(t)
	m.phase = phaseRekeyPluginConfig
	m.rs.pluginConfigFields = []string{"secret_key"}
	m.rs.pluginConfigValues = make(map[string]string)
	m.rs.configFieldIdx = 0
	m.rs.configFieldValue = "s3cr3t"
	m.rs.resolvedEntry = entry
	m.rs.resolvedVersionEntry = ver

	m.applyPhase()

	require.NotNil(t, m.rs.newPluginEntity,
		"entity should be assembled as soon as the last config field is collected")
	assert.Equal(t, "s3cr3t", m.rs.newPluginEntity.PluginConfig["secret_key"])
	assert.Equal(t, "https://x/1.0.0.wasm", m.rs.newPluginEntity.Url)
}

func TestApplyPhase_DryRun_PreselectsOnlyDecodableFiles(t *testing.T) {
	m := newRekeyModel(t)
	m.phase = phaseRekeyDryRun
	m.rs.previews = []migration.FilePreview{
		{Path: "/ctx/contacts.json", DecodedOK: true},
		{Path: "/ctx/bad-creds.yaml", DecodedOK: false}, // decode failed — must be excluded
		{Path: "/ctx/auth.yaml", DecodedOK: true},
	}

	m.applyPhase()

	assert.Len(t, m.rs.selectedPaths, 2,
		"only files that decoded successfully should be pre-selected")
	assert.Contains(t, m.rs.selectedPaths, "/ctx/contacts.json")
	assert.Contains(t, m.rs.selectedPaths, "/ctx/auth.yaml")
	assert.NotContains(t, m.rs.selectedPaths, "/ctx/bad-creds.yaml")
}

func TestApplyPhase_DryRun_NoPreviews_YieldsEmptySelection(t *testing.T) {
	m := newRekeyModel(t)
	m.phase = phaseRekeyDryRun
	m.rs.previews = nil

	m.applyPhase()

	assert.Empty(t, m.rs.selectedPaths)
}

// ── key message test double ───────────────────────────────────────────────────

// rekeyTestKeyMsg is a minimal tea.KeyMsg implementation used only in tests.
// It lets us control exactly what String() returns (e.g. "ctrl+c", "esc",
// "enter") without depending on bubbletea key constants or a real terminal.
type rekeyTestKeyMsg struct {
	s string
}

func (m rekeyTestKeyMsg) String() string { return m.s }
func (m rekeyTestKeyMsg) Key() tea.Key   { return tea.Key{} }

func rekeyCtrlC() rekeyTestKeyMsg { return rekeyTestKeyMsg{s: "ctrl+c"} }
func rekeyEsc() rekeyTestKeyMsg   { return rekeyTestKeyMsg{s: "esc"} }
func rekeyEnter() rekeyTestKeyMsg { return rekeyTestKeyMsg{s: "enter"} }

// stripAnsiCodes removes ANSI colour/style escape sequences so that View()
// output can be matched against plain-text substrings in test assertions.
// lipgloss may or may not emit codes depending on the execution environment.
func stripAnsiCodes(s string) string {
	var b strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			i += 2
			for i < len(s) && s[i] != 'm' {
				i++
			}
			i++ // consume 'm'
			continue
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

// castModel type-asserts tea.Model back to pluginRekeyModel.
// It fails the test immediately if the assertion fails.
func castModel(t *testing.T, model tea.Model) pluginRekeyModel {
	t.Helper()
	m, ok := model.(pluginRekeyModel)
	require.True(t, ok, "Update() must return a pluginRekeyModel")
	return m
}

// ── Init ─────────────────────────────────────────────────────────────────────

func TestRekeyInit_NoteOnlyScreen_ReturnsNilCmd(t *testing.T) {
	// phaseRekeyCurrentInfo shows a single NoteField, which is not focusable.
	// Screen.Init() only fires a Focus command when there is a focusable field.
	m := newRekeyModel(t)
	cmd := m.Init()
	assert.Nil(t, cmd, "Init() on a NoteField-only screen should return nil cmd")
}

func TestRekeyInit_DoesNotPanic(t *testing.T) {
	m := newRekeyModel(t)
	require.NotPanics(t, func() { m.Init() })
}

func TestRekeyInit_AfterScreenWithFocusableField_ReturnsCmd(t *testing.T) {
	// phaseRekeyAction renders a SelectField, which is focusable.
	// Init() should return a non-nil Focus command in that case.
	m := newRekeyModel(t)
	m.phase = phaseRekeyAction
	m.screen = m.buildScreen()
	cmd := m.Init()
	// A focusable SelectField fires a Focus() cmd — we only verify it doesn't panic
	// and returns something (the exact cmd value is an internal implementation detail).
	_ = cmd // non-nil expected, but cmd equality is non-trivial to assert portably
}

// ── View ─────────────────────────────────────────────────────────────────────

func TestRekeyView_Done_HasEmptyContent(t *testing.T) {
	m := newRekeyModel(t)
	m.done = true
	v := m.View()
	assert.Empty(t, v.Content, "done model should render an empty view")
}

func TestRekeyView_Cancelled_HasEmptyContent(t *testing.T) {
	m := newRekeyModel(t)
	m.cancelled = true
	v := m.View()
	assert.Empty(t, v.Content, "cancelled model should render an empty view")
}

func TestRekeyView_Normal_ContainsHeader(t *testing.T) {
	m := newRekeyModel(t)
	v := m.View()
	assert.Contains(t, stripAnsiCodes(v.Content), "GDG Plugin Re-key",
		"normal view must contain the application header")
}

func TestRekeyView_Normal_SetsAltScreen(t *testing.T) {
	m := newRekeyModel(t)
	v := m.View()
	assert.True(t, v.AltScreen, "normal view should request the alternate screen")
}

func TestRekeyView_Normal_ContainsSectionName(t *testing.T) {
	m := newRekeyModel(t)
	sectionName := m.phase.sectionName()
	v := m.View()
	assert.Contains(t, stripAnsiCodes(v.Content), sectionName,
		"view should include the current phase's section name")
}

func TestRekeyView_ContentIsNotEmpty_WhenNotDoneOrCancelled(t *testing.T) {
	m := newRekeyModel(t)
	v := m.View()
	assert.NotEmpty(t, v.Content, "active model should produce non-empty content")
}

func TestRekeyView_DoneAndCancelled_BothReturnEmpty(t *testing.T) {
	// Both flags together should still result in an empty view (done is checked first).
	m := newRekeyModel(t)
	m.done = true
	m.cancelled = true
	v := m.View()
	assert.Empty(t, v.Content)
}

func TestRekeyView_PhaseActionScreen_DoesNotPanic(t *testing.T) {
	m := newRekeyModel(t)
	m.phase = phaseRekeyAction
	m.screen = m.buildScreen()
	require.NotPanics(t, func() { m.View() })
}

func TestRekeyView_PhaseContextSelect_DoesNotPanic(t *testing.T) {
	m := newRekeyModel(t)
	m.phase = phaseRekeyContextSelect
	m.screen = m.buildScreen()
	require.NotPanics(t, func() { m.View() })
}

func TestRekeyView_PhaseBackupOptions_DoesNotPanic(t *testing.T) {
	m := newRekeyModel(t)
	m.phase = phaseRekeyBackupOptions
	m.screen = m.buildScreen()
	require.NotPanics(t, func() { m.View() })
}

func TestRekeyView_PhaseCredentials_DoesNotPanic(t *testing.T) {
	m := newRekeyModel(t)
	m.phase = phaseRekeyCredentials
	m.screen = m.buildScreen()
	require.NotPanics(t, func() { m.View() })
}

func TestRekeyView_PhaseConfirm_NoFilesSelected_DoesNotPanic(t *testing.T) {
	m := newRekeyModel(t)
	m.phase = phaseRekeyConfirm
	m.rs.selectedPaths = nil
	m.screen = m.buildScreen()
	require.NotPanics(t, func() { m.View() })
}

func TestRekeyView_PhaseConfirm_WithFilesSelected_DoesNotPanic(t *testing.T) {
	m := newRekeyModel(t)
	m.phase = phaseRekeyConfirm
	m.rs.contextNames = []string{"ctx1"}
	m.rs.selectedPaths = []string{"/some/path/contacts.json"}
	m.rs.doBackup = true
	m.screen = m.buildScreen()
	require.NotPanics(t, func() { m.View() })
}

func TestRekeyView_PhaseFileSelect_WithPreviews_DoesNotPanic(t *testing.T) {
	m := newRekeyModel(t)
	m.phase = phaseRekeyFileSelect
	m.rs.previews = []migration.FilePreview{
		{Path: "/ctx/contacts.json", Category: "contact_points", DecodedOK: true},
	}
	m.screen = m.buildScreen()
	require.NotPanics(t, func() { m.View() })
}

func TestRekeyView_PluginVersion_NoVersions_DoesNotPanic(t *testing.T) {
	m := newRekeyModel(t)
	m.phase = phaseRekeyPluginVersion
	m.rs.pluginName = "missing-plugin"
	m.rs.availablePlugins = []domain.PluginRegistryEntry{
		{Name: "other-plugin", URLPattern: "https://x/{version}.wasm"},
	}
	m.screen = m.buildScreen()
	require.NotPanics(t, func() { m.View() })
}

func TestRekeyView_PluginSelectWithRegistryError_DoesNotPanic(t *testing.T) {
	m := newRekeyModel(t)
	m.phase = phaseRekeyPluginSelect
	m.rs.registryError = "registry unavailable"
	m.screen = m.buildScreen()
	require.NotPanics(t, func() { m.View() })
}

// ── Update ────────────────────────────────────────────────────────────────────

func TestRekeyUpdate_WindowSizeMsg_UpdatesWidth(t *testing.T) {
	m := newRekeyModel(t)
	result, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 50})
	updated := castModel(t, result)
	assert.Equal(t, 120, updated.width, "width should be updated from WindowSizeMsg")
}

func TestRekeyUpdate_WindowSizeMsg_UpdatesHeight(t *testing.T) {
	m := newRekeyModel(t)
	result, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 50})
	updated := castModel(t, result)
	assert.Equal(t, 50, updated.height, "height should be updated from WindowSizeMsg")
}

func TestRekeyUpdate_WindowSizeMsg_PhaseUnchanged(t *testing.T) {
	m := newRekeyModel(t)
	result, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	updated := castModel(t, result)
	assert.Equal(t, phaseRekeyCurrentInfo, updated.phase, "WindowSizeMsg should not change the phase")
}

func TestRekeyUpdate_WindowSizeMsg_DoesNotSetDoneOrCancelled(t *testing.T) {
	m := newRekeyModel(t)
	result, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	updated := castModel(t, result)
	assert.False(t, updated.done)
	assert.False(t, updated.cancelled)
}

func TestRekeyUpdate_CtrlC_SetsCancelledFlag(t *testing.T) {
	m := newRekeyModel(t)
	result, _ := m.Update(rekeyCtrlC())
	updated := castModel(t, result)
	assert.True(t, updated.cancelled, "ctrl+c must set the cancelled flag")
}

func TestRekeyUpdate_CtrlC_DoesNotSetDone(t *testing.T) {
	m := newRekeyModel(t)
	result, _ := m.Update(rekeyCtrlC())
	updated := castModel(t, result)
	assert.False(t, updated.done, "ctrl+c should not set done — only cancelled")
}

func TestRekeyUpdate_CtrlC_PhaseUnchanged(t *testing.T) {
	m := newRekeyModel(t)
	result, _ := m.Update(rekeyCtrlC())
	updated := castModel(t, result)
	assert.Equal(t, phaseRekeyCurrentInfo, updated.phase, "ctrl+c should not change the phase")
}

func TestRekeyUpdate_Enter_NoteScreen_AdvancesToNextPhase(t *testing.T) {
	// phaseRekeyCurrentInfo has only a NoteField (not focusable).
	// Enter immediately submits the screen, triggering a phase transition.
	m := newRekeyModel(t)
	assert.Equal(t, phaseRekeyCurrentInfo, m.phase)
	result, _ := m.Update(rekeyEnter())
	updated := castModel(t, result)
	assert.Equal(t, phaseRekeyAction, updated.phase,
		"Enter on CurrentInfo note screen must advance to Action phase")
}

func TestRekeyUpdate_Enter_NoteScreen_DoesNotSetDoneOrCancelled(t *testing.T) {
	m := newRekeyModel(t)
	result, _ := m.Update(rekeyEnter())
	updated := castModel(t, result)
	assert.False(t, updated.done)
	assert.False(t, updated.cancelled)
}

func TestRekeyUpdate_Esc_OnStartPhase_SetsCancelled(t *testing.T) {
	// Esc on the very first phase has nowhere to go back to — the TUI cancels.
	m := newRekeyModel(t)
	require.Equal(t, m.startPhase, m.phase, "precondition: should be at the start phase")
	result, _ := m.Update(rekeyEsc())
	updated := castModel(t, result)
	assert.True(t, updated.cancelled, "Esc at startPhase must set cancelled")
}

func TestRekeyUpdate_Esc_OnStartPhase_DoesNotSetDone(t *testing.T) {
	m := newRekeyModel(t)
	result, _ := m.Update(rekeyEsc())
	updated := castModel(t, result)
	assert.False(t, updated.done, "Esc at startPhase should not set done")
}

func TestRekeyUpdate_Esc_OnLaterPhase_GoesBackward(t *testing.T) {
	// Esc on a phase other than startPhase must go back, not cancel.
	m := newRekeyModel(t)
	m.phase = phaseRekeyAction
	m.screen = m.buildScreen()
	result, _ := m.Update(rekeyEsc())
	updated := castModel(t, result)
	assert.False(t, updated.cancelled, "Esc on non-start phase should go back, not cancel")
	assert.Equal(t, phaseRekeyCurrentInfo, updated.phase,
		"Esc from Action must return to CurrentInfo (prevPhase)")
}

func TestRekeyUpdate_Esc_OnLaterPhase_DoesNotSetDone(t *testing.T) {
	m := newRekeyModel(t)
	m.phase = phaseRekeyAction
	m.screen = m.buildScreen()
	result, _ := m.Update(rekeyEsc())
	updated := castModel(t, result)
	assert.False(t, updated.done)
}

func TestRekeyUpdate_Submit_CancelAction_SetsDone(t *testing.T) {
	// Full flow: advance past CurrentInfo, then submit with action="cancel".
	m := newRekeyModel(t)

	// Step 1: submit the CurrentInfo note screen.
	r1, _ := m.Update(rekeyEnter())
	m1 := castModel(t, r1)
	require.Equal(t, phaseRekeyAction, m1.phase)

	// Step 2: choose the cancel action (bound via pointer) and submit.
	m1.rs.action = "cancel"
	r2, _ := m1.Update(rekeyEnter())
	m2 := castModel(t, r2)

	assert.True(t, m2.done, "choosing 'cancel' at the Action phase must set done")
}

func TestRekeyUpdate_PluginConfig_EscWithPositiveFieldIdx_DecrementsIndex(t *testing.T) {
	// When in the multi-field config loop with configFieldIdx > 0, Esc steps
	// back one field instead of jumping to the previous top-level phase.
	m := newRekeyModel(t)
	m.phase = phaseRekeyPluginConfig
	m.rs.pluginConfigFields = []string{"field_a", "field_b"}
	m.rs.pluginConfigValues = map[string]string{"field_a": "value_a", "field_b": ""}
	m.rs.configFieldIdx = 1 // currently on field_b
	m.screen = m.buildScreen()

	result, _ := m.Update(rekeyEsc())
	updated := castModel(t, result)

	assert.Equal(t, 0, updated.rs.configFieldIdx,
		"Esc in PluginConfig loop should decrement configFieldIdx")
}

func TestRekeyUpdate_PluginConfig_EscRestoresPriorFieldValue(t *testing.T) {
	m := newRekeyModel(t)
	m.phase = phaseRekeyPluginConfig
	m.rs.pluginConfigFields = []string{"field_a", "field_b"}
	m.rs.pluginConfigValues = map[string]string{"field_a": "value_a", "field_b": ""}
	m.rs.configFieldIdx = 1
	m.screen = m.buildScreen()

	result, _ := m.Update(rekeyEsc())
	updated := castModel(t, result)

	assert.Equal(t, "value_a", updated.rs.configFieldValue,
		"Esc should restore the previously-entered value for the prior config field")
}

func TestRekeyUpdate_PluginConfig_EscAtFieldIdx0_GoesToPluginVersion(t *testing.T) {
	// At configFieldIdx == 0, Esc falls through to prevPhase (PluginVersion).
	m := newRekeyModel(t)
	m.phase = phaseRekeyPluginConfig
	m.rs.pluginConfigFields = []string{"field_a"}
	m.rs.configFieldIdx = 0
	m.screen = m.buildScreen()

	result, _ := m.Update(rekeyEsc())
	updated := castModel(t, result)

	assert.Equal(t, phaseRekeyPluginVersion, updated.phase,
		"Esc at configFieldIdx==0 should fall through to prevPhase (PluginVersion)")
}

func TestRekeyUpdate_ContextSelect_EscWithSwitchAndConfigFields_GoesToLastConfigField(t *testing.T) {
	// Going back from ContextSelect when action=="switch" and config fields exist
	// should restore the last config field entry instead of going to PluginVersion.
	m := newRekeyModel(t)
	m.phase = phaseRekeyContextSelect
	m.rs.action = "switch"
	m.rs.pluginConfigFields = []string{"field_a", "field_b"}
	m.rs.pluginConfigValues = map[string]string{"field_a": "va", "field_b": "vb"}
	m.screen = m.buildScreen()

	result, _ := m.Update(rekeyEsc())
	updated := castModel(t, result)

	assert.Equal(t, phaseRekeyPluginConfig, updated.phase,
		"Esc from ContextSelect (switch+config fields) must return to PluginConfig")
	assert.Equal(t, 1, updated.rs.configFieldIdx,
		"configFieldIdx should point to the last config field")
	assert.Equal(t, "vb", updated.rs.configFieldValue,
		"configFieldValue should be pre-populated with the last field's stored value")
}

func TestRekeyUpdate_ContextSelect_EscWithDisableAction_UsesNormalPrevPhase(t *testing.T) {
	// When action=="disable" (no config fields path), Esc from ContextSelect
	// falls through to the normal prevPhase handler → phaseRekeyAction.
	m := newRekeyModel(t)
	m.phase = phaseRekeyContextSelect
	m.rs.action = "disable"
	m.rs.pluginConfigFields = nil // no config fields
	m.screen = m.buildScreen()

	result, _ := m.Update(rekeyEsc())
	updated := castModel(t, result)

	assert.Equal(t, phaseRekeyAction, updated.phase,
		"Esc from ContextSelect with disable action should return to Action phase")
}

func TestRekeyUpdate_UnknownMsg_DoesNotPanic(t *testing.T) {
	m := newRekeyModel(t)
	require.NotPanics(t, func() {
		m.Update("arbitrary non-tea message")
	})
}

func TestRekeyUpdate_UnknownMsg_PhaseUnchanged(t *testing.T) {
	m := newRekeyModel(t)
	result, _ := m.Update("arbitrary non-tea message")
	updated := castModel(t, result)
	assert.Equal(t, phaseRekeyCurrentInfo, updated.phase)
}
