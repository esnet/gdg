package test

import (
	"context"
	"fmt"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service"
	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"golang.org/x/exp/maps"
	"gopkg.in/yaml.v3"
	"log"
	"log/slog"
	"os"
	"slices"
	"sync"
	"time"

	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

var minioContainer testcontainers.Container
var grafnaContainer testcontainers.Container

type Containers struct {
	Cancel    context.CancelFunc
	Container testcontainers.Container
}

func setupMinioContainer(wg *sync.WaitGroup, channels chan Containers) {
	// pulls an image, creates a container based on it and runs it
	defer wg.Done()

	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "bitnami/minio:latest",
		ExposedPorts: []string{"9000/tcp", "9001/tcp"},
		Env:          map[string]string{"MINIO_ROOT_USER": "test", "MINIO_ROOT_PASSWORD": "secretsss"},
		WaitingFor:   wait.ForListeningPort("9000/tcp"),
	}
	minioC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		panic(err)
	}

	slog.Info("Minio container is up and running")
	cancel := func() {
		if err := minioC.Terminate(ctx); err != nil {
			panic(err)
		} else {
			slog.Info("Minio container has been terminated")
		}
	}
	result := Containers{
		Cancel:    cancel,
		Container: minioC,
	}
	channels <- result

}

func setupGrafanaContainer(wg *sync.WaitGroup, channels chan Containers) {
	// pulls an image, creates a container based on it and runs it
	defer wg.Done()
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "grafana/grafana:10.0.0-ubuntu",
		ExposedPorts: []string{"3000/tcp"},
		Env: map[string]string{
			"GF_INSTALL_PLUGINS":        "grafana-googlesheets-datasource",
			"GF_AUTH_ANONYMOUS_ENABLED": "true",
		},
		WaitingFor: wait.ForListeningPort("3000/tcp"),
	}
	grafanaC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		panic(err)
	}
	cancel := func() {
		if err := grafanaC.Terminate(ctx); err != nil {
			panic(err)
		} else {
			slog.Info("Grafana Container has been terminated")
		}
	}
	result := Containers{
		Cancel:    cancel,
		Container: grafanaC,
	}
	channels <- result

}

func TestMain(m *testing.M) {
	channels := make(chan Containers, 2)
	var wg = new(sync.WaitGroup)
	wg.Add(2)
	slog.Info("Starting at", "time", time.Now().String())
	go setupMinioContainer(wg, channels)
	go setupGrafanaContainer(wg, channels)
	wg.Wait()
	close(channels)
	slog.Info("Ending at", "end", time.Now().String())

	for entry := range channels {
		defer entry.Cancel()
		str, err := entry.Container.Ports(context.Background())
		if err != nil {
			slog.Error("unable to obtain bound ports for container")
			continue
		}
		keys := maps.Keys(str)
		if slices.Contains(keys, "9000/tcp") {
			minioContainer = entry.Container
		}
		if slices.Contains(keys, "3000/tcp") {
			grafnaContainer = entry.Container

		}

	}
	exitVal := m.Run()

	os.Exit(exitVal)
}

func initTest(t *testing.T, cfgName *string) (service.GrafanaService, *viper.Viper, func() error) {
	apiClient, v := createSimpleClient(t, cfgName)
	noOp := func() error { return nil }

	if os.Getenv("TEST_TOKEN_CONFIG") != "1" {
		return apiClient, v, noOp
	}

	testData, _ := os.ReadFile(v.ConfigFileUsed())
	data := map[string]interface{}{}
	err := yaml.Unmarshal(testData, &data)
	assert.Nil(t, err)

	apiClient.DeleteAllTokens() //Remove any old data
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
	err = os.WriteFile(newCfg, updatedCfg, 0644)
	assert.Nil(t, err)

	cleanUp := func() error {
		return os.Remove(newCfg)
	}

	apiClient, v = createSimpleClient(t, &newCfg)
	return apiClient, v, cleanUp

}

func createSimpleClient(t *testing.T, cfgName *string) (service.GrafanaService, *viper.Viper) {
	if cfgName == nil {
		cfgName = new(string)
		*cfgName = "testing.yml"
	}

	actualPort, err := grafnaContainer.Endpoint(context.Background(), "")
	err = os.Setenv("GDG_CONTEXTS__TESTING__URL", fmt.Sprintf("http://%s", actualPort))
	assert.Nil(t, err)

	config.InitConfig(*cfgName, "'")
	conf := config.Config().GetViperConfig(config.ViperGdgConfig)
	assert.NotNil(t, conf)
	//Hack for Local testing
	contextName := conf.GetString("context_name")
	conf.Set(fmt.Sprintf("context.%s.url", contextName), fmt.Sprintf("http://localhost:%s", actualPort))
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

func SetupCloudFunction(params []string) (context.Context, service.GrafanaService) {
	_ = os.Setenv(service.InitBucket, "true")
	bucketName := params[1]

	actualPort, err := minioContainer.Endpoint(context.Background(), "")
	var m = map[string]string{
		service.InitBucket: "true",
		service.CloudType:  params[0],
		service.Prefix:     "dummy",
		service.AccessId:   "test",
		service.SecretKey:  "secretsss",
		service.BucketName: bucketName,
		service.Kind:       "cloud",
		service.Custom:     "true",
		service.Endpoint:   fmt.Sprintf("http://%s", actualPort),
		service.SSLEnabled: "false",
	}

	cfgObj := config.Config().GetGDGConfig()
	defaultCfg := config.Config().GetDefaultGrafanaConfig()
	defaultCfg.Storage = "test"
	cfgObj.StorageEngine["test"] = m
	apiClient := service.NewApiService("dummy")

	ctx := context.Background()
	ctx = context.WithValue(ctx, service.StorageContext, m)
	configMap := map[string]string{}
	for key, value := range m {
		configMap[key] = fmt.Sprintf("%v", value)
	}

	s, err := service.NewCloudStorage(ctx)
	if err != nil {
		log.Fatalf("Could not instantiate cloud storage for type: %s", params[0])
	}
	dash := apiClient.(*service.DashNGoImpl)
	dash.SetStorage(s)

	return ctx, apiClient
}
