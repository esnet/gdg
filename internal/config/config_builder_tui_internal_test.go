// config_builder_tui_internal_test.go exercises the pure-logic helpers and
// state-machine transitions defined in config_builder_tui.go.
//
// These tests use white-box access (package config) to reach unexported types
// and functions without requiring a real terminal.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/esnet/gdg/internal/adapter/plugins/secure/noop"
	"github.com/esnet/gdg/internal/config/config_domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── test helpers ──────────────────────────────────────────────────────────────

// newTUITestApp writes a minimal gdg.yml to a temp dir and returns the parsed app.
func newTUITestApp(t *testing.T) *config_domain.GDGAppConfiguration {
	t.Helper()
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "gdg-tui-test.yml")
	yaml := fmt.Sprintf(`
context_name: tui-test
contexts:
  tui-test:
    url: http://localhost:3000
    output_path: %s
    watched:
      - General
storage_engine: {}
`, dir)
	require.NoError(t, os.WriteFile(cfgPath, []byte(yaml), 0o600))
	return NewConfig(cfgPath)
}

// newTUIModel builds a configBuilderModel with a noop encoder and no registry
// client — sufficient for all tests that do not touch plugin-registry screens.
func newTUIModel(t *testing.T) configBuilderModel {
	t.Helper()
	app := newTUITestApp(t)
	return newConfigBuilderModel(app, "tui-test", noop.NoOpEncoder{}, nil)
}

// ── builderPhase.sectionName ──────────────────────────────────────────────────

func TestSectionName_AuthPhases(t *testing.T) {
	for _, p := range []builderPhase{phaseAuthType, phaseBasicCreds, phaseTokenCreds} {
		assert.Equal(t, "Authentication", p.sectionName(),
			"phase %d should map to Authentication", p)
	}
}

func TestSectionName_ServerSettings(t *testing.T) {
	assert.Equal(t, "Server Settings", phaseServerSettings.sectionName())
}

func TestSectionName_UserSettings(t *testing.T) {
	assert.Equal(t, "User Settings", phaseUserSettings.sectionName())
	assert.Equal(t, "User Settings", phaseUserLengths.sectionName())
}

func TestSectionName_FolderPhases(t *testing.T) {
	folderPhases := []builderPhase{
		phaseFolderScope, phaseFolderMenu, phaseFolderAdd,
		phaseFolderTest, phaseFolderTestResult,
	}
	for _, p := range folderPhases {
		assert.Equal(t, "Watched Folders", p.sectionName(),
			"phase %d should map to Watched Folders", p)
	}
}

func TestSectionName_ConnectionPhases(t *testing.T) {
	connPhases := []builderPhase{
		phaseConnectionToggle, phaseFilterToggle, phaseFilterInput,
		phaseDefaultCreds, phaseCredRuleToggle, phaseCredRuleFile,
		phaseCredRuleMatcher, phaseCredRuleCreds,
	}
	for _, p := range connPhases {
		assert.Equal(t, "Connection Settings", p.sectionName(),
			"phase %d should map to Connection Settings", p)
	}
}

func TestSectionName_StoragePhases(t *testing.T) {
	storagePhases := []builderPhase{
		phaseStorageToggle, phaseStorageProvider, phaseStorageProviderInfo,
		phaseStorageCustomConfig, phaseStorageCustomCreds,
		phaseStorageCustomOptions, phaseStorageAssign,
	}
	for _, p := range storagePhases {
		assert.Equal(t, "Cloud Storage", p.sectionName(),
			"phase %d should map to Cloud Storage", p)
	}
}

func TestSectionName_PluginPhases(t *testing.T) {
	pluginPhases := []builderPhase{
		phasePluginToggle, phasePluginSelect, phasePluginVersion, phasePluginConfig,
	}
	for _, p := range pluginPhases {
		assert.Equal(t, "Cipher Plugin", p.sectionName(),
			"phase %d should map to Cipher Plugin", p)
	}
}

func TestSectionName_DonePhase(t *testing.T) {
	assert.Equal(t, "", phaseDone.sectionName())
}

// ── Layout helpers ────────────────────────────────────────────────────────────

func TestLeftWidth_HalfOfWidth(t *testing.T) {
	m := newTUIModel(t)
	m.width = 100
	assert.Equal(t, 50, m.leftWidth())
}

func TestLeftWidth_OddWidth(t *testing.T) {
	m := newTUIModel(t)
	m.width = 101
	assert.Equal(t, 50, m.leftWidth())
}

func TestLeftWidth_EnforcesMinimum(t *testing.T) {
	m := newTUIModel(t)
	m.width = 10 // half would be 5, below minimum of 30
	assert.Equal(t, 30, m.leftWidth())
}

func TestLeftWidth_ExactlyMinimumThreshold(t *testing.T) {
	m := newTUIModel(t)
	m.width = 60 // half = 30 exactly; should not trigger the clamp
	assert.Equal(t, 30, m.leftWidth())
}

func TestRightWidth_ComplementsLeft(t *testing.T) {
	m := newTUIModel(t)
	m.width = 100
	assert.Equal(t, m.width-m.leftWidth(), m.rightWidth())
}

