package config

import (
	"fmt"
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

//ViperConfig returns the the loaded configuration via a viper reference
func (s *ConfigStruct) ViperConfig() *viper.Viper {
	return s.defaultConfig
}

//IsDebug returns true if debug mode is enabled
func (s ConfigStruct) IsDebug() bool {
	return s.defaultConfig.GetBool("global.debug")
}

//IgnoreSSL returns true if SSL errors should be ignored
func (s ConfigStruct) IgnoreSSL() bool {
	return s.defaultConfig.GetBool("global.ignore_ssl_errors")
}

// func Config
func Config() *ConfigStruct {
	return configData
}

//GetContext returns the name of the selected context
func GetContext() string {
	name := Config().ViperConfig().GetString("context_name")
	return name
}

//SetContext will try to find the specified context, if it exists in the file, will re-write the importer.yml
//with the selected context
func SetContext(context string) {
	v := LoadConfigProvider("importer")
	m := v.GetStringMap(fmt.Sprintf("contexts.%s", context))
	if len(m) == 0 {
		log.Fatal("Cannot set context.  No valid configuration found in importer.yml")
	}
	v.Set("context_name", context)
	v.WriteConfig()
}

//GetContexts returns all available contexts
func GetContexts() []string {
	return funk.Keys(configData.contextMap).([]string)
}

//GetGrafanaConfig returns the selected context or terminates app if not found
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

//GetDefaultGrafanaConfig returns the default aka. selected grafana config
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
