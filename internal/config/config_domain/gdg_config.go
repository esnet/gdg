package config_domain

import (
	"fmt"
	"log"
	"log/slog"
	"maps"
	"os"
	"slices"
	"strings"

	"github.com/esnet/gdg/internal/adapter/storage"
	"github.com/gosimple/slug"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type GDGAppConfigurationOption func(*GDGAppConfiguration)

// GDGAppConfiguration is the configuration for the application
type GDGAppConfiguration struct {
	ViperConfig   *viper.Viper                 `mapstructure:"-" yaml:"-"`
	ContextName   string                       `mapstructure:"context_name" yaml:"context_name"`
	StorageEngine map[string]map[string]string `mapstructure:"storage_engine" yaml:"storage_engine"`
	Contexts      map[string]*GrafanaConfig    `mapstructure:"contexts" yaml:"contexts"`
	Global        *AppGlobals                  `mapstructure:"global" yaml:"global"`
	SecureConfig  map[string][]string          `mapstructure:"secure_config" yaml:"secure_config"`
	PluginConfig  PluginConfig                 `mapstructure:"plugins" yaml:"plugins"`
}

type PluginConfig struct {
	Disabled     bool          `mapstructure:"disabled" yaml:"disabled"`
	CipherPlugin *PluginEntity `mapstructure:"cipher" yaml:"cipher"`
}

type PluginEntity struct {
	Url          string            `mapstructure:"url" yaml:"url"`
	FilePath     string            `mapstructure:"file_path" yaml:"file_path"`
	PluginConfig map[string]string `mapstructure:"config" yaml:"config"`
	processed    bool
}

// SetPluginConfig sets the plugin configuration to the provided map and marks the entity as unprocessed,
// so that subsequent calls to GetPluginConfig will re-evaluate any environment variable or file references.
func (pe *PluginEntity) SetPluginConfig(config map[string]string) {
	pe.PluginConfig = config
	pe.processed = false
}

// GetPluginConfig returns the plugin configuration map after resolving any dynamic value references.
// Values prefixed with "env:" are resolved from environment variables. Values prefixed with "file:" are
// resolved by reading the referenced file, with environment variable expansion applied to the file path.
// If an environment variable is not set, the original string value is retained. If a file cannot be read,
// the original string value is used and a warning is logged. Results are cached so subsequent calls return
// the previously resolved configuration without reprocessing.
func (pe *PluginEntity) GetPluginConfig() map[string]string {
	if pe.processed {
		return pe.PluginConfig
	}
	m := make(map[string]string)
	for k, v := range pe.PluginConfig {
		if strings.Contains(v, "env:") {
			val := os.Getenv(strings.TrimPrefix(v, "env:"))
			if val != "" {
				m[k] = val
				continue
			}
		} else if after, ok := strings.CutPrefix(v, "file:"); ok {
			loc := after
			expandedFile := os.ExpandEnv(loc)
			raw, err := os.ReadFile(expandedFile) // #nosec G304
			if err == nil {
				m[k] = string(raw)
				continue
			}
			slog.Warn(fmt.Sprintf("unable to read file from variable `%s`, using it value as string", expandedFile))
		}
		m[k] = v
	}
	pe.processed = true
	pe.PluginConfig = m
	return pe.PluginConfig
}

// GetSecureEntities returns the SecureModelConfig, initializing it if nil.
func (app *GDGAppConfiguration) GetSecureEntities() map[string][]string {
	if app.SecureConfig == nil {
		app.SecureConfig = make(map[string][]string)
	}
	return app.SecureConfig
}

// SecureModelConfig defines the field and path of sensitive data tha should be encrypted
type SecureModelConfig struct {
	SecureEntities map[string]SecureEntity `mapstructure:"secure_fields" yaml:"secure_fields"`
}

// SecureFieldNames returns a slice of names for all secure entities.
func (s *SecureModelConfig) SecureFieldNames() []string {
	res := slices.Collect(maps.Keys(s.SecureEntities))
	slices.Sort(res)
	return res
}

type SecureEntity struct {
	Patterns []string `mapstructure:"patterns" yaml:"patterns"`
}

// IgnoreSSL returns true if SSL errors should be ignored
func (app *GDGAppConfiguration) IgnoreSSL() bool {
	return app.GetViperConfig().GetBool("global.ignore_ssl_errors")
}

// IsDebug returns true if debug mode is enabled
func (app *GDGAppConfiguration) IsDebug() bool {
	if val := app.GetViperConfig(); val != nil {
		return val.GetBool("global.debug")
	}
	return false
}

// IsApiDebug returns true if debug mode is enabled for APIs
func (app *GDGAppConfiguration) IsApiDebug() bool {
	if val := app.GetViperConfig(); val != nil {
		return val.GetBool("global.api_debug")
	}
	return false
}

// GetCloudConfiguration Returns storage type and configuration
func (app *GDGAppConfiguration) GetCloudConfiguration(configName string) (string, map[string]string) {
	appData := app.StorageEngine[configName]
	if appData == nil {
		appData = make(map[string]string)
	}

	storageType := "local"
	if len(appData) != 0 {
		storageType = "cloud"
		if appData[storage.CloudType] == storage.Custom {
			grafanaCfg := app.GetDefaultGrafanaConfig()
			m := grafanaCfg.GetCloudAuth()
			// Clear out hard coded values
			appData[storage.SecretKey] = m[storage.SecretKey]
			appData[storage.AccessId] = m[storage.AccessId]
		} else {
			delete(appData, storage.AccessId)
			delete(appData, storage.SecretKey)
		}
	}
	return storageType, appData
}

func (app *GDGAppConfiguration) GetViperConfig() *viper.Viper {
	return app.ViperConfig
}

// PrintContext outputs the YAML representation of the named context and the config file used.
func (app *GDGAppConfiguration) PrintContext(name string) {
	name = strings.ToLower(name)
	grafana, ok := app.GetContexts()[name]
	if !ok {
		slog.Error("context was not found", "context", name)
		return
	}
	d, err := yaml.Marshal(grafana)
	if err != nil {
		log.Fatal("failed to serialize context", "err", err)
	}

	fmt.Printf("config file: %s\n", app.GetViperConfig().ConfigFileUsed())
	fmt.Printf("---context: %s\n%s\n", name, string(d))
}

// PrintContextAll outputs the same YAML as PrintContext plus the plugin configuration
// and any storage engines associated with the named context (or all engines if none
// is assigned).  It is used by "gdg tools contexts show --all".
func (app *GDGAppConfiguration) PrintContextAll(name string) {
	// Always print the base context first.
	app.PrintContext(name)

	// Plugin configuration.
	if !app.PluginConfig.Disabled && app.PluginConfig.CipherPlugin != nil {
		d, err := yaml.Marshal(app.PluginConfig)
		if err == nil {
			fmt.Printf("---plugins:\n%s\n", string(d))
		}
	}

	// Storage engines: show only the one assigned to this context, or all if none assigned.
	if len(app.StorageEngine) > 0 {
		name = strings.ToLower(name)
		ctx := app.GetContexts()[name]
		assigned := ""
		if ctx != nil {
			assigned = ctx.Storage
		}

		engines := app.StorageEngine
		if assigned != "" {
			if entry, ok := engines[assigned]; ok {
				engines = map[string]map[string]string{assigned: entry}
			}
		}

		d, err := yaml.Marshal(map[string]any{"storage_engine": engines})
		if err == nil {
			fmt.Printf("---%s", string(d))
		}
	}
}

// GetDefaultGrafanaConfig returns the default aka. selected grafana config
func (app *GDGAppConfiguration) GetDefaultGrafanaConfig() *GrafanaConfig {
	name := app.GetContext()

	val, ok := app.GetContexts()[name]
	if ok {
		return val
	}
	log.Fatalf("Context: '%s' is not found.  Please check your config", name)
	return nil
}

// UpdateContextNames sets each context's internal name to a slugified version of its key.
func (app *GDGAppConfiguration) UpdateContextNames() {
	for key, cfg := range app.Contexts {
		cfg.contextName = slug.Make(key)
	}
}

// GetContext returns the current context name in lower case for consistent lookup.
func (app *GDGAppConfiguration) GetContext() string {
	return strings.ToLower(app.ContextName)
}

// GetContexts returns the map of context names to their GrafanaConfig.
func (app *GDGAppConfiguration) GetContexts() map[string]*GrafanaConfig {
	return app.Contexts
}

// ChangeContext changes active context
func (app *GDGAppConfiguration) ChangeContext(name string) {
	app.SetContext(name)
	err := app.SaveToDisk(false)
	if err != nil {
		log.Fatal("Failed to make save changes")
	}
	slog.Info("Changed context", "context", name)
}

// SaveToDisk Persists current configuration to disk
func (app *GDGAppConfiguration) SaveToDisk(useViper bool) error {
	if useViper {
		return app.GetViperConfig().WriteConfig()
	}

	file := app.GetViperConfig().ConfigFileUsed()
	data, err := yaml.Marshal(app)
	if err == nil {
		err = os.WriteFile(file, data, 0o600)
	}

	return err
}

// SetContext sets the active context by name after validating its existence.
func (app *GDGAppConfiguration) SetContext(name string) {
	name = strings.ToLower(name)
	_, ok := app.GetContexts()[name]
	if !ok {
		log.Fatalf("context %s was not found", name)
	}

	app.ContextName = name
}

// GetAppGlobals returns the global configuration, initializing it if nil.
func (app *GDGAppConfiguration) GetAppGlobals() *AppGlobals {
	if app.Global == nil {
		app.Global = &AppGlobals{}
	}
	return app.Global
}
