package integration_tests

import (
	"context"
	"fmt"
	"github.com/graymeta/stow"
	"github.com/graymeta/stow/s3"
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

func SetupCloudFunction(apiClient api.ApiService) context.Context {
	bucketName := "testing"
	m := map[string]interface{}{
		s3.ConfigAccessKeyID: "test",
		s3.ConfigSecretKey:   "secretsss",
		s3.ConfigEndpoint:    "127.0.0.1:9000",
		api.CloudType:        "s3",
		api.BucketName:       bucketName,
		s3.ConfigDisableSSL:  "true",
	}
	ctx := context.Background()
	ctx = context.WithValue(ctx, api.StorageContext, m)
	configMap := stow.ConfigMap{}
	for key, value := range m {
		configMap[key] = fmt.Sprintf("%v", value)
	}
	location, err := stow.Dial(s3.Kind, configMap)
	if err != nil {
		log.Panic("Unable to connect to S3 Minio Storage")
	}
	container, err := location.Container(bucketName)
	if err == nil {
		log.Infof("bucket %s already exists skipping", container.Name())

	} else {
		container, err = location.CreateContainer(bucketName)
		if err != nil {
			log.WithError(err).Fatal("Ignoring failure to bucket creation")
		}
	}

	s := api.NewCloudStorage(ctx)
	dash := apiClient.(*api.DashNGoImpl)
	dash.SetStorage(s)

	return ctx
}