func TestRightWidth_NarrowWindow(t *testing.T) {
	m := newTUIModel(t)
	m.width = 40
	// leftWidth clamps to 30; rightWidth = 40 - 30 = 10
	assert.Equal(t, 10, m.rightWidth())
}

func TestBodyHeight_NormalHeight(t *testing.T) {
	m := newTUIModel(t)
	m.height = 40
	expected := 40 - tuiHeaderHeight - tuiFooterHeight
	assert.Equal(t, expected, m.bodyHeight())
}

func TestBodyHeight_EnforcesMinimum(t *testing.T) {
	m := newTUIModel(t)
	m.height = 5 // would produce negative body; minimum is tuiMinBodyH
	assert.Equal(t, tuiMinBodyH, m.bodyHeight())
}

func TestBodyHeight_ExactMinimum(t *testing.T) {
	m := newTUIModel(t)
	m.height = tuiHeaderHeight + tuiFooterHeight + tuiMinBodyH
	assert.Equal(t, tuiMinBodyH, m.bodyHeight())
}

func TestBodyHeight_JustAboveMinimum(t *testing.T) {
	m := newTUIModel(t)
	m.height = tuiHeaderHeight + tuiFooterHeight + tuiMinBodyH + 1
	assert.Equal(t, tuiMinBodyH+1, m.bodyHeight())
}

// ── validatePositiveInt ───────────────────────────────────────────────────────

func TestValidatePositiveInt_Valid(t *testing.T) {
	require.NoError(t, validatePositiveInt("1"))
	require.NoError(t, validatePositiveInt("8"))
	require.NoError(t, validatePositiveInt("20"))
	require.NoError(t, validatePositiveInt("100"))
}

func TestValidatePositiveInt_ValidWithWhitespace(t *testing.T) {
	require.NoError(t, validatePositiveInt("  12  "))
}

func TestValidatePositiveInt_Zero(t *testing.T) {
	require.Error(t, validatePositiveInt("0"))
}

func TestValidatePositiveInt_Negative(t *testing.T) {
	require.Error(t, validatePositiveInt("-1"))
	require.Error(t, validatePositiveInt("-100"))
}

func TestValidatePositiveInt_NonNumeric(t *testing.T) {
	require.Error(t, validatePositiveInt("abc"))
	require.Error(t, validatePositiveInt(""))
	require.Error(t, validatePositiveInt("   "))
}

func TestValidatePositiveInt_Float(t *testing.T) {
	require.Error(t, validatePositiveInt("1.5"))
}

// ── newConfigBuilderModel — startPhase logic ──────────────────────────────────

func TestNewConfigBuilderModel_NoPlugin_StartsAtPluginToggle(t *testing.T) {
	// Default app has plugins.disabled not set and no cipher_plugin, so
	// startPhase should be phasePluginToggle (wizard begins with plugin question).
	app := newTUITestApp(t)
	m := newConfigBuilderModel(app, "tui-test", noop.NoOpEncoder{}, nil)
	assert.Equal(t, phasePluginToggle, m.phase)
	assert.Equal(t, phasePluginToggle, m.startPhase)
}

func TestNewConfigBuilderModel_WithActivePlugin_StartsAtAuthType(t *testing.T) {
	// If a cipher plugin is already configured and not disabled, the wizard
	// skips the plugin setup and starts at auth configuration.
	app := newTUITestApp(t)
	app.PluginConfig.Disabled = false
	app.PluginConfig.CipherPlugin = &config_domain.PluginEntity{
		Url: "https://example.com/plugin.wasm",
	}
	m := newConfigBuilderModel(app, "tui-test", noop.NoOpEncoder{}, nil)
	assert.Equal(t, phaseAuthType, m.phase)
	assert.Equal(t, phaseAuthType, m.startPhase)
}

func TestNewConfigBuilderModel_InitialisesBuilderState(t *testing.T) {
	m := newTUIModel(t)
	require.NotNil(t, m.bs)
	assert.Equal(t, "tui-test", m.bs.contextName)
	require.NotNil(t, m.bs.config)
	require.NotNil(t, m.bs.secure)
	assert.Equal(t, "Main Org.", m.bs.config.OrganizationName)
	require.NotNil(t, m.bs.config.ConnectionSettings)
}

// ── nextPhase — forward transitions ──────────────────────────────────────────

func TestNextPhase_PluginToggle_SkipPlugin(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phasePluginToggle
	m.bs.configurePlugin = false
	assert.Equal(t, phaseAuthType, m.nextPhase())
}

func TestNextPhase_PluginToggle_ConfigurePlugin(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phasePluginToggle
	m.bs.configurePlugin = true
	assert.Equal(t, phasePluginSelect, m.nextPhase())
}

func TestNextPhase_PluginSelect_LoadError(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phasePluginSelect
	m.bs.pluginLoadErr = true
	assert.Equal(t, phaseAuthType, m.nextPhase())
}

func TestNextPhase_PluginSelect_Success(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phasePluginSelect
	m.bs.pluginLoadErr = false
	assert.Equal(t, phasePluginVersion, m.nextPhase())
}

func TestNextPhase_PluginVersion_NoConfigFields(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phasePluginVersion
	m.bs.pluginConfigFields = nil
	assert.Equal(t, phaseAuthType, m.nextPhase())
}

