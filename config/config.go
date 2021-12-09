package config

import (
	"fmt"
	"github.com/thoas/go-funk"
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

//setMapValueEnvOverride recursively iterate over the keys and updates the map value accordingly
func setMapValueEnvOverride(keys []string, mapValue map[string]interface{}, value interface{}) map[string]interface{} {
	if len(keys) > 1  {
		innerMapType := fmt.Sprintf("%T", mapValue[keys[0]])
		if innerMapType  != "map[string]interface {}" {
			log.Warn("cannot traverse full map path.  Unable to set ENV override. Returning ")
			return mapValue
		}
		setMapValueEnvOverride(keys[1:], mapValue[keys[0]].(map[string]interface{}), value)
	} else {
		mapValue[keys[0]] = value
	}

	return mapValue
}

//applyEnvOverrides a bit of a hack to get around a viper limitation.
// GetStringMap does not apply env overrides, so we have to traverse it again
// and reset missing values
func applyEnvOverrides(contexts map[string]interface{}, mapName string, config *viper.Viper) map[string]interface{} {
	keys := config.AllKeys()
	filteredKeys := funk.Filter(keys, func(s string) bool { return strings.Contains(s, "contexts.testing")})
	keys = filteredKeys.([]string)
	for _, key := range keys {
		value := config.Get(key)
		newKey := strings.Replace(key, fmt.Sprintf("%s.", mapName), "", 1)
		contexts = setMapValueEnvOverride(strings.Split(newKey, "."), contexts, value)
	}

	return contexts
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
	contexts = applyEnvOverrides(contexts, "contexts", configData.defaultConfig)

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
	v.SetEnvPrefix("GDG")
	replacer := strings.NewReplacer(".", "__")
	v.SetEnvKeyReplacer(replacer)
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
