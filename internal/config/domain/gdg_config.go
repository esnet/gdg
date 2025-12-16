package domain

import (
	"strings"

	"github.com/gosimple/slug"
)

// GDGAppConfiguration is the configuration for the application
type GDGAppConfiguration struct {
	ContextName   string                       `mapstructure:"context_name" yaml:"context_name"`
	StorageEngine map[string]map[string]string `mapstructure:"storage_engine" yaml:"storage_engine"`
	Contexts      map[string]*GrafanaConfig    `mapstructure:"contexts" yaml:"contexts"`
	Global        *AppGlobals                  `mapstructure:"global" yaml:"global"`
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

// GetAppGlobals returns the global configuration, initializing it if nil.
func (app *GDGAppConfiguration) GetAppGlobals() *AppGlobals {
	if app.Global == nil {
		app.Global = &AppGlobals{}
	}
	return app.Global
}
