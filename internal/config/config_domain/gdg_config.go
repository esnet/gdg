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
	// CipherPlugin holds the entire "plugins.cipher" block: the cipher
	// plugin's own enabled/disabled switch (CipherConfig.Disabled) sits
	// alongside its PluginEntity configuration (url/file_path/config),
	// mirroring how Lookup.Disabled sits alongside the lookup providers
	// rather than being a top-level, plugins-wide kill switch.
	CipherPlugin *CipherConfig `mapstructure:"cipher" yaml:"cipher"`

	// Lookup holds the entire "plugins.lookup" block: a subsystem-wide
	// enabled/disabled switch alongside every named lookup provider's
	// configuration.
	Lookup LookupConfig `mapstructure:"lookup" yaml:"lookup"`
}

// CipherConfig is the "plugins.cipher" block. Disabled is a named field
// scoped to the cipher plugin alone (there is no longer a top-level,
// plugins-wide "plugins.disabled" switch); PluginEntity is squashed/inlined
// so the YAML shape stays flat:
//
//	cipher:
//	  disabled: false
//	  url: https://example.com/cipher_aes256_gcm.wasm
//	  config:
//	    passphrase: hello_world
type CipherConfig struct {
	// Disabled turns off the cipher-plugin code path. When true, or when
	// CipherPlugin is nil, files are read/written in plaintext via
	// NoOpEncoder instead of the configured cipher plugin.
	Disabled bool `mapstructure:"disabled" yaml:"disabled"`

	// PluginEntity is squashed (mapstructure) / inlined (yaml) so its
	// fields (url, file_path, config) sit flat alongside Disabled under
	// "plugins.cipher", rather than nesting under another key.
	PluginEntity `mapstructure:",squash" yaml:",inline"`
}

// LookupConfig is the "plugins.lookup" block:
//
//	lookup:
//	  disabled: false
//	  gsm:
//	    url: https://example.com/lookup_gsm.wasm
//	    config:
//	      credentials: env:GOOGLE_APPLICATION_CREDENTIALS
//
// Disabled is a named field so a single flag can turn off every provider
// at once; every other key under "lookup" (e.g. "gsm", "vault") is a named
// provider and falls through into Plugins, keyed by the provider name
// referenced in a "lookup:<name>:<key>" value. Each entry is resolved the
// same way a cipher plugin is.
type LookupConfig struct {
	// Disabled turns off the entire lookup-plugin code path — every named
	// provider in Plugins at once — with a single flag, so a user who
	// doesn't want the feature at all isn't required to disable each
	// configured provider individually.
	Disabled bool `mapstructure:"disabled" yaml:"disabled"`

	// Plugins holds every other key under "lookup" as a named provider's
	// PluginEntity config. The "mapstructure:,remain" / "yaml:,inline"
	// tags make Disabled and Plugins coexist at the same YAML level: named
	// fields (Disabled) are matched first, and anything left over falls
	// into this map instead of requiring its own nesting level.
	Plugins map[string]*PluginEntity `mapstructure:",remain" yaml:",inline"`
}

// LookupEnabled reports whether the lookup plugin subsystem should be
// active. Both the global plugin kill switch (Disabled) and the
// lookup-specific one (Lookup.Disabled) must allow it; callers building a
// LookupResolver should check this once, up front, rather than requiring
// each named provider to carry its own enabled/disabled flag.
func (pc *PluginConfig) LookupEnabled() bool {
	return !pc.Lookup.Disabled
}

// CipherEnabled reports whether a cipher plugin is configured and not
// disabled. It mirrors LookupEnabled's role for the lookup subsystem: a nil
// CipherPlugin or CipherPlugin.Disabled == true both mean "use NoOpEncoder".
func (pc *PluginConfig) CipherEnabled() bool {
	return pc.CipherPlugin != nil && !pc.CipherPlugin.Disabled
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
	if app.PluginConfig.CipherEnabled() {
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