func TestNextPhase_PluginVersion_HasConfigFields(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phasePluginVersion
	m.bs.pluginConfigFields = []string{"key"}
	assert.Equal(t, phasePluginConfig, m.nextPhase())
}

func TestNextPhase_PluginConfig_MoreFields(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phasePluginConfig
	m.bs.pluginConfigFields = []string{"key1", "key2"}
	m.bs.pluginConfigIdx = 1 // still one field left (idx 1 < len 2)
	assert.Equal(t, phasePluginConfig, m.nextPhase())
}

func TestNextPhase_PluginConfig_LastField(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phasePluginConfig
	m.bs.pluginConfigFields = []string{"key1", "key2"}
	m.bs.pluginConfigIdx = 2 // all fields done
	assert.Equal(t, phaseAuthType, m.nextPhase())
}

func TestNextPhase_AuthType_Basic(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseAuthType
	m.bs.authType = basicAuthForm.String()
	assert.Equal(t, phaseBasicCreds, m.nextPhase())
}

func TestNextPhase_AuthType_Token(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseAuthType
	m.bs.authType = tokenAuthForm.String()
	assert.Equal(t, phaseTokenCreds, m.nextPhase())
}

func TestNextPhase_AuthType_Both(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseAuthType
	m.bs.authType = bothAuthForm.String()
	assert.Equal(t, phaseBasicCreds, m.nextPhase())
}

func TestNextPhase_BasicCreds_BothAuth(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseBasicCreds
	m.bs.authType = bothAuthForm.String()
	assert.Equal(t, phaseTokenCreds, m.nextPhase())
}

func TestNextPhase_BasicCreds_BasicOnly(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseBasicCreds
	m.bs.authType = basicAuthForm.String()
	assert.Equal(t, phaseServerSettings, m.nextPhase())
}

func TestNextPhase_TokenCreds(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseTokenCreds
	assert.Equal(t, phaseServerSettings, m.nextPhase())
}

func TestNextPhase_ServerSettings(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseServerSettings
	assert.Equal(t, phaseUserSettings, m.nextPhase())
}

func TestNextPhase_UserSettings_RandomPassword(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseUserSettings
	m.bs.userRandomPassword = true
	assert.Equal(t, phaseUserLengths, m.nextPhase())
}

func TestNextPhase_UserSettings_NoRandomPassword(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseUserSettings
	m.bs.userRandomPassword = false
	assert.Equal(t, phaseFolderScope, m.nextPhase())
}

func TestNextPhase_UserLengths(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseUserLengths
	assert.Equal(t, phaseFolderScope, m.nextPhase())
}

func TestNextPhase_FolderScope_All(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseFolderScope
	m.bs.folderScope = "all"
	assert.Equal(t, phaseConnectionToggle, m.nextPhase())
}

func TestNextPhase_FolderScope_Allowlist(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseFolderScope
	m.bs.folderScope = "allowlist"
	assert.Equal(t, phaseFolderMenu, m.nextPhase())
}

func TestNextPhase_FolderMenu_Add(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseFolderMenu
	m.bs.folderAction = "add"
	assert.Equal(t, phaseFolderAdd, m.nextPhase())
}

func TestNextPhase_FolderMenu_Test(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseFolderMenu
	m.bs.folderAction = "test"
	assert.Equal(t, phaseFolderTest, m.nextPhase())
}

func TestNextPhase_FolderMenu_Done(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseFolderMenu
	m.bs.folderAction = "done"
	assert.Equal(t, phaseConnectionToggle, m.nextPhase())
}

func TestNextPhase_FolderAdd(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseFolderAdd
	assert.Equal(t, phaseFolderMenu, m.nextPhase())
}

func TestNextPhase_FolderTest(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseFolderTest
	assert.Equal(t, phaseFolderTestResult, m.nextPhase())
}

func TestNextPhase_FolderTestResult_TestMore(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseFolderTestResult
	m.bs.folderTestMore = true
	assert.Equal(t, phaseFolderTest, m.nextPhase())
}

func TestNextPhase_FolderTestResult_Done(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseFolderTestResult
	m.bs.folderTestMore = false
	assert.Equal(t, phaseFolderMenu, m.nextPhase())
}

func TestNextPhase_ConnectionToggle_Configure(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseConnectionToggle
	m.bs.configureConnections = true
	assert.Equal(t, phaseFilterToggle, m.nextPhase())
}

func TestNextPhase_ConnectionToggle_Skip(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseConnectionToggle
	m.bs.configureConnections = false
	assert.Equal(t, phaseStorageToggle, m.nextPhase())
}

func TestNextPhase_FilterToggle_AddFilters(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseFilterToggle
	m.bs.addFilters = true
	assert.Equal(t, phaseFilterInput, m.nextPhase())
}

func TestNextPhase_FilterToggle_Skip(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseFilterToggle
	m.bs.addFilters = false
	assert.Equal(t, phaseDefaultCreds, m.nextPhase())
}

func TestNextPhase_FilterInput_AddMore(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseFilterInput
	m.bs.addMoreFilters = true
	assert.Equal(t, phaseFilterInput, m.nextPhase())
}

func TestNextPhase_FilterInput_Done(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseFilterInput
	m.bs.addMoreFilters = false
	assert.Equal(t, phaseDefaultCreds, m.nextPhase())
}

