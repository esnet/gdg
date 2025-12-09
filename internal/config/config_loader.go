package config

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	assets "github.com/esnet/gdg/config"
	"github.com/esnet/gdg/internal/config/domain"
	"github.com/esnet/gdg/internal/tools"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

func (s *Configuration) GetViperConfig() *viper.Viper {
	return s.gdgViperConfig
}

func (s *Configuration) DefaultConfig() string {
	cfg, err := assets.GetFile("importer-example.yml")
	if err != nil {
		slog.Warn("unable to find load default configuration", "err", err)
	}
	return cfg
}

func (s *Configuration) ClearContexts() {
	newContext := make(map[string]*domain.GrafanaConfig)
	newContext["example"] = &domain.GrafanaConfig{}
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
func (s *Configuration) GetDefaultGrafanaConfig() *domain.GrafanaConfig {
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
	// Validate context
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
	name = strings.ToLower(name) // ensure name is lower case
	contexts := s.GetGDGConfig().GetContexts()
	_, ok := contexts[name]
	if !ok {
		log.Fatalf("Context not found, cannot delete context: %s", name)
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

func (s *Configuration) SetContext(name string) {
	name = strings.ToLower(name)
	_, ok := s.GetGDGConfig().GetContexts()[name]
	if !ok {
		log.Fatalf("context %s was not found", name)
	}

	s.GetGDGConfig().ContextName = name
}

// ChangeContext changes active context
func (s *Configuration) ChangeContext(name string) {
	s.SetContext(name)
	err := s.SaveToDisk(false)
	if err != nil {
		log.Fatal("Failed to make save changes")
	}
	slog.Info("Changed context", "context", name)
}

// SaveToDisk Persists current configuration to disk
func (s *Configuration) SaveToDisk(useViper bool) error {
	if useViper {
		return s.GetViperConfig().WriteConfig()
	}

	file := s.GetViperConfig().ConfigFileUsed()
	data, err := yaml.Marshal(s.gdgConfig)
	if err == nil {
		err = os.WriteFile(file, data, 0o600)
	}

	return err
}

var (
	configData        *Configuration
	configSearchPaths = []string{"config", ".", "/etc/gdg"}
)

// GetCloudConfiguration Returns storage type and configuration
func (s *Configuration) GetCloudConfiguration(configName string) (string, map[string]string) {
	appData := s.GetGDGConfig().StorageEngine[configName]
	storageType := "local"
	if len(appData) != 0 {
		storageType = "cloud"
	}
	return storageType, appData
}

// GetContexts returns map of all contexts
func (s *Configuration) GetContexts() map[string]*domain.GrafanaConfig {
	return s.GetGDGConfig().GetContexts()
}

// IsDebug returns true if debug mode is enabled
func (s *Configuration) IsDebug() bool {
	if val := s.GetViperConfig(); val != nil {
		return val.GetBool("global.debug")
	}
	return false
}

// IsApiDebug returns true if debug mode is enabled for APIs
func (s *Configuration) IsApiDebug() bool {
	if val := s.GetViperConfig(); val != nil {
		return val.GetBool("global.api_debug")
	}
	return false
}

// IgnoreSSL returns true if SSL errors should be ignored
func (s *Configuration) IgnoreSSL() bool {
	return s.GetViperConfig().GetBool("global.ignore_ssl_errors")
}

func Config() *Configuration {
	return configData
}

// GetGDGConfig return instance of gdg app configuration
func (s *Configuration) GetGDGConfig() *domain.GDGAppConfiguration {
	return s.gdgConfig
}

// GetTemplateConfig return instance of gdg app configuration
func (s *Configuration) GetTemplateConfig() *domain.TemplatingConfig {
	return s.templatingConfig
}

// buildConfigSearchPath common pattern used when loading configuration for both CLI tools.
func buildConfigSearchPath(configFilePath string) (configDirs []string, configName, ext string) {
	if configFilePath != "" {
		ext = filepath.Ext(configFilePath)
		configName = strings.TrimSuffix(filepath.Base(configFilePath), ext)

		configFilePathDir := filepath.Dir(configFilePath)
		if configFilePathDir != "" {
			configDirs = append([]string{configFilePathDir}, configSearchPaths...)
		}

		if ext == "" {
			ext = "yaml"
		} else {
			ext = ext[1:] // strip leading dot
		}
	} else {
		configDirs = append(configDirs, configSearchPaths...)
	}

	return configDirs, configName, ext
}

func InitGdgConfig(override string) {
	if override == "" && configData != nil {
		return
	}
	configData = &Configuration{}
	var configDirs []string
	var ext, configName string
	if override == "" {
		configDirs, configName, ext = buildConfigSearchPath("config/importer.yml")
	} else {
		configDirs, configName, ext = buildConfigSearchPath(override)
	}
	var err error
	var v *viper.Viper
	configData.gdgConfig = new(domain.GDGAppConfiguration)

	v, err = readViperConfig(configName, configDirs, configData.gdgConfig, ext)
	if err != nil {
		log.Fatal("No configuration file has been found or config is invalid. " + 
			"Expected a file named 'importer.yml' in one of the following folders: ['.', 'config', '/etc/gdg']. " +
			"Try using `gdg default-config > config/importer.yml` go use the default example")
	}
	configData.gdgConfig.UpdateContextNames()
	configData.gdgViperConfig = v
}

// readViperConfig utilizes the viper library to load the config from the selected paths
func readViperConfig[T any](configName string, configDirs []string, object *T, ext string) (*viper.Viper, error) {
	v := viper.New()
	v.SetEnvPrefix("GDG")
	replacer := strings.NewReplacer(".", "__")
	v.SetEnvKeyReplacer(replacer)
	v.SetConfigName(configName)
	if ext == "" {
		v.SetConfigType("yaml") // REQUIRED if the config file does not have the extension in the name
	} else {
		v.SetConfigType(ext)
	}
	for _, dir := range configDirs {
		v.AddConfigPath(dir)
	}

	v.AutomaticEnv()

	err := v.ReadInConfig()
	if err == nil {
		// Marshall the data read into app struct
		err = v.Unmarshal(object)
	}

	return v, err
}

func InitTemplateConfig(override string) {
	if configData == nil {
		log.Fatal("GDG configuration was not able to be loaded, cannot continue")
	}
	var ext, configName string
	var configDirs []string
	if override == "" {
		configDirs, configName, ext = buildConfigSearchPath("config/templates.yml")
	} else {
		configDirs, configName, ext = buildConfigSearchPath(override)
	}
	configData.templatingConfig = new(domain.TemplatingConfig)

	_, err := readViperConfig(configName, configDirs, configData.templatingConfig, ext)
	if err != nil {
		log.Fatal("unable to read templating configuration")
	}
}
