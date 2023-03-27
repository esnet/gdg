package integration_tests

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
	"testing"

	"github.com/esnet/gdg/api"
	"github.com/esnet/gdg/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func initTest(t *testing.T) (api.ApiService, *viper.Viper) {
	config.InitConfig("testing.yml", "'")
	conf := config.Config().ViperConfig()
	assert.NotNil(t, conf)
	conf.Set("context_name", "testing")
	//Hack for Local testing
	conf.Set("context.testing.url", "http://localhost:3000")
	contextName := conf.GetString("context_name")
	assert.Equal(t, contextName, "testing")
	client := api.NewApiService("dummy")
	path, _ := os.Getwd()
	if strings.Contains(path, "integration_tests") {
		err := os.Chdir("..")
		if err != nil {
			log.Warning("unable to set directory to parent")
		}
	}
	return client, conf
}

func SetupCloudFunction(apiClient api.ApiService, params []string) context.Context {
	bucketName := params[1]
	var m = map[string]interface{}{
		api.CloudType:  params[0],
		api.Prefix:     "dummy",
		api.BucketName: bucketName,
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, api.StorageContext, m)
	configMap := map[string]string{}
	for key, value := range m {
		configMap[key] = fmt.Sprintf("%v", value)
	}

	s, err := api.NewCloudStorage(ctx)
	if err != nil {
		log.Fatalf("Could not instantiate cloud storage for type: %s", params[0])
	}
	dash := apiClient.(*api.DashNGoImpl)
	dash.SetStorage(s)

	return ctx
}
