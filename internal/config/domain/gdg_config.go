package domain

import (
	"strings"
)

// GDGAppConfiguration is the configuration for the application
type GDGAppConfiguration struct {
	ContextName   string                       `mapstructure:"context_name" yaml:"context_name"`
	StorageEngine map[string]map[string]string `mapstructure:"storage_engine" yaml:"storage_engine"`
	Contexts      map[string]*GrafanaConfig    `mapstructure:"contexts" yaml:"contexts"`
	Global        *AppGlobals                  `mapstructure:"global" yaml:"global"`
}

func (app *GDGAppConfiguration) GetContext() string {
	return strings.ToLower(app.ContextName)
}

func (app *GDGAppConfiguration) GetContexts() map[string]*GrafanaConfig {
	return app.Contexts
}

func (app *GDGAppConfiguration) GetAppGlobals() *AppGlobals {
	if app.Global == nil {
		app.Global = &AppGlobals{}
	}
	return app.Global
}
