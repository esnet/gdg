package config_test

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/config/domain"
	"github.com/esnet/gdg/internal/storage"
	resourceTypes "github.com/esnet/gdg/pkg/config/domain"
	"github.com/esnet/gdg/pkg/plugins/secure"
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
	confobj := config.InitGdgConfig(common.DefaultTestConfig)
	conf := confobj.GetViperConfig()
	slog.Info(conf.ConfigFileUsed())

	slog.Info(confobj.ContextName)
	assert.NotNil(t, conf)
	context := conf.GetString("context_name")
	assert.Equal(t, context, "qa")
	grafanaConf := confobj.GetDefaultGrafanaConfig()
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
	confobj := config.InitGdgConfig(common.DefaultTestConfig)
	conf := confobj.GetViperConfig()
	slog.Info(conf.ConfigFileUsed())

	slog.Info(confobj.ContextName)
	assert.NotNil(t, conf)
	context := conf.GetString("context_name")
	assert.Equal(t, context, "qa")
	grafanaConf := confobj.GetDefaultGrafanaConfig()
	assert.NotNil(t, grafanaConf)
	grafanaConf.MonitoredFoldersOverride = []domain.MonitoredOrgFolders{{
		OrganizationName: "Your Org",
		Folders:          []string{"General", "SpecialFolder"},
	}}
	folders := grafanaConf.GetMonitoredFolders(false)
	assert.True(t, slices.Contains(folders, "SpecialFolder"))
	grafanaConf.OrganizationName = "DumbDumb"
	folders = grafanaConf.GetMonitoredFolders(false)
	assert.False(t, slices.Contains(folders, "SpecialFolder"))
	assert.True(t, slices.Contains(folders, "Folder2"))
	grafanaConf.OrganizationName = "Main Org."
	grafanaConf.MonitoredFoldersOverride = nil
	folders = grafanaConf.GetMonitoredFolders(false)
	assert.False(t, slices.Contains(folders, "SpecialFolder"))
	assert.True(t, slices.Contains(folders, "Folder2"))
}

// Ensures that if the config is on a completely different path, the searchPath is updated accordingly
func TestSetupDifferentPath(t *testing.T) {
	cfgFile := DuplicateConfig(t)
	cfg := config.InitGdgConfig(cfgFile)
	conf := cfg.GetViperConfig()
	assert.NotNil(t, conf)
	context := conf.GetString("context_name")
	assert.Equal(t, context, "production")
	grafanaConf := cfg.GetDefaultGrafanaConfig()
	assert.NotNil(t, grafanaConf)
	assert.Equal(t, grafanaConf.OutputPath, "prod")
}

func TestConfigEnv(t *testing.T) {
	os.Setenv("GDG_CONTEXT_NAME", "testing")
	os.Setenv("GDG_CONTEXTS__TESTING__URL", "www.google.com")
	cfg := config.InitGdgConfig(common.DefaultTestConfig)
	conf := cfg.GetViperConfig()
	context := conf.GetString("context_name")
	assert.Equal(t, context, "testing")
	url := conf.GetString("contexts.testing.url")
	assert.Equal(t, url, "www.google.com")
	grafanaConfig := cfg.GetDefaultGrafanaConfig()
	assert.Equal(t, grafanaConfig.URL, url)
	os.Setenv("GDG_CONTEXT_NAME", "production")
	os.Setenv("GDG_CONTEXTS__PRODUCTION__URL", "grafana.com")
	cfg = config.InitGdgConfig(common.DefaultTestConfig)
	conf = cfg.GetViperConfig()
	url = conf.GetString("contexts.production.url")
	assert.Equal(t, url, "grafana.com")
}

