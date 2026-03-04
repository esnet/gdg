package config

import (
	"strings"
	"testing"

	"github.com/esnet/gdg/pkg/test_tooling/path"
	"github.com/matryer/is"
)

func TestAssets(t *testing.T) {
	path.FixTestDir("config", "..")
	is := is.New(t)
	file, err := GetFile("gdg-example.yml")
	is.NoErr(err)
	is.True(strings.Contains(file, "storage_engine"))
	// failing test
	file, err = GetFile("dummy")
	is.True(err != nil)
	is.Equal(err.Error(), "open dummy: file does not exist")
	is.Equal(file, "")
}
