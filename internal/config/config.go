package config

import (
	"fmt"
	"github.com/thoas/go-funk"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type Configuration struct {
	defaultConfig *viper.Viper
	contextMap    map[string]*GrafanaConfig
}

var (
	configData        *Configuration
	configSearchPaths = []string{".", "../../config", "../config", "conf", "config", "/etc/gdg"}
)

// ViperConfig returns the loaded configuration via a viper reference
func (s *Configuration) ViperConfig() *viper.Viper {
	return s.defaultConfig
}

// Contexts returns map of all contexts
func (s *Configuration) Contexts() map[string]*GrafanaConfig {
	return s.contextMap
}

// IsDebug returns true if debug mode is enabled
func (s *Configuration) IsDebug() bool {
	return s.defaultConfig.GetBool("global.debug")
}

// IgnoreSSL returns true if SSL errors should be ignored
func (s *Configuration) IgnoreSSL() bool {
	return s.defaultConfig.GetBool("global.ignore_ssl_errors")
}

func Config() *Configuration {
	return configData
}

// setMapValueEnvOverride recursively iterate over the keys and updates the map value accordingly
func setMapValueEnvOverride(keys []string, mapValue map[string]interface{}, value interface{}) {
	if len(keys) > 1 {
		rawInnerObject, ok := mapValue[keys[0]]
		if !ok {
			log.Warn("No Inner map exists, cannot set Env Override")
			return
		}

		innerMap, ok := rawInnerObject.(map[string]interface{})
		if !ok {
			log.Warn("cannot traverse full map path.  Unable to set ENV override. Returning ")
			return
		}
		setMapValueEnvOverride(keys[1:], innerMap, value)
	} else {
		mapValue[keys[0]] = value
	}

}

// applyEnvOverrides a bit of a hack to get around a viper limitation.
// GetStringMap does not apply env overrides, so we have to traverse it again
// and reset missing values
func applyEnvOverrides(contexts map[string]interface{}, mapName string, config *viper.Viper) map[string]interface{} {
	keys := config.AllKeys()
	matchKey := fmt.Sprintf("contexts.%s", config.GetString("context_name"))
	filteredKeys := funk.Filter(keys, func(s string) bool { return strings.Contains(s, matchKey) })
	keys = filteredKeys.([]string)
	for _, key := range keys {
		value := config.Get(key)
		newKey := strings.Replace(key, fmt.Sprintf("%s.", mapName), "", 1)
		setMapValueEnvOverride(strings.Split(newKey, "."), contexts, value)
	}

	return contexts
}

func InitConfig(override, defaultConfig string) {
	configData = &Configuration{}
	appName := "importer"
	var configDirs []string
	if override != "" {
		overrideDir := filepath.Dir(override)
		if overrideDir != "" {
			configDirs = append([]string{overrideDir}, configSearchPaths...)
		}
		appName = filepath.Base(override)
		appName = strings.TrimSuffix(appName, filepath.Ext(appName))
	} else {
		configDirs = append(configDirs, configSearchPaths...)
	}
	var err error

	configData.defaultConfig, err = readViperConfig(appName, configDirs)
	if err != nil {
		err = os.MkdirAll("config", os.ModePerm)
		if err != nil {
			log.Fatal("unable to create configuration folder: 'config'")
		}
		err = os.WriteFile("config/importer.yml", []byte(defaultConfig), 0600)
		if err != nil {
			log.Panic("Could not persist default config locally")
		}
		appName = "importer"

		configData.defaultConfig, err = readViperConfig(appName, configDirs)
		if err != nil {
			log.Panic(err)
		}

	}
	contexts := configData.defaultConfig.GetStringMap("contexts")
	contexts = applyEnvOverrides(contexts, "contexts", configData.defaultConfig)

	contextMaps, err := yaml.Marshal(contexts)
	if err != nil {
		log.Fatal("Failed to decode context map, please check your configuration")
	}
	err = yaml.Unmarshal(contextMaps, &configData.contextMap)
	if err != nil {
		log.Fatal("No valid configuration file has been found")
	}

}

// readViperConfig utilizes the viper library to load the config from the selected paths
func readViperConfig(appName string, configDirs []string) (*viper.Viper, error) {
	v := viper.New()
	v.SetEnvPrefix("GDG")
	replacer := strings.NewReplacer(".", "__")
	v.SetEnvKeyReplacer(replacer)
	v.SetConfigName(appName)
	for _, dir := range configDirs {
		v.AddConfigPath(dir)
	}

	v.AutomaticEnv()

	// global defaults
	v.SetDefault("json_logs", false)
	v.SetDefault("loglevel", "debug")

	err := v.ReadInConfig()

	return v, err
}