func TestNextPhase_DefaultCreds(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseDefaultCreds
	assert.Equal(t, phaseCredRuleToggle, m.nextPhase())
}

func TestNextPhase_CredRuleToggle_AddRules(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseCredRuleToggle
	m.bs.addCredRules = true
	assert.Equal(t, phaseCredRuleFile, m.nextPhase())
}

func TestNextPhase_CredRuleToggle_Skip(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseCredRuleToggle
	m.bs.addCredRules = false
	assert.Equal(t, phaseStorageToggle, m.nextPhase())
}

func TestNextPhase_CredRuleFile(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseCredRuleFile
	assert.Equal(t, phaseCredRuleMatcher, m.nextPhase())
}

func TestNextPhase_CredRuleMatcher_AddMoreMatchers(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseCredRuleMatcher
	m.bs.credAddMoreMatcher = true
	assert.Equal(t, phaseCredRuleMatcher, m.nextPhase())
}

func TestNextPhase_CredRuleMatcher_Done(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseCredRuleMatcher
	m.bs.credAddMoreMatcher = false
	assert.Equal(t, phaseCredRuleCreds, m.nextPhase())
}

func TestNextPhase_CredRuleCreds_AddMoreRules(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseCredRuleCreds
	m.bs.addMoreCredRules = true
	assert.Equal(t, phaseCredRuleFile, m.nextPhase())
}

func TestNextPhase_CredRuleCreds_Done(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseCredRuleCreds
	m.bs.addMoreCredRules = false
	assert.Equal(t, phaseStorageToggle, m.nextPhase())
}

func TestNextPhase_StorageToggle_Configure(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseStorageToggle
	m.bs.configureStorage = true
	assert.Equal(t, phaseStorageProvider, m.nextPhase())
}

func TestNextPhase_StorageToggle_Skip(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseStorageToggle
	m.bs.configureStorage = false
	assert.Equal(t, phaseDone, m.nextPhase())
}

func TestNextPhase_StorageProvider_Custom(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseStorageProvider
	m.bs.storageProvider = string(providerCustom)
	assert.Equal(t, phaseStorageCustomConfig, m.nextPhase())
}

func TestNextPhase_StorageProvider_Managed(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseStorageProvider
	m.bs.storageProvider = string(providerAWS)
	assert.Equal(t, phaseStorageProviderInfo, m.nextPhase())
}

func TestNextPhase_StorageProviderInfo(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseStorageProviderInfo
	assert.Equal(t, phaseDone, m.nextPhase())
}

func TestNextPhase_StorageCustomSequence(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseStorageCustomConfig
	assert.Equal(t, phaseStorageCustomCreds, m.nextPhase())

	m.phase = phaseStorageCustomCreds
	assert.Equal(t, phaseStorageCustomOptions, m.nextPhase())

	m.phase = phaseStorageCustomOptions
	assert.Equal(t, phaseStorageAssign, m.nextPhase())

	m.phase = phaseStorageAssign
	assert.Equal(t, phaseDone, m.nextPhase())
}

// ── prevPhase — reverse transitions ──────────────────────────────────────────

func TestPrevPhase_PluginSelect(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phasePluginSelect
	assert.Equal(t, phasePluginToggle, m.prevPhase())
}

func TestPrevPhase_PluginVersion(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phasePluginVersion
	assert.Equal(t, phasePluginSelect, m.prevPhase())
}

func TestPrevPhase_PluginConfig_FirstField(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phasePluginConfig
	m.bs.pluginConfigIdx = 0
	assert.Equal(t, phasePluginVersion, m.prevPhase())
}

func TestPrevPhase_AuthType_NoPlugin(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseAuthType
	m.bs.configurePlugin = false
	assert.Equal(t, phasePluginToggle, m.prevPhase())
}

func TestPrevPhase_AuthType_WithPlugin_NoFields(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseAuthType
	m.bs.configurePlugin = true
	m.bs.pluginLoadErr = false
	m.bs.pluginConfigFields = nil
	assert.Equal(t, phasePluginVersion, m.prevPhase())
}

func TestPrevPhase_AuthType_WithPlugin_LoadErr(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseAuthType
	m.bs.configurePlugin = true
	m.bs.pluginLoadErr = true
	assert.Equal(t, phasePluginSelect, m.prevPhase())
}

func TestPrevPhase_AuthType_WithPlugin_HasConfigFields(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseAuthType
	m.bs.configurePlugin = true
	m.bs.pluginLoadErr = false
	m.bs.pluginConfigFields = []string{"key1"}
	assert.Equal(t, phasePluginConfig, m.prevPhase())
}

func TestPrevPhase_BasicCreds(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseBasicCreds
	assert.Equal(t, phaseAuthType, m.prevPhase())
}

func TestPrevPhase_TokenCreds_BasicOnly(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseTokenCreds
	m.bs.authType = tokenAuthForm.String()
	assert.Equal(t, phaseAuthType, m.prevPhase())
}

func TestPrevPhase_TokenCreds_BothAuth(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseTokenCreds
	m.bs.authType = bothAuthForm.String()
	assert.Equal(t, phaseBasicCreds, m.prevPhase())
}

