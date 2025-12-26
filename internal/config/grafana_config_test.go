package config

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/esnet/gdg/internal/config/domain"
	"github.com/esnet/gdg/pkg/test_tooling/common"
	"github.com/esnet/gdg/pkg/test_tooling/path"
	"github.com/stretchr/testify/assert"
)

func TestGrafanaConfig(t *testing.T) {
	config := domain.NewGrafanaConfig("testing")
	config.URL = "  http://localhost  "

	expected := "http://localhost/"

	if expected != config.GetURL() {
		t.Errorf("expected %s, got %s", expected, config.GetURL())
	}
}

func TestPrintConfig(t *testing.T) {
	assert := assert.New(t)
	assert.NoError(path.FixTestDir("config", "../.."))
	confobj := InitGdgConfig(common.DefaultTestConfig)
	backupStd := os.Stdout
	backupErr := os.Stderr
	r, w, e := os.Pipe()
	if e != nil {
		panic(e)
	}
	defer func() {
		os.Stdout = backupStd
		os.Stderr = backupErr
	}()
	os.Stdout = w
	os.Stderr = w
	confobj.PrintContext("testing")
	assert.NoError(w.Close())

	out, _ := io.ReadAll(r)
	output := string(out)
	assert.True(strings.Contains(output, "testing.yml"))
	assert.True(strings.Contains(output, "dashboard_settings:"))
	assert.True(strings.Contains(output, "output_path: test/data"))
}
