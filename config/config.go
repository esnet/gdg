package config

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

type ConfigStruct struct {
	defaultConfig *viper.Viper
	contextMap    map[string]*GrafanaConfig
}

var configData *ConfigStruct

// Config returns a default config providers
func Config() *viper.Viper {
	return configData.defaultConfig
}

func GetContext() string {
	name := Config().GetString("context_name")
	return name
}

func SetContext(context string) {
	v := LoadConfigProvider("importer")
	v.Set("context_name", context)
	v.WriteConfig()
}

func GetContexts() []string {
	return funk.Keys(configData.contextMap).([]string)
}

func GetGrafanaConfig(name string) *GrafanaConfig {
	val, ok := configData.contextMap[name]
	if ok {
		return val
	} else {
		log.Error("Context is not found.  Please check your config")
		os.Exit(1)
	}

	return nil
}

func GetDefaultGrafanaConfig() *GrafanaConfig {
	return GetGrafanaConfig(GetContext())
}

// LoadConfigProvider returns a configured viper instance
func LoadConfigProvider(appName string) *viper.Viper {
	return readViperConfig(appName)
}

func init() {
	configData = &ConfigStruct{}
	configData.defaultConfig = readViperConfig("importer")
	contexts := configData.defaultConfig.GetStringMap("contexts")
	contextMaps, _ := yaml.Marshal(contexts)
	err := yaml.Unmarshal([]byte(contextMaps), &configData.contextMap)
	if err != nil {
		log.Error("No valid configuration file has been found")
		os.Exit(1)
	}

}

func readViperConfig(appName string) *viper.Viper {
	v := viper.New()
	v.SetEnvPrefix(appName)
	v.SetConfigName(appName)
	v.AddConfigPath(".")
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
