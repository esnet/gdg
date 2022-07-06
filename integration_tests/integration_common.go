package integration_tests

import (
	"context"
	"fmt"
	"github.com/graymeta/stow"
	"github.com/graymeta/stow/s3"
	"github.com/graymeta/stow/sftp"
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

func SetupCloudFunction(apiClient api.ApiService, cloudType string) context.Context {
	bucketName := "testing"
	var m map[string]interface{}
	if cloudType == "s3" {
		m = map[string]interface{}{
			s3.ConfigAccessKeyID: "test",
			s3.ConfigSecretKey:   "secretsss",
			s3.ConfigEndpoint:    "127.0.0.1:9000",
			api.CloudType:        "s3",
			api.BucketName:       bucketName,
			s3.ConfigDisableSSL:  "true",
		}
	} else {
		m = map[string]interface{}{
			api.CloudType:       "sftp",
			sftp.ConfigPort:     "2222",
			sftp.ConfigHost:     "127.0.0.1",
			sftp.ConfigUsername: "foo",
			sftp.ConfigPassword: "pass",
			api.BucketName:      bucketName,
		}
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, api.StorageContext, m)
	configMap := stow.ConfigMap{}
	for key, value := range m {
		configMap[key] = fmt.Sprintf("%v", value)
	}

	//Initiate S3
	if cloudType == "s3" {
		location, err := stow.Dial(m[api.CloudType].(string), configMap)
		if err != nil {
			log.Panic("Unable to connect to S3 Minio Storage")
		}
		container, err := location.Container(bucketName)
		if err == nil {
			log.Infof("bucket %s already exists skipping", container.Name())

		} else {
			_, err = location.CreateContainer(bucketName)
			if err != nil {
				log.WithError(err).Fatalf("Ignoring failure to bucket creation %s", bucketName)
			}
		}
	}

	s, err := api.NewCloudStorage(ctx)
	if err != nil {
		log.Fatalf("Could not instantiate cloud storage for type: %s", cloudType)
	}
	dash := apiClient.(*api.DashNGoImpl)
	dash.SetStorage(s)

	return ctx
}