func TestPrevPhase_ServerSettings_Basic(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseServerSettings
	m.bs.authType = basicAuthForm.String()
	assert.Equal(t, phaseBasicCreds, m.prevPhase())
}

func TestPrevPhase_ServerSettings_Token(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseServerSettings
	m.bs.authType = tokenAuthForm.String()
	assert.Equal(t, phaseTokenCreds, m.prevPhase())
}

func TestPrevPhase_ServerSettings_Both(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseServerSettings
	m.bs.authType = bothAuthForm.String()
	assert.Equal(t, phaseTokenCreds, m.prevPhase())
}

func TestPrevPhase_UserSettings(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseUserSettings
	assert.Equal(t, phaseServerSettings, m.prevPhase())
}

func TestPrevPhase_UserLengths(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseUserLengths
	assert.Equal(t, phaseUserSettings, m.prevPhase())
}

func TestPrevPhase_FolderScope_RandomPassword(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseFolderScope
	m.bs.userRandomPassword = true
	assert.Equal(t, phaseUserLengths, m.prevPhase())
}

func TestPrevPhase_FolderScope_NoRandomPassword(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseFolderScope
	m.bs.userRandomPassword = false
	assert.Equal(t, phaseUserSettings, m.prevPhase())
}

func TestPrevPhase_FolderMenu(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseFolderMenu
	assert.Equal(t, phaseFolderScope, m.prevPhase())
}

func TestPrevPhase_FolderAdd(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseFolderAdd
	assert.Equal(t, phaseFolderMenu, m.prevPhase())
}

func TestPrevPhase_FolderTest(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseFolderTest
	assert.Equal(t, phaseFolderMenu, m.prevPhase())
}

func TestPrevPhase_FolderTestResult(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseFolderTestResult
	assert.Equal(t, phaseFolderTest, m.prevPhase())
}

func TestPrevPhase_ConnectionToggle_FolderScopeAll(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseConnectionToggle
	m.bs.folderScope = "all"
	assert.Equal(t, phaseFolderScope, m.prevPhase())
}

func TestPrevPhase_ConnectionToggle_FolderAllowlist(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseConnectionToggle
	m.bs.folderScope = "allowlist"
	assert.Equal(t, phaseFolderMenu, m.prevPhase())
}

func TestPrevPhase_FilterToggle(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseFilterToggle
	assert.Equal(t, phaseConnectionToggle, m.prevPhase())
}

func TestPrevPhase_FilterInput(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseFilterInput
	assert.Equal(t, phaseFilterToggle, m.prevPhase())
}

func TestPrevPhase_DefaultCreds_WithFilters(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseDefaultCreds
	m.bs.addFilters = true
	assert.Equal(t, phaseFilterInput, m.prevPhase())
}

func TestPrevPhase_DefaultCreds_NoFilters(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseDefaultCreds
	m.bs.addFilters = false
	assert.Equal(t, phaseFilterToggle, m.prevPhase())
}

func TestPrevPhase_CredRuleToggle(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseCredRuleToggle
	assert.Equal(t, phaseDefaultCreds, m.prevPhase())
}

func TestPrevPhase_CredRuleFile_FirstRule(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseCredRuleFile
	m.bs.credRules = nil
	assert.Equal(t, phaseCredRuleToggle, m.prevPhase())
}

func TestPrevPhase_CredRuleFile_SubsequentRule(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseCredRuleFile
	m.bs.credRules = []*config_domain.RegexMatchesList{
		{SecureData: "first.yaml"},
	}
	assert.Equal(t, phaseCredRuleCreds, m.prevPhase())
}

func TestPrevPhase_StorageToggle_NoConnections(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseStorageToggle
	m.bs.configureConnections = false
	assert.Equal(t, phaseConnectionToggle, m.prevPhase())
}

func TestPrevPhase_StorageToggle_ConnectionsWithCredRules(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseStorageToggle
	m.bs.configureConnections = true
	m.bs.addCredRules = true
	assert.Equal(t, phaseCredRuleCreds, m.prevPhase())
}

func TestPrevPhase_StorageToggle_ConnectionsNoCredRules(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseStorageToggle
	m.bs.configureConnections = true
	m.bs.addCredRules = false
	assert.Equal(t, phaseCredRuleToggle, m.prevPhase())
}

func TestPrevPhase_StorageProvider(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseStorageProvider
	assert.Equal(t, phaseStorageToggle, m.prevPhase())
}

func TestPrevPhase_StorageProviderInfo(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseStorageProviderInfo
	assert.Equal(t, phaseStorageProvider, m.prevPhase())
}

func TestPrevPhase_StorageCustomSequence(t *testing.T) {
	m := newTUIModel(t)

	m.phase = phaseStorageCustomConfig
	assert.Equal(t, phaseStorageProvider, m.prevPhase())

	m.phase = phaseStorageCustomCreds
	assert.Equal(t, phaseStorageCustomConfig, m.prevPhase())

	m.phase = phaseStorageCustomOptions
	assert.Equal(t, phaseStorageCustomCreds, m.prevPhase())

	m.phase = phaseStorageAssign
	assert.Equal(t, phaseStorageCustomOptions, m.prevPhase())
}

// ── applyPhase — state mutations ──────────────────────────────────────────────

