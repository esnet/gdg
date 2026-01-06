package config

import (
	"maps"
	"slices"
	"testing"

	assets "github.com/esnet/gdg/config"
	"github.com/esnet/gdg/internal/config/domain"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestSecureUnmarshall(t *testing.T) {
	assert := assert.New(t)
	raw, err := assets.GetFile("secure.yml")
	assert.NoError(err)
	assert.NotEmpty(raw)
	cfg := new(domain.GDGAppConfiguration)
	err = yaml.Unmarshal([]byte(raw), cfg)
	// plugins
	assert.True(cfg.PluginConfig.Disabled)
	assert.NotNil(cfg.PluginConfig.CipherPlugin)
	assert.Equal(cfg.PluginConfig.CipherPlugin.Url, "https://raw.githubusercontent.com/esnet/gdg-plugins/refs/heads/main/plugins/cipher_aes256_gcm.wasm")
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
