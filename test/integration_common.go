package test

import (
	"context"
	"fmt"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func initTest(t *testing.T, cfgName *string) (service.GrafanaService, *viper.Viper) {
	if cfgName == nil {
		cfgName = new(string)
		*cfgName = "testing.yml"
	}

	config.InitConfig(*cfgName, "'")
	conf := config.Config().ViperConfig()
	assert.NotNil(t, conf)
	//Hack for Local testing
	conf.Set("context.testing.url", "http://localhost:3000")
	contextName := conf.GetString("context_name")
	assert.Equal(t, contextName, "testing")
	client := service.NewApiService("dummy")
	path, _ := os.Getwd()
	if strings.Contains(path, "test") {
		err := os.Chdir("..")
		if err != nil {
			log.Warning("unable to set directory to parent")
		}
	}
	return client, conf
}

func SetupCloudFunction(t *testing.T, cfg *viper.Viper, apiClient service.GrafanaService, params []string) (context.Context, service.GrafanaService) {
	_ = os.Setenv(service.InitBucket, "true")
	bucketName := params[1]
	var m = map[string]string{
		service.InitBucket: "true",
		service.CloudType:  params[0],
		service.Prefix:     "dummy",
		service.AccessId:   "test",
		service.SecretKey:  "secretsss",
		service.BucketName: bucketName,
		service.Kind:       "cloud",
		service.Custom:     "true",
		service.Endpoint:   "http://localhost:9000",
		service.SSLEnabled: "false",
	}

	cfgObj := config.Config().GetAppConfig()
	defaultCfg := config.Config().GetDefaultGrafanaConfig()
	defaultCfg.Storage = "test"
	cfgObj.StorageEngine["test"] = m
	apiClient = service.NewApiService("dummy")

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