func TestConfigSecureCloud(t *testing.T) {
	assert := assert.New(t)
	path.FixTestDir("config", "../..")
	os.Setenv("GDG_CONTEXT_NAME", "testing")
	os.Setenv("GDG_CONTEXTS__TESTING__URL", "www.google.com")
	cfg := config.InitGdgConfig(common.DefaultTestConfig)
	grafanaCfg := cfg.GetDefaultGrafanaConfig()

	t.Run("Base Test", func(t *testing.T) {
		grafanaCfg.Storage = "test"
		m := grafanaCfg.GetCloudAuth()
		assert.Equal(len(m), 2)
		assert.Equal(m[storage.AccessId], "test")
		assert.Equal(m[storage.SecretKey], "secretsss")
	})

	t.Run("No Storage Test", func(t *testing.T) {
		grafanaCfg.Storage = ""
		m := grafanaCfg.GetCloudAuth()
		assert.Equal(len(m), 0)
	})

	t.Run("Override Config Value Test", func(t *testing.T) {
		grafanaCfg = cfg.GetDefaultGrafanaConfig()
		grafanaCfg.Storage = "test"
		storageType, appData := cfg.GetCloudConfiguration(cfg.GetDefaultGrafanaConfig().Storage)
		assert.Equal(storageType, "cloud")
		// Set values in config
		appData[storage.CloudType] = storage.Custom
		appData[storage.AccessId] = "moo"
		appData[storage.SecretKey] = "foo"
		// Ensures they get overridden by the secure values
		storageType, appData = cfg.GetCloudConfiguration(cfg.GetDefaultGrafanaConfig().Storage)
		assert.Equal(appData[storage.AccessId], "test")
		assert.Equal(appData[storage.SecretKey], "secretsss")
	})
}

func TestConfigSecurePath(t *testing.T) {
	os.Setenv("GDG_CONTEXT_NAME", "testing")
	os.Setenv("GDG_CONTEXTS__TESTING__URL", "www.google.com")
	cfg := config.InitGdgConfig(common.DefaultTestConfig)
	grafanaCfg := cfg.GetDefaultGrafanaConfig()
	override := domain.SecureModel{
		Password: "allyourbasesaremine!",
		Token:    "1234",
	}
	assert.NoError(t, grafanaCfg.TestSetSecureAuth(override))
	assert.Equal(t, grafanaCfg.GetPassword(), override.Password)
	assert.Equal(t, grafanaCfg.GetAPIToken(), override.Token)
	// Validate Secure Path behavior
	grafanaCfg.SecureLocationOverride = "/tmp/foobar"
	assert.Equal(t, grafanaCfg.SecureLocation(), grafanaCfg.SecureLocationOverride)
	grafanaCfg.SecureLocationOverride = "../tmp/foobar"
	location := grafanaCfg.SecureLocation()
	assert.True(t, strings.Contains(location, "foobar"))
	assert.True(t, strings.Contains(location, "test"))
}

func validateGrafanaQA(t *testing.T, grafana *domain.GrafanaConfig) {
	assert.Equal(t, "https://staging.grafana.com", grafana.URL)
	assert.Equal(t, "<CHANGEME>", grafana.GetAPIToken())
	assert.Equal(t, "", grafana.UserName)
	assert.Equal(t, "", grafana.GetPassword())
	folders := grafana.GetMonitoredFolders(false)
	assert.True(t, slices.Contains(folders, "Folder1"))
	assert.True(t, slices.Contains(folders, "Folder2"))
	assert.Equal(t, "test/data/org_your-org/connections", grafana.GetPath(resourceTypes.ConnectionResource, grafana.GetOrganizationName()))
	assert.Equal(t, "test/data/org_your-org/dashboards", grafana.GetPath(resourceTypes.DashboardResource, grafana.GetOrganizationName()))
	dsSettings := grafana.ConnectionSettings
	request := models.AddDataSourceCommand{}
	assert.Equal(t, len(grafana.ConnectionSettings.MatchingRules), 3)
	// Last Entry is the default
	secureLoc := grafana.SecureLocation()
	defaultSettings, err := grafana.ConnectionSettings.MatchingRules[2].GetConnectionAuth(secureLoc, secure.NoOpEncoder{})
	assert.Nil(t, err)
	assert.Equal(t, "user", defaultSettings.User())
	assert.Equal(t, "password", defaultSettings.Password())

	request.Name = "Complex Name"
	defaultSettings, _ = dsSettings.GetCredentials(request, secureLoc, secure.NoOpEncoder{})
	assert.Equal(t, "test", defaultSettings.User())
	assert.Equal(t, "secret", defaultSettings.Password())
}
