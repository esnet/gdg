package config

import (
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

type ConfigStruct struct {
	defaultConfig *viper.Viper
	contextMap    map[string]*GrafanaConfig
}

var configData *ConfigStruct

//ViperConfig returns the loaded configuration via a viper reference
func (s *ConfigStruct) ViperConfig() *viper.Viper {
	return s.defaultConfig
}

//Contexts returns map of all contexts
func (s *ConfigStruct) Contexts() map[string]*GrafanaConfig {
	return s.contextMap
}

//IsDebug returns true if debug mode is enabled
func (s ConfigStruct) IsDebug() bool {
	return s.defaultConfig.GetBool("global.debug")
}

//IgnoreSSL returns true if SSL errors should be ignored
func (s ConfigStruct) IgnoreSSL() bool {
	return s.defaultConfig.GetBool("global.ignore_ssl_errors")
}

func Config() *ConfigStruct {
	return configData
}

func InitConfig(override string) {
	configData = &ConfigStruct{}
	appName := "importer"
	if override != "" {
		appName = filepath.Base(override)
		appName = strings.TrimSuffix(appName, filepath.Ext(appName))
	}

	configData.defaultConfig = readViperConfig(appName)
	contexts := configData.defaultConfig.GetStringMap("contexts")
	contextMaps, err := yaml.Marshal(contexts)
	if err != nil {
		log.Fatal("Failed to decode context map, please check your configuration")
	}
	err = yaml.Unmarshal([]byte(contextMaps), &configData.contextMap)
	if err != nil {
		log.Fatal("No valid configuration file has been found")
	}

}

//readViperConfig utilizes the viper library to load the config from the selected paths
func readViperConfig(appName string) *viper.Viper {
	v := viper.New()
	v.SetEnvPrefix(appName)
	v.SetConfigName(appName)
	v.AddConfigPath(".")
	v.AddConfigPath("../conf")
	v.AddConfigPath("conf")
	v.AutomaticEnv()

	// global defaults
	v.SetDefault("json_logs", false)
	v.SetDefault("loglevel", "debug")

	err := v.ReadInConfig()
	if err != nil {
		panic(err)
	}

	return v
}
