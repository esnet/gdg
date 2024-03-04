package config

import (
	"encoding/json"
	"errors"
	"github.com/spf13/viper"
	"log/slog"
	"os"
	"path/filepath"
)

const (
	ViperGdgConfig          = "gdg"
	ViperTemplateConfig     = "template"
	DefaultOrganizationName = "Main Org."
	DefaultOrganizationId   = 1
)

type Configuration struct {
	viperConfiguration map[string]*viper.Viper
	gdgConfig          *GDGAppConfiguration
	templatingConfig   *TemplatingConfig
}

type TemplatingConfig struct {
	Entities TemplateEntities `mapstructure:"entities"`
}

type TemplateEntities struct {
	Dashboards []TemplateDashboards `mapstructure:"dashboards"`
}

type TemplateDashboards struct {
	TemplateName      string                    `mapstructure:"template_name"`
	DashboardEntities []TemplateDashboardEntity `mapstructure:"output"`
}

type TemplateDashboardEntity struct {
	Folder           string                 `mapstructure:"folder"`
	OrganizationName string                 `mapstructure:"organization_name"`
	DashboardName    string                 `mapstructure:"dashboard_name"`
	TemplateData     map[string]interface{} `mapstructure:"template_data"`
}

// AppGlobals is the global configuration for the application
type AppGlobals struct {
	Debug           bool `mapstructure:"debug" yaml:"debug"`
	IgnoreSSLErrors bool `mapstructure:"ignore_ssl_errors" yaml:"ignore_ssl_errors"`
}

// GDGAppConfiguration is the configuration for the application
type GDGAppConfiguration struct {
	ContextName   string                       `mapstructure:"context_name" yaml:"context_name"`
	StorageEngine map[string]map[string]string `mapstructure:"storage_engine" yaml:"storage_engine"`
	Contexts      map[string]*GrafanaConfig    `mapstructure:"contexts" yaml:"contexts"`
	Global        *AppGlobals                  `mapstructure:"global" yaml:"global"`
}

// GrafanaConfig model wraps auth and watched list for grafana
type GrafanaConfig struct {
	Storage                  string                `mapstructure:"storage" yaml:"storage"`
	adminEnabled             bool                  `mapstructure:"-" yaml:"-"`
	EnterpriseSupport        bool                  `mapstructure:"enterprise_support" yaml:"enterprise_support"`
	URL                      string                `mapstructure:"url" yaml:"url"`
	APIToken                 string                `mapstructure:"token" yaml:"token"`
	UserName                 string                `mapstructure:"user_name" yaml:"user_name"`
	Password                 string                `mapstructure:"password" yaml:"password"`
	OrganizationName         string                `mapstructure:"organization_name" yaml:"organization_name"`
	MonitoredFoldersOverride []MonitoredOrgFolders `mapstructure:"watched_folders_override" yaml:"watched_folders_override"`
	MonitoredFolders         []string              `mapstructure:"watched" yaml:"watched"`
	ConnectionSettings       *ConnectionSettings   `mapstructure:"connections" yaml:"connections"`
	FilterOverrides          *FilterOverrides      `mapstructure:"filter_override" yaml:"filter_override"`
	OutputPath               string                `mapstructure:"output_path" yaml:"output_path"`
}

type MonitoredOrgFolders struct {
	OrganizationName string   `json:"organization_name" yaml:"organization_name"`
	Folders          []string `json:"folders" yaml:"folders"`
}

// GetOrganizationName returns the id of the organization (defaults to 1 if unset)
func (s *GrafanaConfig) GetOrganizationName() string {
	if s.OrganizationName != "" {
		return s.OrganizationName
	}
	return DefaultOrganizationName
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
	Rules      []MatchingRule     `mapstructure:"rules" yaml:"rules,omitempty"`
	SecureData string             `mapstructure:"secure_data" yaml:"secure_data,omitempty"`
	LegacyAuth *GrafanaConnection `mapstructure:"auth" yaml:"auth,omitempty" json:"auth,omitempty"`
}

func (r RegexMatchesList) GetAuth(path string) (*GrafanaConnection, error) {
	if r.LegacyAuth != nil && len(*r.LegacyAuth) > 0 {
		slog.Warn("the 'auth' key is deprecated, please update to use 'secure_data'")
	}
	if r.SecureData == "" {
		return r.LegacyAuth, nil
	}
	secretLocation := filepath.Join(path, r.SecureData)
	result := new(GrafanaConnection)
	raw, err := os.ReadFile(secretLocation)
	if err != nil {
		msg := "unable to read secrets at location"
		slog.Error(msg, slog.String("file", secretLocation))
		return nil, errors.New(msg)
	}
	err = json.Unmarshal(raw, result)
	if err != nil {
		msg := "unable to read JSON secrets"
		slog.Error(msg, slog.Any("err", err), slog.String("file", secretLocation))
		return nil, errors.New(msg)
	}

	return result, nil
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
	//	pattern         *regexp.Regexp
}

// GrafanaConnection Default connection credentials
type GrafanaConnection map[string]string

func (g GrafanaConnection) User() string {
	return g["user"]
}

func (g GrafanaConnection) Password() string {
	return g["basicAuthPassword"]
}