func TestApplyPhase_BasicCreds(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseBasicCreds
	m.bs.userName = "admin"
	m.bs.password = "s3cr3t"
	m.applyPhase()
	assert.Equal(t, "admin", m.bs.config.UserName)
	assert.Equal(t, "s3cr3t", m.bs.secure.Password)
}

func TestApplyPhase_TokenCreds(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseTokenCreds
	m.bs.token = "glsa_abc123"
	m.applyPhase()
	assert.Equal(t, "glsa_abc123", m.bs.secure.Token)
}

func TestApplyPhase_ServerSettings(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseServerSettings
	m.bs.url = "http://grafana.example.com"
	m.bs.outputPath = "/tmp/backups"
	m.applyPhase()
	assert.Equal(t, "http://grafana.example.com", m.bs.config.URL)
	assert.Equal(t, "/tmp/backups", m.bs.config.OutputPath)
}

func TestApplyPhase_UserSettings_RandomPassword(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseUserSettings
	m.bs.userRandomPassword = true
	m.applyPhase()
	// When random password is true, UserSettings should NOT be set in this phase
	// (it's set in phaseUserLengths instead).
	assert.Nil(t, m.bs.config.UserSettings)
}

func TestApplyPhase_UserSettings_DeterministicPassword(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseUserSettings
	m.bs.userRandomPassword = false
	m.applyPhase()
	require.NotNil(t, m.bs.config.UserSettings)
	assert.False(t, m.bs.config.UserSettings.RandomPassword)
}

func TestApplyPhase_UserLengths_ValidValues(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseUserLengths
	m.bs.userMinLength = "10"
	m.bs.userMaxLength = "30"
	m.applyPhase()
	require.NotNil(t, m.bs.config.UserSettings)
	assert.True(t, m.bs.config.UserSettings.RandomPassword)
	assert.Equal(t, 10, m.bs.config.UserSettings.MinLength)
	assert.Equal(t, 30, m.bs.config.UserSettings.MaxLength)
}

func TestApplyPhase_UserLengths_DefaultsOnInvalid(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseUserLengths
	m.bs.userMinLength = "abc" // invalid — should fall back to 8
	m.bs.userMaxLength = "0"   // invalid — should fall back to 20
	m.applyPhase()
	require.NotNil(t, m.bs.config.UserSettings)
	assert.Equal(t, 8, m.bs.config.UserSettings.MinLength)
	assert.Equal(t, 20, m.bs.config.UserSettings.MaxLength)
}

func TestApplyPhase_FolderScope_All(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseFolderScope
	m.bs.folderScope = "all"
	m.applyPhase()
	require.NotNil(t, m.bs.config.DashboardSettings)
	assert.True(t, m.bs.config.DashboardSettings.IgnoreFilters)
}

func TestApplyPhase_FolderScope_Allowlist(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseFolderScope
	m.bs.folderScope = "allowlist"
	m.applyPhase()
	// IgnoreFilters should NOT be set for allowlist scope.
	if m.bs.config.DashboardSettings != nil {
		assert.False(t, m.bs.config.DashboardSettings.IgnoreFilters)
	}
}

func TestApplyPhase_FolderMenu_Done_EmptyFolderList(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseFolderMenu
	m.bs.folderAction = "done"
	m.bs.folders = nil
	m.applyPhase()
	// Default fallback: "General"
	assert.Equal(t, []string{"General"}, m.bs.config.MonitoredFolders)
}

func TestApplyPhase_FolderMenu_Done_PreservesExistingFolders(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseFolderMenu
	m.bs.folderAction = "done"
	m.bs.folders = []string{"Prod", "Dev"}
	m.applyPhase()
	assert.Equal(t, []string{"Prod", "Dev"}, m.bs.config.MonitoredFolders)
}

func TestApplyPhase_FolderAdd_PlainName(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseFolderAdd
	m.bs.folderName = "My Dashboards"
	m.applyPhase()
	require.Len(t, m.bs.folders, 1)
	// Plain names get URL-encoded, so the stored value must not have a raw space.
	assert.NotContains(t, m.bs.folders[0], " ")
}

func TestApplyPhase_FolderAdd_RegexName(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseFolderAdd
	m.bs.folderName = "Prod.*"
	m.applyPhase()
	require.Len(t, m.bs.folders, 1)
	// Regex names are stored verbatim (not encoded).
	assert.Equal(t, "Prod.*", m.bs.folders[0])
}

func TestApplyPhase_FolderAdd_Empty_NoOp(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseFolderAdd
	m.bs.folderName = "   " // whitespace only
	m.applyPhase()
	assert.Empty(t, m.bs.folders)
}

func TestApplyPhase_FolderTest_Match(t *testing.T) {
	m := newTUIModel(t)
	m.bs.folders = []string{"General"}
	m.phase = phaseFolderTest
	m.bs.folderTestValue = "General"
	m.applyPhase()
	assert.Contains(t, m.bs.folderTestResult, "MATCH")
}

func TestApplyPhase_FolderTest_NoMatch(t *testing.T) {
	m := newTUIModel(t)
	m.bs.folders = []string{"General"}
	m.phase = phaseFolderTest
	m.bs.folderTestValue = "Production"
	m.applyPhase()
	assert.Contains(t, m.bs.folderTestResult, "NO MATCH")
}

