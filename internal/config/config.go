package config

import (
	"encoding/json"
	"fmt"
	"github.com/esnet/gdg/internal/tools"
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
	AppConfig     *AppConfig
}

type AppGlobals struct {
	Debug           bool `mapstructure:"debug" yaml:"debug"`
	IgnoreSSLErrors bool `mapstructure:"ignore_ssl_errors" yaml:"ignore_ssl_errors"`
}

type AppConfig struct {
	ContextName   string                       `mapstructure:"context_name" yaml:"context_name"`
	StorageEngine map[string]map[string]string `mapstructure:"storage_engine" yaml:"storage_engine"`
	Contexts      map[string]*GrafanaConfig    `mapstructure:"contexts" yaml:"contexts"`
	Global        *AppGlobals                  `mapstructure:"global" yaml:"global"`
}

func (s *Configuration) ClearContexts() {
	newContext := make(map[string]*GrafanaConfig)
	newContext["example"] = &GrafanaConfig{
		APIToken: "dummy",
	}
	appCfg := s.GetAppConfig()
	appCfg.Contexts = newContext
	appCfg.ContextName = "example"
	err := s.SaveToDisk(false)
	if err != nil {
		log.Fatal("Failed to make save changes")
	}

	log.Info("All contexts were cleared")

}

// GetDefaultGrafanaConfig returns the default aka. selected grafana config
func (s *Configuration) GetDefaultGrafanaConfig() *GrafanaConfig {
	name := s.GetAppConfig().GetContext()

	val, ok := s.GetAppConfig().GetContexts()[name]
	if ok {
		return val
	} else {
		log.Fatalf("Context: '%s' is not found.  Please check your config", name)
	}
	return nil

}

// CopyContext Makes a copy of the specified context and write to disk
func (s *Configuration) CopyContext(src, dest string) {
	//Validate context
	contexts := s.GetAppConfig().GetContexts()
	if len(contexts) == 0 {
		log.Fatal("Cannot set context.  No valid configuration found in importer.yml")
	}
	cfg, ok := contexts[src]
	if !ok {
		log.Fatalf("Cannot find context to: '%s'.  No valid configuration found in importer.yml", src)
	}
	newCopy, err := tools.DeepCopy(*cfg)
	if err != nil {
		log.Fatal("unable to make a copy of contexts")

	}
	contexts[dest] = newCopy
	s.GetAppConfig().ContextName = dest
	err = s.SaveToDisk(false)
	if err != nil {
		log.Fatal("Failed to make save changes")
	}
	log.Infof("Copied %s context to %s please check your config to confirm", src, dest)
}

func (s *Configuration) PrintContext(name string) {
	name = strings.ToLower(name)
	grafana, ok := s.GetAppConfig().GetContexts()[name]
	if !ok {
		log.Errorf("context %s was not found", name)
		return
	}
	d, err := yaml.Marshal(grafana)
	if err != nil {
		log.WithError(err).Fatal("failed to serialize context")
	}
	fmt.Printf("---%s:\n%s\n\n", name, string(d))

}

// DeleteContext remove a given context
func (s *Configuration) DeleteContext(name string) {
	name = strings.ToLower(name) //ensure name is lower case
	contexts := s.GetAppConfig().GetContexts()
	_, ok := contexts[name]
	if !ok {
		log.Infof("Context not found, cannot delete context named '%s'", name)
		return
	}
	delete(contexts, name)
	if len(contexts) != 0 {
		for key, _ := range contexts {
			s.GetAppConfig().ContextName = key
			break
		}
	}

	err := s.SaveToDisk(false)
	if err != nil {
		log.Fatal("Failed to make save changes")
	}
	log.Infof("Delete %s context and set new context to %s", name, s.GetAppConfig().ContextName)
}

// ChangeContext
func (s *Configuration) ChangeContext(name string) {
	name = strings.ToLower(name)
	_, ok := s.GetAppConfig().GetContexts()[name]
	if !ok {
		log.Fatalf("context %s was not found", name)
	}
	s.GetAppConfig().ContextName = name
	err := s.SaveToDisk(false)
	if err != nil {
		log.Fatal("Failed to make save changes")
	}
	log.Infof("Change context to: '%s'", name)
}

// SaveToDisk Persists current configuration to disk
func (s *Configuration) SaveToDisk(useViper bool) error {
	if useViper {
		return s.ViperConfig().WriteConfig()
	}

	file := s.ViperConfig().ConfigFileUsed()
	data, err := yaml.Marshal(s.AppConfig)
	if err == nil {
		err = os.WriteFile(file, data, 0600)
	}

	return err
}

func (app *AppConfig) GetContext() string {
	return strings.ToLower(app.ContextName)
}

// Temporary function
func (app *AppConfig) GetContextMap() map[string]interface{} {
	response := make(map[string]interface{})
	data, err := json.Marshal(app.Contexts)
	if err != nil {
		log.Errorf("could not serialize contexts")
		return response
	}
	err = json.Unmarshal(data, &response)
	if err != nil {
		return make(map[string]interface{})
	}

	return response

}

var (
	configData        *Configuration
	configSearchPaths = []string{".", "../../config", "../config", "conf", "config", "/etc/gdg"}
)

// GetCloudConfiguration Returns storage type and configuration
func (s *Configuration) GetCloudConfiguration(configName string) (string, map[string]string) {
	appData := s.AppConfig.StorageEngine[configName]
	storageType := "local"
	if len(appData) != 0 {
		storageType = appData["kind"]
	}
	return storageType, appData
}

// ViperConfig returns the loaded configuration via a viper reference
func (s *Configuration) ViperConfig() *viper.Viper {
	return s.defaultConfig
}

func (app *AppConfig) GetContexts() map[string]*GrafanaConfig {
	return app.Contexts
}

// GetContexts returns map of all contexts
func (s *Configuration) GetContexts() map[string]*GrafanaConfig {
	return s.GetAppConfig().GetContexts()
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

func (s *Configuration) GetAppConfig() *AppConfig {
	return s.AppConfig
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

	configData.defaultConfig, configData.AppConfig, err = readViperConfig(appName, configDirs)
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

		configData.defaultConfig, configData.AppConfig, err = readViperConfig(appName, configDirs)
		if err != nil {
			log.Panic(err)
		}

	}

	//unmarshall struct

	contexts := configData.defaultConfig.GetStringMap("contexts")
	contexts = applyEnvOverrides(contexts, "contexts", configData.defaultConfig)

	contextMaps, err := yaml.Marshal(contexts)
	if err != nil {
		log.Fatal("Failed to decode context map, please check your configuration")
	}
	err = yaml.Unmarshal(contextMaps, &configData.AppConfig.Contexts)
	if err != nil {
		log.Fatal("No valid configuration file has been found")
	}

}

// readViperConfig utilizes the viper library to load the config from the selected paths
func readViperConfig(appName string, configDirs []string) (*viper.Viper, *AppConfig, error) {
	app := &AppConfig{}
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
	//v.SetDefault("globals.json_logs", false)
	//v.SetDefault("loglevel", "debug")

	err := v.ReadInConfig()
	if err == nil {
		//Marshall the data read into a app struct
		err = v.Unmarshal(app)
	}

	return v, app, err
}
