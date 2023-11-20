package config_test

import (
	"fmt"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/grafana-swagger-api-golang/goclient/models"
	"golang.org/x/exp/slices"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thoas/go-funk"
)

func DuplicateConfig(t *testing.T) string {
	dir, _ := os.Getwd()
	var err error
	//Fix test path
	if strings.Contains(dir, "config") {
		err = os.Chdir("../..")
		assert.Nil(t, err, "Failed to change to base directory ")
	}

	err = os.Setenv("GDG_CONTEXT_NAME", "production")
	assert.Nil(t, err, "Failed to set override GDG context name via ENV")
	data, err := os.ReadFile("config/testing.yml")
	assert.Nil(t, err, "Failed to read test configuration file")
	destination := os.TempDir()
	cfgFile := fmt.Sprintf("%s/config.yml", destination)
	err = os.WriteFile(cfgFile, data, 0600)
	assert.Nil(t, err, "Failed to save configuration file")

	return cfgFile
}

func TestSetup(t *testing.T) {
	os.Setenv("GDG_CONTEXT_NAME", "qa")
	//clear all ENV values
	for _, key := range os.Environ() {
		if strings.Contains(key, "GDG_") {
			os.Unsetenv(key)
		}
	}

	os.Setenv("GDG_CONTEXT_NAME", "qa")
	config.InitConfig("testing.yml", "")
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
	//clear all ENV values
	for _, key := range os.Environ() {
		if strings.Contains(key, "GDG_") {
			os.Unsetenv(key)
		}
	}

	os.Setenv("GDG_CONTEXT_NAME", "qa")
	config.InitConfig("testing.yml", "")
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
		OrganizationId: 0,
		Folders:        []string{"General", "SpecialFolder"},
	}}
	folders := grafanaConf.GetMonitoredFolders()
	assert.True(t, slices.Contains(folders, "SpecialFolder"))
	grafanaConf.OrganizationId = 2
	folders = grafanaConf.GetMonitoredFolders()
	assert.False(t, slices.Contains(folders, "SpecialFolder"))
	assert.True(t, slices.Contains(folders, "Folder2"))
	grafanaConf.OrganizationId = 0
	grafanaConf.MonitoredFoldersOverride = nil
	folders = grafanaConf.GetMonitoredFolders()
	assert.False(t, slices.Contains(folders, "SpecialFolder"))
	assert.True(t, slices.Contains(folders, "Folder2"))

}

// Ensures that if the config is on a completely different path, the searchPath is updated accordingly
func TestSetupDifferentPath(t *testing.T) {
	cfgFile := DuplicateConfig(t)
	config.InitConfig(cfgFile, "")
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
	config.InitConfig("testing.yml", "")
	conf := config.Config().GetViperConfig(config.ViperGdgConfig)
	context := conf.GetString("context_name")
	assert.Equal(t, context, "testing")
	url := conf.GetString("contexts.testing.url")
	assert.Equal(t, url, "www.google.com")
	grafanaConfig := config.Config().GetDefaultGrafanaConfig()
	assert.Equal(t, grafanaConfig.URL, url)
	os.Setenv("GDG_CONTEXT_NAME", "production")
	os.Setenv("GDG_CONTEXTS__PRODUCTION__URL", "grafana.com")
	config.InitConfig("testing.yml", "")
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
	assert.True(t, funk.Contains(folders, "Folder1"))
	assert.True(t, funk.Contains(folders, "Folder2"))
	assert.Equal(t, "qa/org_1/connections", grafana.GetPath(config.ConnectionResource))
	assert.Equal(t, "qa/org_1/dashboards", grafana.GetPath(config.DashboardResource))
	dsSettings := grafana.DataSourceSettings
	request := models.AddDataSourceCommand{}
	assert.Equal(t, len(grafana.DataSourceSettings.MatchingRules), 3)
	//Last Entry is the default
	defaultSettings := grafana.DataSourceSettings.MatchingRules[2].Auth
	assert.Equal(t, "user", defaultSettings.User)
	assert.Equal(t, "password", defaultSettings.Password)

	request.Name = "Complex Name"
	defaultSettings, _ = dsSettings.GetCredentials(request)
	assert.Equal(t, "test", defaultSettings.User)
	assert.Equal(t, "secret", defaultSettings.Password)
}