func TestApplyPhase_FolderTest_EmptyValue(t *testing.T) {
	m := newTUIModel(t)
	m.bs.folders = []string{"General"}
	m.phase = phaseFolderTest
	m.bs.folderTestValue = ""
	m.applyPhase()
	assert.Contains(t, m.bs.folderTestResult, "No value entered")
}

func TestApplyPhase_FilterInput_ValidRule(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseFilterInput
	m.bs.filterField = "name"
	m.bs.filterRegex = `DEV-.*`
	m.bs.filterInclusive = false
	m.applyPhase()
	require.Len(t, m.bs.filters, 1)
	assert.Equal(t, "name", m.bs.filters[0].Field)
	assert.Equal(t, `DEV-.*`, m.bs.filters[0].Regex)
	assert.Equal(t, m.bs.filters, m.bs.config.ConnectionSettings.FilterRules)
}

func TestApplyPhase_FilterInput_InvalidRegex_NoAppend(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseFilterInput
	m.bs.filterField = "name"
	m.bs.filterRegex = "[invalid"
	m.applyPhase()
	assert.Empty(t, m.bs.filters)
}

func TestApplyPhase_StorageAssign_Enabled(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseStorageAssign
	m.bs.storageLabel = "my-minio"
	m.bs.storageAssign = true
	m.applyPhase()
	assert.Equal(t, "my-minio", m.bs.config.Storage)
}

func TestApplyPhase_StorageAssign_Disabled(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseStorageAssign
	m.bs.storageLabel = "my-minio"
	m.bs.storageAssign = false
	m.applyPhase()
	assert.Empty(t, m.bs.config.Storage)
}

func TestApplyPhase_PluginConfig_StoresValue(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phasePluginConfig
	m.bs.pluginConfigFields = []string{"key1", "key2"}
	m.bs.pluginConfigIdx = 0
	m.bs.pluginConfigValues = make(map[string]string)
	m.bs.pluginConfigCurrentValue = "value1"
	m.applyPhase()
	assert.Equal(t, "value1", m.bs.pluginConfigValues["key1"])
	assert.Equal(t, 1, m.bs.pluginConfigIdx)
	// pluginResult not yet set — still one field remaining.
	assert.Nil(t, m.bs.pluginResult)
}

func TestApplyPhase_PluginConfig_FinalField_BuildsResult(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phasePluginConfig
	m.bs.pluginConfigFields = []string{"key1"}
	m.bs.pluginConfigIdx = 0
	m.bs.pluginConfigValues = make(map[string]string)
	m.bs.pluginConfigCurrentValue = "secret"
	m.bs.pluginResolvedURL = "https://example.com/plugin.wasm"
	m.applyPhase()
	require.NotNil(t, m.bs.pluginResult)
	assert.Equal(t, "https://example.com/plugin.wasm", m.bs.pluginResult.Url)
	assert.Equal(t, "secret", m.bs.pluginResult.PluginConfig["key1"])
}

func TestApplyPhase_CredRuleCreds_AppendRule(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseCredRuleCreds
	m.bs.credCurrentRules = []config_domain.MatchingRule{
		{Field: "name", Regex: `.*esproxy.*`},
	}
	m.bs.credSecureData = "elastic.yaml"
	m.bs.credUser = "elastic"
	m.bs.credPassword = "pass"
	m.bs.addMoreCredRules = false
	m.applyPhase()
	require.Len(t, m.bs.credRules, 1)
	assert.Equal(t, "elastic.yaml", m.bs.credRules[0].SecureData)
	require.Len(t, m.bs.pendingCreds, 1)
	assert.Equal(t, "elastic", m.bs.pendingCreds[0].user)
	// After apply, a default catch-all rule is always appended.
	rules := m.bs.config.ConnectionSettings.MatchingRules
	last := rules[len(rules)-1]
	assert.Equal(t, "default.yaml", last.SecureData)
}

// ── renderPreview ─────────────────────────────────────────────────────────────

func TestRenderPreview_ContainsContextName(t *testing.T) {
	m := newTUIModel(t)
	preview := m.bs.renderPreview()
	assert.Contains(t, preview, "tui-test")
}

func TestRenderPreview_IsValidYAML(t *testing.T) {
	m := newTUIModel(t)
	preview := m.bs.renderPreview()
	// A valid YAML document will not be empty and should contain a colon.
	assert.NotEmpty(t, preview)
	assert.Contains(t, preview, ":")
}

func TestRenderPreview_WithURL(t *testing.T) {
	m := newTUIModel(t)
	m.bs.config.URL = "http://grafana.example.com"
	preview := m.bs.renderPreview()
	assert.Contains(t, preview, "grafana.example.com")
}

func TestRenderPreview_WithPluginResult(t *testing.T) {
	m := newTUIModel(t)
	m.bs.pluginResult = &config_domain.PluginEntity{
		Url: "https://example.com/cipher.wasm",
	}
	preview := m.bs.renderPreview()
	assert.Contains(t, preview, "cipher.wasm")
	// The plugins section must appear only when pluginResult is set.
	assert.Contains(t, preview, "plugins")
}

