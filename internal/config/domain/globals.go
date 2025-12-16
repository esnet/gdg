package domain

import (
	"log/slog"
	"time"
)

const (
	AuthPrefix = "auth"
)

// AppGlobals is the global configuration for the application
type AppGlobals struct {
	Debug           bool           `mapstructure:"debug" yaml:"debug"`
	ApiDebug        bool           `mapstructure:"api_debug" yaml:"api_debug"`
	IgnoreSSLErrors bool           `mapstructure:"ignore_ssl_errors" yaml:"ignore_ssl_errors"`
	RetryCount      int            `mapstructure:"retry_count" yaml:"retry_count"`
	RetryDelay      string         `mapstructure:"retry_delay" yaml:"retry_delay"`
	ClearOutput     bool           `mapstructure:"clear_output" yaml:"clear_output"`
	retryTimeout    *time.Duration `mapstructure:"-" yaml:"-"`
}

// GetRetryTimeout returns 100ms, by default otherwise the parsed value
func (app *AppGlobals) GetRetryTimeout() time.Duration {
	defaultBehavior := func() {
		d := time.Millisecond * 100
		app.retryTimeout = &d
	}
	if app.RetryDelay == "" {
		defaultBehavior()
	}
	if app.retryTimeout != nil {
		return *app.retryTimeout
	}
	d, err := time.ParseDuration(app.RetryDelay)
	if err != nil {
		slog.Warn("Unable to parse the retry_delay value.  Falling back on default to 100ms")
		defaultBehavior()
	} else {
		app.retryTimeout = &d
	}

	return *app.retryTimeout
}
