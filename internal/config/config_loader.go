package config

import (
	"log"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"

	assets "github.com/esnet/gdg/config"
	"github.com/esnet/gdg/internal/config/config_domain"
	"github.com/esnet/gdg/internal/logging"
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

// NewConfig initializes the global configuration from a file or defaults.
// It loads gdg.yml (or importer.yml) using Viper, updates context names,
// and stores the configuration in a global variable for later use.
func NewConfig(override string, opts ...config_domain.GDGAppConfigurationOption) *config_domain.GDGAppConfiguration {
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

	gdgConfig := new(config_domain.GDGAppConfiguration)
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

	if gdgConfig.Global != nil {
		if gdgConfig.Global.Debug {
			gdgConfig.Global.Logging.Verbose = true
			slog.Warn("globals.debug is deprecated. Please use globals.logging.verbose instead")
		}

		if gdgConfig.Global.ApiDebug {
			gdgConfig.Global.Logging.HTTPTraffic = true
			slog.Warn("globals.api_debug is deprecated. Please use globals.logging.http_traffic instead")
		}

		gdgConfig.HTTPClient = setupHTTPClient(
			gdgConfig.IsHTTPTrafficLogged(),
			gdgConfig.IsHTTPBodyLogged(),
			gdgConfig.IgnoreSSL(),
		)
	} else {
		gdgConfig.HTTPClient = http.DefaultClient
	}

	for _, opt := range opts {
		opt(gdgConfig)
	}

	return gdgConfig
}

func loadDefaultSecureConfig(gdgConfig *config_domain.GDGAppConfiguration) error {
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

func setupHTTPClient(logHTTPTraffic, includeBody, skipSSL bool) *http.Client {
	if logHTTPTraffic == false {
		return http.DefaultClient
	}

	t := http.DefaultTransport.(*http.Transport).Clone()
	t.TLSClientConfig.InsecureSkipVerify = skipSSL

	lt := logging.NewTransport(t, includeBody)

	return &http.Client{
		Transport: lt,
	}
}

// InitTemplateConfig loads templating configuration from a file or defaults.
// It builds the search path, reads the config with Viper, and returns a
// populated *config_domain.TemplatingConfig instance.```
func InitTemplateConfig(override string) *config_domain.TemplatingConfig {
	var ext, configName string
	var configDirs []string
	if override == "" {
		configDirs, configName, ext = buildConfigSearchPath("templates.yml")
	} else {
		configDirs, configName, ext = buildConfigSearchPath(override)
	}
	templatingConfig := new(config_domain.TemplatingConfig)

	v, err := readViperConfig(configName, configDirs, templatingConfig, ext)
	if err != nil {
		log.Fatalf("unable to read templating configuration, %v", err)
	}
	templatingConfig.ViperConfig = v

	return templatingConfig
}
