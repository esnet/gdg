package config

import (
	"github.com/spf13/viper"
	"regexp"
)

type Configuration struct {
	defaultConfig *viper.Viper
	AppConfig     *AppConfig
}

// AppGlobals is the global configuration for the application
type AppGlobals struct {
	Debug           bool `mapstructure:"debug" yaml:"debug"`
	IgnoreSSLErrors bool `mapstructure:"ignore_ssl_errors" yaml:"ignore_ssl_errors"`
}

// AppConfig is the configuration for the application
type AppConfig struct {
	ContextName   string                       `mapstructure:"context_name" yaml:"context_name"`
	StorageEngine map[string]map[string]string `mapstructure:"storage_engine" yaml:"storage_engine"`
	Contexts      map[string]*GrafanaConfig    `mapstructure:"contexts" yaml:"contexts"`
	Global        *AppGlobals                  `mapstructure:"global" yaml:"global"`
}

// GrafanaConfig model wraps auth and watched list for grafana
type GrafanaConfig struct {
	Storage            string              `mapstructure:"storage" yaml:"storage"`
	adminEnabled       bool                `mapstructure:"-" yaml:"-"`
	EnterpriseSupport  bool                `mapstructure:"enterprise_support" yaml:"enterprise_support"`
	URL                string              `mapstructure:"url" yaml:"url"`
	APIToken           string              `mapstructure:"token" yaml:"token"`
	UserName           string              `mapstructure:"user_name" yaml:"user_name"`
	Password           string              `mapstructure:"password" yaml:"password"`
	OrganizationId     int64               `mapstructure:"organization_id" yaml:"organization_id"`
	MonitoredFolders   []string            `mapstructure:"watched" yaml:"watched"`
	DataSourceSettings *ConnectionSettings `mapstructure:"connections" yaml:"connections"`
	//Datasources are deprecated, please use Connections
	LegacyConnectionSettings map[string]interface{} `mapstructure:"datasources" yaml:"datasources"`
	FilterOverrides          *FilterOverrides       `mapstructure:"filter_override" yaml:"filter_override"`
	OutputPath               string                 `mapstructure:"output_path" yaml:"output_path"`
}

// GetOrganizationId returns the id of the organization (defaults to 1 if unset)
func (s *GrafanaConfig) GetOrganizationId() int64 {
	if s.OrganizationId > 1 {
		return s.OrganizationId
	}
	return 1
}

// SetAdmin sets true if user has admin permissions
func (s *GrafanaConfig) SetAdmin(admin bool) {
	s.adminEnabled = admin
}

// IsBasicAuth returns true if user has basic auth enabled
func (s *GrafanaConfig) IsBasicAuth() bool {
	if s.UserName != "" && s.Password != "" {
		return true
	}

	return false
}

// ConnectionSettings contains Filters and Matching Rules for Grafana
type ConnectionSettings struct {
	FilterRules   []MatchingRule     `mapstructure:"exclude_filters" yaml:"exclude_filters,omitempty"`
	MatchingRules []RegexMatchesList `mapstructure:"credential_rules" yaml:"credential_rules,omitempty"`
}

// RegexMatchesList model wraps regex matches list for grafana
type RegexMatchesList struct {
	Rules []MatchingRule     `mapstructure:"rules" yaml:"rules,omitempty"`
	Auth  *GrafanaConnection `mapstructure:"auth" yaml:"auth,omitempty"`
}

// CredentialRule model wraps regex and auth for grafana
type CredentialRule struct {
	RegexMatchesList
	Auth *GrafanaConnection `mapstructure:"auth" yaml:"auth,omitempty"`
}

// MatchingRule defines a single matching rule for Grafana Connections
type MatchingRule struct {
	Field     string `yaml:"field,omitempty"`
	Regex     string `yaml:"regex,omitempty"`
	Inclusive bool   `yaml:"inclusive,omitempty"`
}

// FilterOverrides model wraps filter overrides for grafana
type FilterOverrides struct {
	IgnoreDashboardFilters bool `yaml:"ignore_dashboard_filters"`
}

// ConnectionFilters model wraps connection filters for grafana
type ConnectionFilters struct {
	NameExclusions  string   `yaml:"name_exclusions"`
	ConnectionTypes []string `yaml:"valid_types"`
	pattern         *regexp.Regexp
}

// GrafanaConnection Default connection credentials
type GrafanaConnection struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}
