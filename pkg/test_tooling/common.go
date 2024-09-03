package test_tooling

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service"
	"github.com/esnet/gdg/pkg/test_tooling/common"
	"github.com/esnet/gdg/pkg/test_tooling/containers"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"gopkg.in/yaml.v3"
)

const (
	GrafanaTestVersionEnv = "GRAFANA_TEST_VERSION"

	// #nosec G101
	EnableTokenTestsEnv = "TEST_TOKEN_CONFIG"
)

func InitTest(t *testing.T, cfgName *string, envProp map[string]string) (service.GrafanaService, *viper.Viper, testcontainers.Container, func() error) {
	var (
		suffix string
		err    error
	)

	if len(envProp) == 0 {
		envProp = containers.DefaultGrafanaEnv()
	}
	if _, ok := envProp[containers.EnterpriseLicenceKey]; ok {
		suffix = "-enterprise"
	}
	localGrafanaContainer, cancel := containers.SetupGrafanaContainer(envProp, "", suffix)
	apiClient, v := CreateSimpleClient(t, cfgName, localGrafanaContainer)
	noOp := func() error {
		cancel()
		return nil
	}

	if os.Getenv(EnableTokenTestsEnv) != "1" {
		return apiClient, v, localGrafanaContainer, noOp
	}

	testData, _ := os.ReadFile(v.ConfigFileUsed())
	data := map[string]interface{}{}
	err = yaml.Unmarshal(testData, &data)
	assert.Nil(t, err)

	apiClient.DeleteAllTokens() // Remove any old data
	tokenName, _ := uuid.NewUUID()
	newKey, err := apiClient.CreateAPIKey(tokenName.String(), "admin", 0)
	assert.Nil(t, err)

	wrapper := map[string]*config.GrafanaConfig{}
	_ = wrapper

	level1 := data["contexts"].(map[string]interface{})
	level2 := level1["testing"].(map[string]interface{})
	level2["token"] = newKey.Key
	level2["user_name"] = ""
	level2["password"] = ""

	updatedCfg, err := yaml.Marshal(data)
	assert.Nil(t, err)
	tokenCfg, err := os.CreateTemp("config", "token*.yml")
	assert.Nil(t, err, "Unable to create token configuration file")
	newCfg := tokenCfg.Name()
	err = os.WriteFile(newCfg, updatedCfg, 0o600)
	assert.Nil(t, err)

	cleanUp := func() error {
		cancel()
		return os.Remove(newCfg)
	}

	apiClient, v = CreateSimpleClient(t, &newCfg, localGrafanaContainer)
	return apiClient, v, localGrafanaContainer, cleanUp
}

func CreateSimpleClient(t *testing.T, cfgName *string, container testcontainers.Container) (service.GrafanaService, *viper.Viper) {
	if cfgName == nil {
		cfgName = new(string)
		*cfgName = common.DefaultTestConfig
	}

	actualPort, err := container.Endpoint(context.Background(), "")
	grafanaHost := fmt.Sprintf("http://%s", actualPort)
	err = os.Setenv("GDG_CONTEXTS__TESTING__URL", grafanaHost)
	assert.Nil(t, err)
	dockerContainer, ok := container.(*testcontainers.DockerContainer)
	if ok {
		slog.Info("Grafana Test container running", slog.String("host", grafanaHost+"/login"), slog.String("imageVersion", dockerContainer.Image))
	}

	config.InitGdgConfig(*cfgName, "'")
	conf := config.Config().GetViperConfig(config.ViperGdgConfig)
	assert.NotNil(t, conf)
	// Hack for Local testing
	contextName := conf.GetString("context_name")
	conf.Set(fmt.Sprintf("context.%s.url", contextName), grafanaHost)
	assert.Equal(t, contextName, "testing")
	client := service.NewApiService("dummy")
	path, _ := os.Getwd()
	if strings.Contains(path, "test") {
		err := os.Chdir("..")
		if err != nil {
			slog.Warn("unable to set directory to parent")
		}
	}
	return client, conf
}
