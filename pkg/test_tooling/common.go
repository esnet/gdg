package test_tooling

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/esnet/gdg/internal/adapter/grafana/api"
	"github.com/esnet/gdg/internal/adapter/grafana/extended"
	"github.com/esnet/gdg/internal/adapter/grafana/resources"
	"github.com/esnet/gdg/internal/adapter/plugins/secure/cipher"
	"github.com/esnet/gdg/internal/adapter/plugins/secure/noop"
	"github.com/esnet/gdg/internal/adapter/storage"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/config/config_domain"
	"github.com/esnet/gdg/internal/ports/outbound"
	"github.com/esnet/gdg/pkg/test_tooling/common"
	"github.com/esnet/gdg/pkg/test_tooling/containers"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
)

const (
	GrafanaTestVersionEnv = "GRAFANA_TEST_VERSION"
	// #nosec G101
	EnableTokenTestsEnv = "TEST_TOKEN_CONFIG"
	TestPlugSecretEnv   = "TEST_PLUG_SECRET"
	FeatureEnabled      = "1"
	FeatureDisabled     = "0"
)

type InitContainerResult struct {
	ApiClient outbound.GrafanaService
	Container testcontainers.Container
	CleanUp   func() error
	Err       error
}

func SkipTokenBasedTests(t *testing.T) {
	if IsTokenBasedTest() {
		t.Skip("skipping token based tests")
	}
}

func SkipEnterpriseTests(t *testing.T) {
	if os.Getenv(containers.DisableEnterpriseTest) == "true" {
		t.Skip("Enterprise tests disabled by environment variable")
	}
}

// IsTokenBasedTest returns true if the TEST_TOKEN_CONFIG environment variable is set to "1", indicating that
// token-based authentication tests should be run.
func IsTokenBasedTest() bool {
	return os.Getenv(EnableTokenTestsEnv) == "1"
}

func WithSecureAuth(auth config_domain.SecureModel) config_domain.GDGAppConfigurationOption {
	return func(s *config_domain.GDGAppConfiguration) {
		someCfg := s.GetDefaultGrafanaConfig()
		someCfg.Apply(config_domain.WithSecureAuth(auth))
	}
}

// NewInitContainerResult creates an InitContainerResult linking a Grafana API client, container and cleanup function.
// It sets Err if the container is not running.
func NewInitContainerResult(client outbound.GrafanaService, container testcontainers.Container, cleanUp func() error) *InitContainerResult {
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
func InitTest(t *testing.T, cfg *config_domain.GDGAppConfiguration, envProp map[string]string) *InitContainerResult {
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
	apiClient := CreateSimpleClientWithConfig(t, cfg, localGrafanaContainer)
	cleanUp := func() error {
		cancel()
		return nil
	}

	if !IsTokenBasedTest() {
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
	grafana := cfg.GetDefaultGrafanaConfig()
	grafana.UserName = ""
	secureSettings := config_domain.SecureModel{
		Password: "",
		Token:    newKey.Key,
	}

	grafana.Apply(config_domain.WithSecureAuth(secureSettings))

	apiClient = CreateSimpleClientWithConfig(t, cfg, localGrafanaContainer)
	return NewInitContainerResult(apiClient, localGrafanaContainer, cleanUp)
}

// CreateSimpleClientWithConfig creates a GrafanaService for tests using the provided config provider and testcontainers container.
func CreateSimpleClientWithConfig(t *testing.T, cfg *config_domain.GDGAppConfiguration, container testcontainers.Container) outbound.GrafanaService {
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

	storageType, appData := cfg.GetCloudConfiguration(cfg.GetDefaultGrafanaConfig().Storage)
	var encoder outbound.CipherEncoder
	if !cfg.PluginConfig.Disabled && cfg.PluginConfig.CipherPlugin != nil {
		encoder, err = cipher.NewPluginCipherEncoder(cfg.PluginConfig.CipherPlugin, cfg.SecureConfig)
		assert.NoError(t, err, "failed to load cipher plugin")
	} else {
		encoder = noop.NoOpEncoder{}
	}

	storageEngine, err := storage.NewStorageFromConfig(storageType, appData, encoder)
	assert.NoError(t, err)
	client := api.NewDashNGo(cfg, encoder, storageEngine, extended.NewExtendedApi(cfg), resources.NewHelpers())
	client.Login()
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
func CreateSimpleClient(t *testing.T, cfg *config_domain.GDGAppConfiguration, cfgName *string, container testcontainers.Container) (outbound.GrafanaService, *viper.Viper) {
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

	config.NewConfig(*cfgName)
	conf := cfg.GetViperConfig()
	assert.NotNil(t, conf)
	// Hack for Local testing
	contextName := conf.GetString("context_name")
	conf.Set(fmt.Sprintf("context.%s.url", contextName), grafanaHost)
	assert.Equal(t, contextName, "testing")
	//If needed
	/*
		cfg := rootSvc.LoadConfig(configPath, contextOverride)
		var encoder contract.CipherEncoder
		if !cfg.PluginConfig.Disabled && cfg.PluginConfig.CipherPlugin != nil {
			encoder = secure.NewPluginCipherEncoder(cfg.PluginConfig.CipherPlugin, cfg.SecureConfig)
		} else {
			encoder = secure.NoOpEncoder{}
		}


	*/
	storageType, appData := cfg.GetCloudConfiguration(cfg.GetDefaultGrafanaConfig().Storage)
	storageEngine, err := storage.NewStorageFromConfig(storageType, appData, noop.NoOpEncoder{})
	assert.NoError(t, err)
	client := api.NewDashNGo(cfg, noop.NoOpEncoder{}, storageEngine, extended.NewExtendedApi(cfg), resources.NewHelpers())
	client.Login()
	currentPath, _ := os.Getwd()
	if strings.Contains(currentPath, "test") {
		err := os.Chdir("..")
		if err != nil {
			slog.Warn("unable to set directory to parent")
		}
	}
	return client, conf
}
