package config

import (
	"log"
	"log/slog"
	"path/filepath"
	"strings"

	assets "github.com/esnet/gdg/config"
	"github.com/esnet/gdg/internal/config/domain"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

const (
	defaultConfigName = "gdg-example.yml"
)

var configSearchPaths = []string{"config", ".", "/etc/gdg"}

func DefaultConfig() string {
	cfg, err := assets.GetFile(defaultConfigName)
	if err != nil {
		slog.Warn("unable to find load default configuration", "err", err)
	}
	return cfg
}

// buildConfigSearchPath common pattern used when loading configuration for both CLI tools.
func buildConfigSearchPath(configFilePath string) (configDirs []string, configName, ext string) {
	configDirs = configSearchPaths

	if configFilePath != "" {
		ext = filepath.Ext(configFilePath)
		configName = strings.TrimSuffix(filepath.Base(configFilePath), ext)

		configFilePathDir := filepath.Dir(configFilePath)
		if configFilePathDir != "." {
			configDirs = append(configDirs, configFilePathDir)
		}

		if len(ext) > 0 {
			ext = ext[1:] // strip leading dot
		}
	}

	return configDirs, configName, ext
}

// InitGdgConfig initializes the global configuration from a file or defaults.
// It loads gdg.yml (or importer.yml) using Viper, updates context names,
// and stores the configuration in a global variable for later use.
func InitGdgConfig(override string) *domain.GDGAppConfiguration {
	var (
		configDirs      []string
		ext, configName string
		overrides       []string
		defaultConfig   bool
		err             error
		v               *viper.Viper
	)

	if override != "" {
		overrides = append(overrides, override)
	} else {
		defaultConfig = true
		// Try gdg.yml and then fallback on importer.yml
		overrides = append(overrides, []string{"config/gdg.yml", "config/importer.yml"}...)
	}

	gdgConfig := new(domain.GDGAppConfiguration)
	parseErr := loadDefaultSecureConfig(gdgConfig)
	if parseErr != nil {
		slog.Warn("unable to find default secure.yml", "err", parseErr)
	}

	for _, configOverride := range overrides {
		configDirs, configName, ext = buildConfigSearchPath(configOverride)
		v, err = readViperConfig(configName, configDirs, gdgConfig, ext)
		if err == nil {
			if defaultConfig && strings.Contains("importer", configName) {
				slog.Warn("importer.yml as default config is deprecated. Please use gdg.yml moving forward.")
			}
			break
		}
	}
	if err != nil {
		log.Fatal("No configuration file has been found or config is invalid. " +
			"Expected a file named 'gdg.yml' in one of the following folders: ['.', 'config', '/etc/gdg']. " +
			"Try using `gdg default-config > config/gdg.yml` go use the default example")
	}
	gdgConfig.UpdateContextNames()
	gdgConfig.ViperConfig = v
	return gdgConfig
}

func loadDefaultSecureConfig(gdgConfig *domain.GDGAppConfiguration) error {
	// PreLoad Secure Defaults
	secureFile, err := assets.GetFile("secure.yml")
	if err != nil {
		return err
	}
	raw := []byte(secureFile)
	err = yaml.Unmarshal(raw, gdgConfig)
	if err != nil {
		return err
	}
	return nil
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

// InitTemplateConfig loads templating configuration from a file or defaults.
// It builds the search path, reads the config with Viper, and returns a
// populated *domain.TemplatingConfig instance.```
func InitTemplateConfig(override string) *domain.TemplatingConfig {
	var ext, configName string
	var configDirs []string
	if override == "" {
		configDirs, configName, ext = buildConfigSearchPath("templates.yml")
	} else {
		configDirs, configName, ext = buildConfigSearchPath(override)
	}
	templatingConfig := new(domain.TemplatingConfig)

	v, err := readViperConfig(configName, configDirs, templatingConfig, ext)
	if err != nil {
		log.Fatalf("unable to read templating configuration, %v", err)
	}
	templatingConfig.ViperConfig = v

	return templatingConfig
}
