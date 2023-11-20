package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/esnet/gdg/internal/tools"
	"github.com/thoas/go-funk"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
	"log"
)

func (s *Configuration) GetViperConfig(name string) *viper.Viper {
	if s.viperConfiguration == nil {
		return nil
	}
	return s.viperConfiguration[name]
}

func (s *Configuration) ClearContexts() {
	newContext := make(map[string]*GrafanaConfig)
	newContext["example"] = &GrafanaConfig{
		APIToken: "dummy",
	}
	appCfg := s.GetGDGConfig()
	appCfg.Contexts = newContext
	appCfg.ContextName = "example"
	err := s.SaveToDisk(false)
	if err != nil {
		log.Fatal("Failed to make save changes")
	}

	slog.Info("All contexts were cleared")

}

// GetDefaultGrafanaConfig returns the default aka. selected grafana config
func (s *Configuration) GetDefaultGrafanaConfig() *GrafanaConfig {
	name := s.GetGDGConfig().GetContext()

	val, ok := s.GetGDGConfig().GetContexts()[name]
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
	contexts := s.GetGDGConfig().GetContexts()
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
	s.GetGDGConfig().ContextName = dest
	err = s.SaveToDisk(false)
	if err != nil {
		log.Fatal("Failed to make save changes")
	}
	slog.Info("Copied context to destination, please check your config to confirm", "sourceContext", src, "destinationContext", dest)
}

func (s *Configuration) PrintContext(name string) {
	name = strings.ToLower(name)
	grafana, ok := s.GetGDGConfig().GetContexts()[name]
	if !ok {
		slog.Error("context was not found", "context", name)
		return
	}
	d, err := yaml.Marshal(grafana)
	if err != nil {
		log.Fatal("failed to serialize context", "err", err)
	}
	fmt.Printf("---%s:\n%s\n\n", name, string(d))

}

// DeleteContext remove a given context
func (s *Configuration) DeleteContext(name string) {
	name = strings.ToLower(name) //ensure name is lower case
	contexts := s.GetGDGConfig().GetContexts()
	_, ok := contexts[name]
	if !ok {
		slog.Info("Context not found, cannot delete context", "context", name)
		return
	}
	delete(contexts, name)
	if len(contexts) != 0 {
		for key := range contexts {
			s.GetGDGConfig().ContextName = key
			break
		}
	}

	err := s.SaveToDisk(false)
	if err != nil {
		log.Fatal("Failed to make save changes")
	}
	slog.Info("Deleted context and set new context to", "deletedContext", name, "newActiveContext", s.GetGDGConfig().ContextName)
}

// ChangeContext changes active context
func (s *Configuration) ChangeContext(name string) {
	name = strings.ToLower(name)
	_, ok := s.GetGDGConfig().GetContexts()[name]
	if !ok {
		log.Fatalf("context %s was not found", name)
	}
	s.GetGDGConfig().ContextName = name
	err := s.SaveToDisk(false)
	if err != nil {
		log.Fatal("Failed to make save changes")
	}
	slog.Info("Changed context", "context", name)
}

// SaveToDisk Persists current configuration to disk
func (s *Configuration) SaveToDisk(useViper bool) error {

	if useViper {
		return s.GetViperConfig(ViperGdgConfig).WriteConfig()
	}

	file := s.GetViperConfig(ViperGdgConfig).ConfigFileUsed()
	data, err := yaml.Marshal(s.gdgConfig)
	if err == nil {
		err = os.WriteFile(file, data, 0600)
	}

	return err
}

func (app *GDGAppConfiguration) GetContext() string {
	return strings.ToLower(app.ContextName)
}

