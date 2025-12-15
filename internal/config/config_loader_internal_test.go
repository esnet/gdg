package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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