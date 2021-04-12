package config

import (
	"time"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

// Provider defines a set of read-only methods for accessing the application
// configuration params as defined in one of the config files.
type Provider interface {
	ConfigFileUsed() string
	Get(key string) interface{}
	GetBool(key string) bool
	GetDuration(key string) time.Duration
	GetFloat64(key string) float64
	GetInt(key string) int
	GetInt64(key string) int64
	GetSizeInBytes(key string) uint
	GetString(key string) string
	GetStringMap(key string) map[string]interface{}
	GetStringMapString(key string) map[string]string
	GetStringMapStringSlice(key string) map[string][]string
	GetStringSlice(key string) []string
	GetTime(key string) time.Time
	InConfig(key string) bool
	IsSet(key string) bool
}

type ConfigStruct struct {
	defaultConfig *viper.Viper
	grafanaConfig *GrafanaConfig
}

var configData *ConfigStruct

// Config returns a default config providers
func Config() Provider {
	return configData.defaultConfig
}

func GetGrafanaConfig() *GrafanaConfig {
	return configData.grafanaConfig
}

// LoadConfigProvider returns a configured viper instance
func LoadConfigProvider(appName string) Provider {
	return readViperConfig(appName)
}

func init() {
	configData = &ConfigStruct{}
	configData.defaultConfig = readViperConfig("importer")
	grafana_config := configData.defaultConfig.GetStringMap("grafana")
	grafana_yaml, _ := yaml.Marshal(grafana_config)
	err := yaml.Unmarshal([]byte(grafana_yaml), &configData.grafanaConfig)
	if err != nil {
		panic(err)
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
