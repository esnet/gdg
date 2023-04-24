package config_test

import (
	"fmt"
	"github.com/esnet/gdg/internal/apphelpers"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/grafana-swagger-api-golang/goclient/models"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thoas/go-funk"
)

func TestSetup(t *testing.T) {
	config.InitConfig("testing.yml", "")
	conf := config.Config().ViperConfig()
	assert.NotNil(t, conf)
	context := conf.GetString("context_name")
	assert.Equal(t, context, "qa")
	grafanaConf := apphelpers.GetCtxDefaultGrafanaConfig()
	assert.NotNil(t, grafanaConf)
	validateGrafanaQA(t, grafanaConf)
}

// Ensures that if the config is on a completely different path, the searchPath is updated accordingly
func TestSetupDifferentPath(t *testing.T) {
	dir, _ := os.Getwd()
	//Fix test path
	if strings.Contains(dir, "config") {
		os.Chdir("../..")
	}

	os.Setenv("GDG_CONTEXT_NAME", "production")
	data, err := os.ReadFile("config/testing.yml")
	assert.Nil(t, err, "Failed to read test configuration file")
	destination := os.TempDir()
	cfgFile := fmt.Sprintf("%s/config.yml", destination)
	err = os.WriteFile(cfgFile, data, 0644)
	assert.Nil(t, err, "Failed to save configuration file")
	config.InitConfig(cfgFile, "")
	conf := config.Config().ViperConfig()
	assert.NotNil(t, conf)
	context := conf.GetString("context_name")
	assert.Equal(t, context, "production")
	grafanaConf := apphelpers.GetCtxDefaultGrafanaConfig()
	assert.NotNil(t, grafanaConf)
	assert.Equal(t, grafanaConf.OutputPath, "prod")
}

func TestConfigEnv(t *testing.T) {
	os.Setenv("GDG_CONTEXT_NAME", "testing")
	os.Setenv("GDG_CONTEXTS__TESTING__URL", "www.google.com")
	config.InitConfig("testing.yml", "")
	conf := config.Config().ViperConfig()
	context := conf.GetString("context_name")
	assert.Equal(t, context, "testing")
	url := conf.GetString("contexts.testing.url")
	assert.Equal(t, url, "www.google.com")
	grafanaConfig := apphelpers.GetCtxDefaultGrafanaConfig()
	assert.Equal(t, grafanaConfig.URL, url)
	os.Setenv("GDG_CONTEXT_NAME", "production")
	os.Setenv("GDG_CONTEXTS__PRODUCTION__URL", "grafana.com")
	config.InitConfig("testing.yml", "")
	conf = config.Config().ViperConfig()
	url = conf.GetString("contexts.production.url")
	assert.Equal(t, url, "grafana.com")
}

func validateGrafanaQA(t *testing.T, grafana *config.GrafanaConfig) {
	assert.False(t, grafana.AdminEnabled)
	assert.Equal(t, "https://staging.grafana.com", grafana.URL)
	assert.Equal(t, "<CHANGEME>", grafana.APIToken)
	assert.Equal(t, "", grafana.UserName)
	assert.Equal(t, "", grafana.Password)
	folders := grafana.GetMonitoredFolders()
	assert.True(t, funk.Contains(folders, "Folder1"))
	assert.True(t, funk.Contains(folders, "Folder2"))
	assert.Equal(t, "qa/datasources", grafana.GetDataSourceOutput())
	assert.Equal(t, "qa/dashboards", grafana.GetDashboardOutput())
	dsSettings := grafana.DataSourceSettings
	request := models.AddDataSourceCommand{}
	defaultSettings, _ := dsSettings.GetCredentials(request)
	assert.Equal(t, len(grafana.DataSourceSettings.MatchingRules), 3)
	//Last Entry is the default
	defaultSettings = grafana.DataSourceSettings.MatchingRules[2].Auth
	assert.Equal(t, "user", defaultSettings.User)
	assert.Equal(t, "password", defaultSettings.Password)

	request.Name = "Complex Name"
	defaultSettings, _ = dsSettings.GetCredentials(request)
	assert.Equal(t, "test", defaultSettings.User)
	assert.Equal(t, "secret", defaultSettings.Password)
}
