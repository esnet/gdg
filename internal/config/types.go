package config

import (
	"github.com/spf13/viper"
	"regexp"
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

// GrafanaConfig model wraps auth and watched list for grafana
type GrafanaConfig struct {
	Storage            string              `mapstructure:"storage" yaml:"storage"`
	AdminEnabled       bool                `mapstructure:"-" yaml:"-"`
	EnterpriseSupport  bool                `mapstructure:"enterprise_support" yaml:"enterprise_support"`
	URL                string              `mapstructure:"url" yaml:"url"`
	APIToken           string              `mapstructure:"token" yaml:"token"`
	UserName           string              `mapstructure:"user_name" yaml:"user_name"`
	Password           string              `mapstructure:"password" yaml:"password"`
	Organization       string              `mapstructure:"organization" yaml:"organization"`
	MonitoredFolders   []string            `mapstructure:"watched" yaml:"watched"`
	DataSourceSettings *ConnectionSettings `mapstructure:"connections" yaml:"connections"`
	FilterOverrides    *FilterOverrides    `mapstructure:"filter_override" yaml:"filter_override"`
	OutputPath         string              `mapstructure:"output_path" yaml:"output_path"`
}

type ConnectionSettings struct {
	FilterRules   []MatchingRule     `mapstructure:"exclude_filters" yaml:"exclude_filters,omitempty"`
	MatchingRules []RegexMatchesList `mapstructure:"credential_rules" yaml:"credential_rules,omitempty"`
}

type RegexMatchesList struct {
	Rules []MatchingRule     `mapstructure:"rules" yaml:"rules,omitempty"`
	Auth  *GrafanaConnection `mapstructure:"auth" yaml:"auth,omitempty"`
}

type CredentialRule struct {
	RegexMatchesList
	Auth *GrafanaConnection `mapstructure:"auth" yaml:"auth,omitempty"`
}

type MatchingRule struct {
	Field     string `yaml:"field,omitempty"`
	Regex     string `yaml:"regex,omitempty"`
	Inclusive bool   `yaml:"inclusive,omitempty"`
}

type FilterOverrides struct {
	IgnoreDashboardFilters bool `yaml:"ignore_dashboard_filters"`
}

type ConnectionFilters struct {
	NameExclusions  string   `yaml:"name_exclusions"`
	ConnectionTypes []string `yaml:"valid_types"`
	pattern         *regexp.Regexp
}

// GrafanaConnection Default datasource credentials
type GrafanaConnection struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}
