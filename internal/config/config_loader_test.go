package config_test

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/pkg/test_tooling/common"
	"github.com/esnet/gdg/pkg/test_tooling/path"
	"github.com/grafana/grafana-openapi-client-go/models"
	"golang.org/x/exp/slices"

	"github.com/stretchr/testify/assert"
)

func DuplicateConfig(t *testing.T) string {
	assert.NoError(t, path.FixTestDir("config", "../.."))
	err := os.Setenv("GDG_CONTEXT_NAME", "production")
	assert.Nil(t, err, "Failed to set override GDG context name via ENV")
	data, err := os.ReadFile("config/" + common.DefaultTestConfig)
	assert.Nil(t, err, "Failed to read test configuration file")
	destination := os.TempDir()
	cfgFile := fmt.Sprintf("%s/config.yml", destination)
	err = os.WriteFile(cfgFile, data, 0o600)
	assert.Nil(t, err, "Failed to save configuration file")

	return cfgFile
}

func TestSetup(t *testing.T) {
	// clear all ENV values
	for _, key := range os.Environ() {
		if strings.Contains(key, "GDG_") {
			os.Unsetenv(key)
		}
	}
	cwd, _ := os.Getwd()
	if strings.Contains(cwd, "config") {
		os.Chdir("../../")
	}

	os.Setenv("GDG_CONTEXT_NAME", "qa")
	config.InitGdgConfig(common.DefaultTestConfig)
	conf := config.Config().GetViperConfig(config.ViperGdgConfig)
	slog.Info(conf.ConfigFileUsed())

	confobj := config.Config().GetGDGConfig()
	slog.Info(confobj.ContextName)
	assert.NotNil(t, conf)
	context := conf.GetString("context_name")
	assert.Equal(t, context, "qa")
	grafanaConf := config.Config().GetDefaultGrafanaConfig()
	assert.NotNil(t, grafanaConf)
	validateGrafanaQA(t, grafanaConf)
}

func TestWatchedFoldersConfig(t *testing.T) {
	// clear all ENV values
	for _, key := range os.Environ() {
		if strings.Contains(key, "GDG_") {
			os.Unsetenv(key)
		}
	}

	assert.NoError(t, os.Setenv("GDG_CONTEXT_NAME", "qa"))
	config.InitGdgConfig(common.DefaultTestConfig)
	conf := config.Config().GetViperConfig(config.ViperGdgConfig)
	slog.Info(conf.ConfigFileUsed())

	confobj := config.Config().GetGDGConfig()
	slog.Info(confobj.ContextName)
	assert.NotNil(t, conf)
	context := conf.GetString("context_name")
	assert.Equal(t, context, "qa")
	grafanaConf := config.Config().GetDefaultGrafanaConfig()
	assert.NotNil(t, grafanaConf)
	grafanaConf.MonitoredFoldersOverride = []config.MonitoredOrgFolders{{
		OrganizationName: "Your Org",
		Folders:          []string{"General", "SpecialFolder"},
	}}
	folders := grafanaConf.GetMonitoredFolders()
	assert.True(t, slices.Contains(folders, "SpecialFolder"))
	grafanaConf.OrganizationName = "DumbDumb"
	folders = grafanaConf.GetMonitoredFolders()
	assert.False(t, slices.Contains(folders, "SpecialFolder"))
	assert.True(t, slices.Contains(folders, "Folder2"))
	grafanaConf.OrganizationName = "Main Org."
	grafanaConf.MonitoredFoldersOverride = nil
	folders = grafanaConf.GetMonitoredFolders()
	assert.False(t, slices.Contains(folders, "SpecialFolder"))
	assert.True(t, slices.Contains(folders, "Folder2"))
}

// Ensures that if the config is on a completely different path, the searchPath is updated accordingly
func TestSetupDifferentPath(t *testing.T) {
	cfgFile := DuplicateConfig(t)
	config.InitGdgConfig(cfgFile)
	conf := config.Config().GetViperConfig(config.ViperGdgConfig)
	assert.NotNil(t, conf)
	context := conf.GetString("context_name")
	assert.Equal(t, context, "production")
	grafanaConf := config.Config().GetDefaultGrafanaConfig()
	assert.NotNil(t, grafanaConf)
	assert.Equal(t, grafanaConf.OutputPath, "prod")
}

func TestConfigEnv(t *testing.T) {
	os.Setenv("GDG_CONTEXT_NAME", "testing")
	os.Setenv("GDG_CONTEXTS__TESTING__URL", "www.google.com")
	config.InitGdgConfig(common.DefaultTestConfig)
	conf := config.Config().GetViperConfig(config.ViperGdgConfig)
	context := conf.GetString("context_name")
	assert.Equal(t, context, "testing")
	url := conf.GetString("contexts.testing.url")
	assert.Equal(t, url, "www.google.com")
	grafanaConfig := config.Config().GetDefaultGrafanaConfig()
	assert.Equal(t, grafanaConfig.URL, url)
	os.Setenv("GDG_CONTEXT_NAME", "production")
	os.Setenv("GDG_CONTEXTS__PRODUCTION__URL", "grafana.com")
	config.InitGdgConfig(common.DefaultTestConfig)
	conf = config.Config().GetViperConfig(config.ViperGdgConfig)
	url = conf.GetString("contexts.production.url")
	assert.Equal(t, url, "grafana.com")
}

func validateGrafanaQA(t *testing.T, grafana *config.GrafanaConfig) {
	assert.Equal(t, "https://staging.grafana.com", grafana.URL)
	assert.Equal(t, "<CHANGEME>", grafana.APIToken)
	assert.Equal(t, "", grafana.UserName)
	assert.Equal(t, "", grafana.Password)
	folders := grafana.GetMonitoredFolders()
	assert.True(t, slices.Contains(folders, "Folder1"))
	assert.True(t, slices.Contains(folders, "Folder2"))
	assert.Equal(t, "test/data/org_your-org/connections", grafana.GetPath(config.ConnectionResource))
	assert.Equal(t, "test/data/org_your-org/dashboards", grafana.GetPath(config.DashboardResource))
	dsSettings := grafana.ConnectionSettings
	request := models.AddDataSourceCommand{}
	assert.Equal(t, len(grafana.ConnectionSettings.MatchingRules), 3)
	// Last Entry is the default
	secureLoc := grafana.GetPath(config.SecureSecretsResource)
	defaultSettings, err := grafana.ConnectionSettings.MatchingRules[2].GetConnectionAuth(secureLoc)
	assert.Nil(t, err)
	assert.Equal(t, "user", defaultSettings.User())
	assert.Equal(t, "password", defaultSettings.Password())

	request.Name = "Complex Name"
	securePath := grafana.GetPath(config.SecureSecretsResource)
	defaultSettings, _ = dsSettings.GetCredentials(request, securePath)
	assert.Equal(t, "test", defaultSettings.User())
	assert.Equal(t, "secret", defaultSettings.Password())
}
