package integration_tests

import (
	"github.com/sirupsen/logrus"
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
	context := conf.GetString("context_name")
	assert.Equal(t, context, "testing")
	client := api.NewApiService()
	path, _ := os.Getwd()
	if strings.Contains(path, "integration_tests") {
		err := os.Chdir("..")
		if err != nil {
			logrus.Warning("unable to set directory to parent")
		}
	}
	return client, conf
}
