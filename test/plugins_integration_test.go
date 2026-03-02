package test

import (
	"fmt"
	"log/slog"
	"os"
	"testing"

	"github.com/esnet/gdg/internal/adapter/grafana/api"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/config/config_domain"
	"github.com/esnet/gdg/pkg/test_tooling"
	"github.com/esnet/gdg/pkg/test_tooling/common"
	"github.com/esnet/gdg/pkg/test_tooling/path"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/matryer/is"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

const (
	cipherKey = "dumbdumb"
	plugFile  = "testing_plugins.yml"
)

type plugTestType string

const (
	filePlugType  plugTestType = "file"
	envPlugType   plugTestType = "env"
	basicPlugType plugTestType = "basic"
)

func TestConnectionsPluginCfg(t *testing.T) {
	// Plugin is only configured for basic auth, skipping token based patterns
	assert.NoError(t, os.Unsetenv(common.ContextNameEnv))
	test_tooling.SkipTokenBasedTests(t)
	is := is.New(t)
	err := path.FixTestDir("test", "..")
	is.NoErr(err)

	testCases := []struct {
		name     string
		testType plugTestType
		enabled  bool
	}{
		{
			name:     "plaintextAuth Plugin test",
			testType: "basic",
			enabled:  true,
		},
		{
			name:     "EnvAuth Plugin test",
			testType: "env",
			enabled:  true,
		},
		{
			name:     "FileAuth Plugin test",
			testType: "file",
			enabled:  true,
		},
	}
	for _, tc := range testCases {
		if !tc.enabled {
			t.Log("Skipping", tc.name)
			continue
		}

		cfg := config.NewConfig(plugFile)
		patchConfig(t, cfg, tc.testType)
		r := test_tooling.InitTest(t, cfg, nil)
		is.True(r != nil)
		defer func() {
			err := r.CleanUp()
			if err != nil {
				slog.Warn("Unable to clean up after test", "test", t.Name())
			}
		}()

		apiClient := r.ApiClient
		filtersEntity := api.NewConnectionFilter("")
		slog.Info("Exporting all connections")
		apiClient.UploadConnections(filtersEntity)
		slog.Info("Listing all connections")
		dataSources := apiClient.ListConnections(filtersEntity)
		assert.Equal(t, len(dataSources), 4)
		dsItem := lo.FirstOrEmpty(lo.Filter(dataSources, func(item models.DataSourceListItemDTO, index int) bool {
			return item.Name == "netsage"
		}))
		assert.NotNil(t, dsItem)
		validateConnection(t, dsItem)
		// Import Dashboards
		slog.Info("Importing connections")
		list := apiClient.DownloadConnections(filtersEntity)
		assert.Equal(t, len(list), len(dataSources))
		slog.Info("Deleting connections")
		deleteList := apiClient.DeleteAllConnections(filtersEntity)
		assert.Equal(t, len(deleteList), len(dataSources))
		slog.Info("List connections again")
		dataSources = apiClient.ListConnections(filtersEntity)
		assert.Equal(t, len(dataSources), 0)
	}
}

func patchConfig(t *testing.T, cfg *config_domain.GDGAppConfiguration, testType plugTestType) {
	cfg.PluginConfig.Disabled = false
	cfg.PluginConfig.CipherPlugin.GetPluginConfig()
	switch testType {
	case envPlugType:
		os.Setenv(test_tooling.TestPlugSecretEnv, cipherKey)
		cfg.PluginConfig.CipherPlugin.SetPluginConfig(map[string]string{
			"passphrase": fmt.Sprintf("env:%s", test_tooling.TestPlugSecretEnv),
		})
	case filePlugType:
		tmpFile, err := os.CreateTemp("", "cipher-secret-*.cipher")
		if err != nil {
			t.Fatal(err)
		}
		err = os.WriteFile(tmpFile.Name(), []byte(cipherKey), 0o644)
		if err != nil {
			t.Fatal(err)
		}
		cfg.PluginConfig.CipherPlugin.SetPluginConfig(map[string]string{
			"passphrase": fmt.Sprintf("file:%s", tmpFile.Name()),
		})
	case basicPlugType:
		cfg.PluginConfig.CipherPlugin.SetPluginConfig(map[string]string{
			"passphrase": cipherKey,
		})
	default:
		t.Fatalf("Unsupported test type: %s", testType)
	}
}
