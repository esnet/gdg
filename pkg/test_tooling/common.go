package test_tooling

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/config/domain"
	"github.com/esnet/gdg/internal/service"
	"github.com/esnet/gdg/pkg/test_tooling/common"
	"github.com/esnet/gdg/pkg/test_tooling/containers"
	"github.com/esnet/gdg/pkg/test_tooling/path"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
)

const (
	GrafanaTestVersionEnv = "GRAFANA_TEST_VERSION"
	// #nosec G101
	EnableTokenTestsEnv = "TEST_TOKEN_CONFIG"
	FeatureEnabled      = "1"
	FeatureDisabled     = "0"
)

type ConfigProviderFunc func() *config.Configuration

type InitContainerResult struct {
	ApiClient service.GrafanaService
	Container testcontainers.Container
	CleanUp   func() error
	Err       error
}

// NewInitContainerResult creates an InitContainerResult linking a Grafana API client, container and cleanup function.
// It sets Err if the container is not running.
func NewInitContainerResult(client service.GrafanaService, container testcontainers.Container, cleanUp func() error) *InitContainerResult {
	obj := &InitContainerResult{
		ApiClient: client,
		Container: container,
		CleanUp:   cleanUp,
	}
	if !obj.Container.IsRunning() {
		obj.Err = fmt.Errorf("container is not currently running")
	}
	return obj
}

// InitTest starts a Grafana test container, creates a client and optionally configures token auth.
// It returns an InitContainerResult with the client, container, cleanup function and error status.
func InitTest(t *testing.T, cfgProvider config.Provider, envProp map[string]string) *InitContainerResult {
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
	apiClient := CreateSimpleClientWithConfig(t, cfgProvider, localGrafanaContainer)
	cleanUp := func() error {
		cancel()
		return nil
	}

	if os.Getenv(EnableTokenTestsEnv) != FeatureEnabled {
		return NewInitContainerResult(apiClient, localGrafanaContainer, cleanUp)
	}

	// Setup Token Auth
	apiClient.DeleteAllServiceAccounts()
	serviceName, _ := uuid.NewUUID()
	serviceAccnt, err := apiClient.CreateServiceAccount(serviceName.String(), "admin", 0)
	assert.NoError(t, err, "Unable to create test service account")
	if err != nil {
		log.Fatalf("unable to start grafana container for test: %s", t.Name())
	}
	newKey, err := apiClient.CreateServiceAccountToken(serviceAccnt.ID, "admin", 0)
	assert.Nil(t, err)
	if err != nil {
		log.Fatalf("unable to start grafana container for test: %s", t.Name())
	}
	cfg := cfgProvider()
	grafana := cfg.GetDefaultGrafanaConfig()
	grafana.UserName = ""
	secureSettings := domain.SecureModel{
		Password: "",
		Token:    newKey.Key,
	}
	WrapTest(func() {
		assert.NoError(t, cfg.GetDefaultGrafanaConfig().TestSetSecureAuth(secureSettings))
	})

	apiClient = CreateSimpleClientWithConfig(t, cfgProvider, localGrafanaContainer)
	return NewInitContainerResult(apiClient, localGrafanaContainer, cleanUp)
}

// CreateSimpleClientWithConfig creates a GrafanaService for tests using the provided config provider and testcontainers container.
func CreateSimpleClientWithConfig(t *testing.T, cfgProvider config.Provider, container testcontainers.Container) service.GrafanaService {
	cfg := cfgProvider()
	if cfg == nil {
		t.Fatal("No valid configuration returned from config provider")
	}

	actualPort, err := container.Endpoint(context.Background(), "")
	grafanaHost := fmt.Sprintf("http://%s", actualPort)
	cfg.GetDefaultGrafanaConfig().URL = grafanaHost
	dockerContainer, ok := container.(*testcontainers.DockerContainer)
	if ok {
		slog.Info("Grafana Test container running", slog.String("host", grafanaHost+"/login"), slog.String("imageVersion", dockerContainer.Image))
	}

	storageEngine, err := service.ConfigureStorage(cfgProvider)
	assert.NoError(t, err)
	client := service.NewTestApiService(storageEngine, cfgProvider)
	currentPath, _ := os.Getwd()
	if strings.Contains(currentPath, "test") {
		pathErr := os.Chdir("..")
		if pathErr != nil {
			slog.Warn("unable to set directory to parent")
		}
	}
	return client
}

// CreateSimpleClient initializes a test Grafana client and Viper config for unit tests.
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

	config.InitGdgConfig(*cfgName)
	conf := config.Config().GetViperConfig()
	assert.NotNil(t, conf)
	// Hack for Local testing
	contextName := conf.GetString("context_name")
	conf.Set(fmt.Sprintf("context.%s.url", contextName), grafanaHost)
	assert.Equal(t, contextName, "testing")
	storageEngine, err := service.ConfigureStorage(func() *config.Configuration {
		return config.Config()
	})
	assert.NoError(t, err)
	client := service.NewTestApiService(storageEngine, nil)
	currentPath, _ := os.Getwd()
	if strings.Contains(currentPath, "test") {
		err := os.Chdir("..")
		if err != nil {
			slog.Warn("unable to set directory to parent")
		}
	}
	return client, conf
}

// MaintainConfigAuth updates Grafana config auth, preserving existing secure data during init.
func MaintainConfigAuth(configVal string) {
	cfg := config.Config().GetDefaultGrafanaConfig()
	var auth *domain.SecureModel
	WrapTest(func() {
		auth = cfg.TestGetSecureAuth()
	})
	WrapTest(func() {
		config.InitGdgConfig(configVal)
	})
	cfg = config.Config().GetDefaultGrafanaConfig()
	if auth != nil {
		WrapTest(func() {
			if err := cfg.TestSetSecureAuth(*auth); err != nil {
				slog.Warn("unable to set grafana auth")
			}
		})
	}
}

func WrapTest(f func()) {
	os.Setenv(path.TestEnvKey, "1") // #nosec G104
	f()
	os.Unsetenv(path.TestEnvKey)
}
