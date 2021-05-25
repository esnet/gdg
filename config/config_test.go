package config_test

import (
	"testing"

	"github.com/netsage-project/grafana-dashboard-manager/config"
	"github.com/stretchr/testify/assert"
	"github.com/thoas/go-funk"
)

func TestSetup(t *testing.T) {
	conf := config.Config()
	assert.NotNil(t, conf)
	context := conf.GetString("context_name")
	assert.Equal(t, context, "qa")
	grafanaConf := config.GetDefaultGrafanaConfig()
	assert.NotNil(t, grafanaConf)
	validateGrafanaQA(t, grafanaConf)
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
	assert.Equal(t, "qa/datasources", grafana.OutputDataSource)
	assert.Equal(t, "qa/dashboards", grafana.OutputDashboard)
	dsSettings := grafana.DataSourceSettings
	defaultSettings := dsSettings["default"]
	assert.Equal(t, "user", defaultSettings.User)
	assert.Equal(t, "password", defaultSettings.Password)
	defaultSettings = dsSettings["complex name"]
	assert.Equal(t, "test", defaultSettings.User)
	assert.Equal(t, "secret", defaultSettings.Password)
}
