package domain

import (
	"os"
	"strings"
	"testing"

	"github.com/esnet/gdg/pkg/test_tooling/path"
	"github.com/stretchr/testify/assert"
)

func TestPlugCfg(t *testing.T) {
	assert.NoError(t, path.FixTestDir("domain", "../../.."))
	assert := assert.New(t)
	plugCfg := PluginEntity{
		PluginConfig: map[string]string{
			"passphrase":    "env:PLUGIN_URL",
			"missing_env":   "env:someblahenvNotSet",
			"file_path":     "file:test/data/secure/complex.yaml",
			"file_path_env": "file:$cwd/test/data/secure/complex.yaml",
			"file_path_bad": "file:test/data/Dummy",
		},
	}

	os.Setenv("PLUGIN_URL", "foobar")
	dir, err := os.Getwd()
	assert.NoError(err)
	os.Setenv("cwd", dir)

	m := plugCfg.GetPluginConfig()
	assert.Equal(m["passphrase"], "foobar")
	assert.Equal(m["missing_env"], "env:someblahenvNotSet")
	assert.Equal(m["file_path_bad"], "file:test/data/Dummy")
	assert.True(strings.Contains(m["file_path"], "basicAuthPassword"))
	assert.True(strings.Contains(m["file_path_env"], "basicAuthPassword"))
}
