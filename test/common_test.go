package test

import (
	"context"
	"fmt"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service"
	"github.com/google/uuid"
	"github.com/ory/dockertest/v3"
	"gopkg.in/yaml.v3"
	"log"
	"log/slog"
	"net"
	"os"
	"sync"
	"time"

	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

var minioPortResource *dockertest.Resource
var grafanaResource *dockertest.Resource

func setupMinioContainer(pool *dockertest.Pool, wg *sync.WaitGroup) {
	// pulls an image, creates a container based on it and runs it
	defer wg.Done()
	resource, err := pool.Run("bitnami/minio", "latest",
		[]string{"MINIO_ROOT_USER=test", "MINIO_ROOT_PASSWORD=secretsss"})
	if err != nil {
		log.Fatal("Could not start resource", "resource", err)
	}
	minioPortResource = resource

	validatePort(resource, 5*time.Second, []string{"9000"}, "Unable to connect to minio container.  Cannot run test")
	slog.Info("Minio container is up and running")

}

func validatePort(resource *dockertest.Resource, delay time.Duration, ports []string, errorMsg string) {
	time.Sleep(delay)
	for _, port := range ports {
		timeout := time.Second
		actualPort := resource.GetPort(fmt.Sprintf("%s/tcp", port))
		conn, err := net.DialTimeout("tcp", net.JoinHostPort("localhost", actualPort), timeout)
		if err != nil {
			fmt.Println("Connecting error:", err)
			log.Fatalf(errorMsg)
		}

		if conn != nil {
			defer conn.Close()
		}
	}

}

func setupGrafanaContainer(pool *dockertest.Pool, wg *sync.WaitGroup) {
	// pulls an image, creates a container based on it and runs it
	defer wg.Done()
	resource, err := pool.Run("grafana/grafana", "10.0.0-ubuntu",
		[]string{"GF_INSTALL_PLUGINS=grafana-googlesheets-datasource", "GF_AUTH_ANONYMOUS_ENABLED=true"})
	if err != nil {
		log.Fatal("Could not start resource", "err", err)
	}
	grafanaResource = resource

	validatePort(resource, 5*time.Second, []string{"3000"}, "Unable to connect to grafana container.  Cannot run test")

	slog.Info("Grafana container is up and running")
}

func setupDockerTest() *dockertest.Pool {
	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatal("Could not construct pool", "err", err)
	}
	// uses pool to try to connect to Docker
	err = pool.Client.Ping()
	if err != nil {
		log.Fatalf("Could not connect to Docker: %s", err)
	}

	return pool

}

func TestMain(m *testing.M) {
	pool := setupDockerTest()
	var wg *sync.WaitGroup = new(sync.WaitGroup)
	wg.Add(2)
	slog.Info("Starting at", "time", time.Now().String())
	go setupMinioContainer(pool, wg)
	go setupGrafanaContainer(pool, wg)
	wg.Wait()
	slog.Info("Ending at", "end", time.Now().String())

	exitVal := m.Run()

	// You can't defer this because os.Exit doesn't care for defer
	for _, resource := range []*dockertest.Resource{minioPortResource, grafanaResource} {
		if resource == nil {
			slog.Warn("No resource set, skipping cleanup")
			continue
		}
		if err := pool.Purge(resource); err != nil {
			log.Fatalf("Could not purge resource: %s", err)
		} else {
			slog.Info("Resource has been purged")
		}
	}

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

	actualPort := grafanaResource.GetPort(fmt.Sprintf("%s/tcp", "3000"))
	err := os.Setenv("GDG_CONTEXTS__TESTING__URL", fmt.Sprintf("http://localhost:%s", actualPort))

	assert.Nil(t, err)

	config.InitConfig(*cfgName, "'")
	conf := config.Config().ViperConfig()
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

	actualPort := minioPortResource.GetPort(fmt.Sprintf("%s/tcp", "9000"))
	var m = map[string]string{
		service.InitBucket: "true",
		service.CloudType:  params[0],
		service.Prefix:     "dummy",
		service.AccessId:   "test",
		service.SecretKey:  "secretsss",
		service.BucketName: bucketName,
		service.Kind:       "cloud",
		service.Custom:     "true",
		service.Endpoint:   fmt.Sprintf("http://localhost:%s", actualPort),
		service.SSLEnabled: "false",
	}

	cfgObj := config.Config().GetAppConfig()
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
