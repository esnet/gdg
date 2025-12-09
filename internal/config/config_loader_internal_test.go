package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigSearchPath(t *testing.T) {
	assert := assert.New(t)
	t.Run("with config file path", func(t *testing.T) {
		configDirs, configName, ext := buildConfigSearchPath("/something/config/importer.yml")
		exptectedConfigDirs := append([]string{"/something/config"}, configSearchPaths...)
		assert.Equal(exptectedConfigDirs, configDirs)
		assert.Equal("importer", configName)
		assert.Equal("yml", ext)
	})
	
	t.Run("empty input", func(t *testing.T) {
		configDirs, configName, ext := buildConfigSearchPath("")
		assert.Equal(configSearchPaths, configDirs)
		assert.Equal("", configName)
		assert.Equal("", ext)
	})
}