func TestRenderPreview_WithoutPluginResult_NoPluginsSection(t *testing.T) {
	m := newTUIModel(t)
	m.bs.pluginResult = nil
	preview := m.bs.renderPreview()
	assert.NotContains(t, preview, "cipher_plugin")
}

func TestRenderPreview_WithFolders(t *testing.T) {
	m := newTUIModel(t)
	m.bs.config.MonitoredFolders = []string{"Production", "Staging"}
	preview := m.bs.renderPreview()
	assert.Contains(t, preview, "Production")
}

func TestRenderPreview_WithOutputPath(t *testing.T) {
	m := newTUIModel(t)
	m.bs.config.OutputPath = "/data/grafana-backups"
	preview := m.bs.renderPreview()
	assert.Contains(t, preview, "grafana-backups")
}

// ── roundtrip: applyPhase then nextPhase ──────────────────────────────────────

// TestRoundtrip_BasicAuthFlow walks through the basic-auth branch of the wizard
// (phaseAuthType → phaseBasicCreds → phaseServerSettings) and verifies that
// both the state mutations and phase transitions are correct end-to-end.
func TestRoundtrip_BasicAuthFlow(t *testing.T) {
	m := newTUIModel(t)

	// Start at phasePluginToggle — skip plugin.
	require.Equal(t, phasePluginToggle, m.phase)
	m.bs.configurePlugin = false
	m.applyPhase()
	m.phase = m.nextPhase()
	assert.Equal(t, phaseAuthType, m.phase)

	// Choose basic auth.
	m.bs.authType = basicAuthForm.String()
	m.applyPhase()
	m.phase = m.nextPhase()
	assert.Equal(t, phaseBasicCreds, m.phase)

	// Fill in credentials.
	m.bs.userName = "alice"
	m.bs.password = "wonderland"
	m.applyPhase()
	m.phase = m.nextPhase()
	assert.Equal(t, phaseServerSettings, m.phase)
	assert.Equal(t, "alice", m.bs.config.UserName)
	assert.Equal(t, "wonderland", m.bs.secure.Password)

	// Fill in server settings.
	m.bs.url = "http://grafana.local"
	m.bs.outputPath = "/tmp/gdg"
	m.applyPhase()
	m.phase = m.nextPhase()
	assert.Equal(t, phaseUserSettings, m.phase)
	assert.Equal(t, "http://grafana.local", m.bs.config.URL)
	assert.Equal(t, "/tmp/gdg", m.bs.config.OutputPath)
}

// TestRoundtrip_StorageSkip confirms that declining storage ends the wizard.
func TestRoundtrip_StorageSkip(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseStorageToggle
	m.bs.configureStorage = false
	m.applyPhase()
	assert.Equal(t, phaseDone, m.nextPhase())
}

// TestRoundtrip_FolderAllowlistAdd adds two folders then finishes.
func TestRoundtrip_FolderAllowlistAdd(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseFolderScope
	m.bs.folderScope = "allowlist"
	m.applyPhase()
	m.phase = m.nextPhase()
	assert.Equal(t, phaseFolderMenu, m.phase)

	// Add "Production".
	m.bs.folderAction = "add"
	m.applyPhase()
	m.phase = m.nextPhase()
	assert.Equal(t, phaseFolderAdd, m.phase)

	m.bs.folderName = "Production"
	m.applyPhase()
	m.phase = m.nextPhase()
	assert.Equal(t, phaseFolderMenu, m.phase)
	require.Len(t, m.bs.folders, 1)

	// Add "Staging".
	m.bs.folderAction = "add"
	m.applyPhase()
	m.phase = m.nextPhase()
	m.bs.folderName = "Staging"
	m.applyPhase()
	m.phase = m.nextPhase()
	require.Len(t, m.bs.folders, 2)

	// Done.
	m.bs.folderAction = "done"
	m.applyPhase()
	m.phase = m.nextPhase()
	assert.Equal(t, phaseConnectionToggle, m.phase)
	assert.Len(t, m.bs.config.MonitoredFolders, 2)
}

// TestRoundtrip_FilterLoop verifies adding two filters in a loop.
func TestRoundtrip_FilterLoop(t *testing.T) {
	m := newTUIModel(t)
	m.phase = phaseFilterToggle
	m.bs.addFilters = true
	m.applyPhase()
	m.phase = m.nextPhase()
	assert.Equal(t, phaseFilterInput, m.phase)

	// First filter.
	m.bs.filterField = "name"
	m.bs.filterRegex = `DEV-.*`
	m.bs.filterInclusive = false
	m.bs.addMoreFilters = true
	m.applyPhase()
	m.phase = m.nextPhase()
	assert.Equal(t, phaseFilterInput, m.phase) // loop back

	// Second filter.
	m.bs.filterField = "type"
	m.bs.filterRegex = "elasticsearch"
	m.bs.filterInclusive = true
	m.bs.addMoreFilters = false
	m.applyPhase()
	m.phase = m.nextPhase()
	assert.Equal(t, phaseDefaultCreds, m.phase)
	require.Len(t, m.bs.filters, 2)

	allFields := strings.Join(func() []string {
		out := make([]string, len(m.bs.filters))
		for i, f := range m.bs.filters {
			out[i] = f.Field
		}
		return out
	}(), ",")
	assert.Contains(t, allFields, "name")
	assert.Contains(t, allFields, "type")
}
