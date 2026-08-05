package config

import (
	"maps"
	"slices"
	"testing"

	assets "github.com/esnet/gdg/config"
	"github.com/esnet/gdg/internal/config/config_domain"
	"github.com/esnet/gdg/pkg/test_tooling/common"
	"github.com/esnet/gdg/pkg/test_tooling/path"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestSecureUnmarshall(t *testing.T) {
	assert := assert.New(t)
	raw, err := assets.GetFile("secure.yml")
	assert.NoError(err)
	assert.NotEmpty(raw)
	cfg := new(config_domain.GDGAppConfiguration)
	err = yaml.Unmarshal([]byte(raw), cfg)
	// plugins
	assert.True(cfg.PluginConfig.Disabled)
	assert.NotNil(cfg.PluginConfig.CipherPlugin)
	assert.Equal(cfg.PluginConfig.CipherPlugin.Url, "https://github.com/esnet/gdg-plugins/raw/refs/tags/0.1.0/plugins/cipher_aes256_gcm.wasm")
	assert.Equal(cfg.PluginConfig.CipherPlugin.FilePath, "")
	assert.Equal(len(cfg.PluginConfig.CipherPlugin.PluginConfig), 1)
	assert.Equal(cfg.PluginConfig.CipherPlugin.PluginConfig["passphrase"], "hello_world")
	assert.NoError(err)
	// Secure
	assert.Equal(len(cfg.SecureConfig), 1)
	keys := slices.Collect(maps.Keys(cfg.SecureConfig))
	const alerting = "alerting"
	assert.True(slices.Contains(keys, alerting))
	assert.True(cfg.SecureConfig[alerting] != nil)
}

// TestDefaultConfig_ReturnsNonEmptyString verifies that the embedded default
// config asset can be loaded and is non-empty.
func TestDefaultConfig_ReturnsNonEmptyString(t *testing.T) {
	cfg := DefaultConfig()
	assert.NotEmpty(t, cfg, "DefaultConfig should return the embedded gdg-example.yml content")
	// The default config is YAML — it should at least contain a "contexts" key.
	assert.Contains(t, cfg, "context", "default config should mention 'context'")
}

// TestLoadDefaultSecureConfig_PopulatesPluginConfig verifies that
// loadDefaultSecureConfig correctly unmarshals the embedded secure.yml into
// the provided GDGAppConfiguration.
func TestLoadDefaultSecureConfig_PopulatesPluginConfig(t *testing.T) {
	cfg := new(config_domain.GDGAppConfiguration)
	err := loadDefaultSecureConfig(cfg)
	require.NoError(t, err)
	// Plugin block should be present and disabled by default.
	assert.True(t, cfg.PluginConfig.Disabled)
	assert.NotNil(t, cfg.PluginConfig.CipherPlugin)
	// Secure config block should contain at least one entry.
	assert.NotEmpty(t, cfg.SecureConfig)
}

// TestInitTemplateConfig_LoadsTemplates verifies that InitTemplateConfig can
// read the example templates config bundled with the test data and return a
// populated TemplatingConfig with at least one dashboard template entry.
func TestInitTemplateConfig_LoadsTemplates(t *testing.T) {
	require.NoError(t, path.FixTestDir("config", "../.."))
	tplCfg := InitTemplateConfig(common.DefaultTemplateConfig)
	require.NotNil(t, tplCfg)
	assert.NotEmpty(t, tplCfg.Entities.Dashboards,
		"template config should contain at least one dashboard template")
	assert.NotEmpty(t, tplCfg.Entities.Dashboards[0].TemplateName,
		"first dashboard template should have a non-empty name")
}

// TestInitTemplateConfig_EmptyOverrideUsesDefault verifies that passing an
// empty string falls back to the default "templates.yml" search, and does not
// panic even when the file is absent (it will fatal — so we only test the
// non-empty path in CI where the file is present).
func TestInitTemplateConfig_ViperConfigSet(t *testing.T) {
	require.NoError(t, path.FixTestDir("config", "../.."))
	tplCfg := InitTemplateConfig(common.DefaultTemplateConfig)
	require.NotNil(t, tplCfg)
	// ViperConfig must be populated by InitTemplateConfig.
	assert.NotNil(t, tplCfg.ViperConfig,
		"InitTemplateConfig should set ViperConfig on the returned struct")
}

func TestConfigSearchPathBuilding(t *testing.T) {
	assert := assert.New(t)
	t.Run("with config file path and yml extension", func(t *testing.T) {
		configDirs, configName, ext := buildConfigSearchPath("/something/config/gdg.yml")
		expectedConfigDirs := append(configSearchPaths, "/something/config")
		assert.Equal(expectedConfigDirs, configDirs)
		assert.Equal("gdg", configName)
		assert.Equal("yml", ext)
	})

	t.Run("with config file path and json extension", func(t *testing.T) {
		configDirs, configName, ext := buildConfigSearchPath("/internal/config/templates.json")
		expectedConfigDirs := append(configSearchPaths, "/internal/config")
		assert.Equal(expectedConfigDirs, configDirs)
		assert.Equal("templates", configName)
		assert.Equal("json", ext)
	})

	t.Run("with config file without directory path", func(t *testing.T) {
		configDirs, configName, ext := buildConfigSearchPath("config.yml")
		assert.Equal(configSearchPaths, configDirs)
		assert.Equal("config", configName)
		assert.Equal("yml", ext)
	})

	t.Run("empty input", func(t *testing.T) {
		configDirs, configName, ext := buildConfigSearchPath("")
		assert.Equal(configSearchPaths, configDirs)
		assert.Equal("", configName)
		assert.Equal("", ext)
	})

	t.Run("with config file path without extension", func(t *testing.T) {
		configDirs, configName, ext := buildConfigSearchPath("/testing/config/gdg")
		expectedConfigDirs := append(configSearchPaths, "/testing/config")
		assert.Equal(expectedConfigDirs, configDirs)
		assert.Equal("gdg", configName)
		assert.Equal("", ext)
	})
}