// Temporary function
func (app *GDGAppConfiguration) GetContextMap() map[string]interface{} {
	response := make(map[string]interface{})
	data, err := json.Marshal(app.Contexts)
	if err != nil {
		slog.Error("could not serialize contexts")
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
	configSearchPaths = []string{"config", ".", "../config", "../../config", "/etc/gdg"}
)

// GetCloudConfiguration Returns storage type and configuration
func (s *Configuration) GetCloudConfiguration(configName string) (string, map[string]string) {
	appData := s.GetGDGConfig().StorageEngine[configName]
	storageType := "local"
	if len(appData) != 0 {
		storageType = appData["kind"]
	}
	return storageType, appData
}

func (app *GDGAppConfiguration) GetContexts() map[string]*GrafanaConfig {
	return app.Contexts
}

// GetContexts returns map of all contexts
func (s *Configuration) GetContexts() map[string]*GrafanaConfig {
	return s.GetGDGConfig().GetContexts()
}

// IsDebug returns true if debug mode is enabled
func (s *Configuration) IsDebug() bool {
	return s.GetViperConfig(ViperGdgConfig).GetBool("global.debug")
}

// IgnoreSSL returns true if SSL errors should be ignored
func (s *Configuration) IgnoreSSL() bool {
	return s.GetViperConfig(ViperGdgConfig).GetBool("global.ignore_ssl_errors")
}

func Config() *Configuration {
	return configData
}

// GetGDGConfig return instance of gdg app configuration
func (s *Configuration) GetGDGConfig() *GDGAppConfiguration {
	return s.gdgConfig
}

// GetTemplateConfig return instance of gdg app configuration
func (s *Configuration) GetTemplateConfig() *TemplatingConfig {
	return s.templatingConfig
}

func (s *TemplatingConfig) GetTemplate(name string) (*TemplateDashboards, bool) {
	for ndx, t := range s.Entities.Dashboards {
		if t.TemplateName == name {
			return &s.Entities.Dashboards[ndx], true
		}
	}

	return nil, false
}

// setMapValueEnvOverride recursively iterate over the keys and updates the map value accordingly
func setMapValueEnvOverride(keys []string, mapValue map[string]interface{}, value interface{}) {
	if len(keys) > 1 {
		rawInnerObject, ok := mapValue[keys[0]]
		if !ok {
			slog.Warn("No Inner map exists, cannot set Env Override")
			return
		}

		innerMap, ok := rawInnerObject.(map[string]interface{})
		if !ok {
			slog.Warn("cannot traverse full map path.  Unable to set ENV override. Returning ")
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

// buildConfigSearchPath common pattern used when loading configuration for both CLI tools.
func buildConfigSearchPath(configFile string, appName *string) []string {
	var configDirs []string
	if configFile != "" {
		configFileDir := filepath.Dir(configFile)
		if configFileDir != "" {
			configDirs = append([]string{configFileDir}, configSearchPaths...)
		}
		*appName = filepath.Base(configFile)
		*appName = strings.TrimSuffix(*appName, filepath.Ext(*appName))
	} else {
		configDirs = append(configDirs, configSearchPaths...)
	}

	return configDirs
}

func InitTemplateConfig(override string) {
	if configData == nil {
		log.Fatal("GDG configuration was not able to be loaded, cannot continue")
	}
	appName := "templates"
	configDirs := buildConfigSearchPath(override, &appName)
	configData.templatingConfig = new(TemplatingConfig)

	v, err := readViperConfig[TemplatingConfig](appName, configDirs, configData.templatingConfig)
	if err != nil {
		log.Fatal("unable to read templating configuration")
	}
	if configData.viperConfiguration == nil {
		configData.viperConfiguration = make(map[string]*viper.Viper)
	}
	configData.viperConfiguration[ViperTemplateConfig] = v
}

func InitConfig(override, defaultConfig string) {
	configData = &Configuration{}
	appName := "importer"
	configDirs := buildConfigSearchPath(override, &appName)
	var err error
	var v *viper.Viper
	configData.gdgConfig = new(GDGAppConfiguration)

	v, err = readViperConfig[GDGAppConfiguration](appName, configDirs, configData.gdgConfig)
	var configFileNotFoundError viper.ConfigFileNotFoundError
	ok := errors.As(err, &configFileNotFoundError)

	if err != nil && ok {
		slog.Info("No configuration file has been found, creating a default configuration")
		err = os.MkdirAll("config", os.ModePerm)
		if err != nil {
			log.Fatal("unable to create configuration folder: 'config'")
		}
		err = os.WriteFile("config/importer.yml", []byte(defaultConfig), 0600)
		if err != nil {
			log.Panic("Could not persist default config locally")
		}
		appName = "importer"

		v, err = readViperConfig[GDGAppConfiguration](appName, configDirs, configData.gdgConfig)
		if err != nil {
			log.Panic(err)
		}

	} else if err != nil { // config is found but is invalid
		log.Fatal("Invalid configuration detected, please fix your configuration and try again.")
	}
	if configData.viperConfiguration == nil {
		configData.viperConfiguration = make(map[string]*viper.Viper, 0)
	}
	configData.viperConfiguration[ViperGdgConfig] = v

	//unmarshall struct
	contexts := configData.GetViperConfig(ViperGdgConfig).GetStringMap("contexts")
	contexts = applyEnvOverrides(contexts, "contexts", v)

	contextMaps, err := yaml.Marshal(contexts)
	if err != nil {
		log.Fatal("Failed to decode context map, please check your configuration")
	}
	err = yaml.Unmarshal(contextMaps, &configData.gdgConfig.Contexts)
	if err != nil {
		log.Fatal("No valid configuration file has been found")
	}

}

// readViperConfig utilizes the viper library to load the config from the selected paths
func readViperConfig[T any](appName string, configDirs []string, object *T) (*viper.Viper, error) {

	v := viper.New()
	v.SetEnvPrefix("GDG")
	replacer := strings.NewReplacer(".", "__")
	v.SetEnvKeyReplacer(replacer)
	v.SetConfigName(appName)
	for _, dir := range configDirs {
		v.AddConfigPath(dir)
	}

	v.AutomaticEnv()

	err := v.ReadInConfig()
	if err == nil {
		//Marshall the data read into a app struct
		err = v.Unmarshal(object)
	}

	return v, err
}
