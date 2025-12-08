package domain

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/esnet/gdg/pkg/test_tooling/path"
	"github.com/spf13/viper"
)

type DashboardSettings struct {
	IgnoreFilters bool `yaml:"ignore_filters" mapstructure:"ignore_filters" `
}

type dashFilter struct {
	Name      string
	UseFilter bool
}

// GrafanaConfig model wraps auth and watched list for grafana
type GrafanaConfig struct {
	contextName string
	secureAuth  *SecureModel
	// APIToken                 string                `mapstructure:"token" yaml:"token"`
	ConnectionSettings       *ConnectionSettings   `mapstructure:"connections" yaml:"connections"`
	DashboardSettings        *DashboardSettings    `mapstructure:"dashboard_settings" yaml:"dashboard_settings"`
	MonitoredFolders         []string              `mapstructure:"watched" yaml:"watched"`
	filterFolder             *dashFilter           `mapstructure:"-" yaml:"-"`
	MonitoredFoldersOverride []MonitoredOrgFolders `mapstructure:"watched_folders_override" yaml:"watched_folders_override"`
	OrganizationName         string                `mapstructure:"organization_name" yaml:"organization_name"`
	SecureLocationOverride   string                `mapstructure:"secure_location" yaml:"secure_location"`
	OutputPath               string                `mapstructure:"output_path" yaml:"output_path"`
	Storage                  string                `mapstructure:"storage" yaml:"storage"`
	URL                      string                `mapstructure:"url" yaml:"url"`
	UserSettings             *UserSettings         `mapstructure:"user" yaml:"user"`
	grafanaAdminEnabled      bool                  `mapstructure:"-" yaml:"-"`
	UserName                 string                `mapstructure:"user_name" yaml:"user_name"`
}

// Only exposed for testing
func (s *GrafanaConfig) TestGetSecureAuth() *SecureModel {
	if os.Getenv(path.TestEnvKey) != "1" {
		return nil
	}
	d := *s.secureAuth
	return &d
}

type MonitoredOrgFolders struct {
	OrganizationName string   `json:"organization_name" yaml:"organization_name" mapstructure:"organization_name"`
	Folders          []string `json:"folders" yaml:"folders" mapstructure:"folders"`
}

// ConnectionSettings contains Filters and Matching Rules for Grafana
type ConnectionSettings struct {
	FilterRules   []MatchingRule     `mapstructure:"filters" yaml:"filters,omitempty"`
	MatchingRules []RegexMatchesList `mapstructure:"credential_rules" yaml:"credential_rules,omitempty"`
}

// RegexMatchesList model wraps regex matches list for grafana
type RegexMatchesList struct {
	Rules      []MatchingRule `mapstructure:"rules" yaml:"rules,omitempty"`
	SecureData string         `mapstructure:"secure_data" yaml:"secure_data,omitempty"`
}

func (s *GrafanaConfig) getSecureAuth() *SecureModel {
	if s.secureAuth != nil {
		return s.secureAuth
	}

	authFile := s.GetAuthLocation()
	obj := SecureModel{}
	v := viper.New()
	v.SetConfigFile(authFile)
	formats := []string{".yaml", ".yml", ".json"}
	for _, ext := range formats {
		filename := authFile + ext
		if _, err := os.Stat(filename); err == nil {
			// File exists
			v.SetConfigFile(filename)
			if err := v.ReadInConfig(); err != nil {
				continue // Try next extension
			}
			marshErr := v.Unmarshal(&obj)
			if marshErr != nil {
				slog.Error("unable to unmarshal auth file", "file", authFile, "error", marshErr)
			}
			s.secureAuth = &obj
			break
		}
	}

	return s.secureAuth
}

func (s *GrafanaConfig) GetPassword() string {
	secureAuth := s.getSecureAuth()
	if secureAuth == nil {
		return ""
	}
	// Backward compatibility to allow Env Override
	// Get Env Value
	envKey := fmt.Sprintf("GDG_CONTEXTS__%s__PASSWORD", strings.ToUpper(s.contextName))
	val := os.Getenv(envKey)
	if val != "" {
		return val
	}

	return secureAuth.Password
}

func (s *GrafanaConfig) GetAPIToken() string {
	secureAuth := s.getSecureAuth()
	if secureAuth == nil {
		return ""
	}
	// Backward compatibility to allow Env Override
	// Get Env Value
	envKey := fmt.Sprintf("GDG_CONTEXTS__%s__TOKEN", strings.ToUpper(s.contextName))
	val := os.Getenv(envKey)
	if val != "" {
		return val
	}

	return secureAuth.Token
}

func (s *GrafanaConfig) GetAuthLocation() string {
	securePath := s.SecureLocation()
	name := fmt.Sprintf("%s_auth", s.contextName)
	authFile := filepath.Join(securePath, name)
	return authFile
}

// TestSetSecureAuth sets the secure authentication model for testing purposes.
func (s *GrafanaConfig) TestSetSecureAuth(auth SecureModel) error {
	if os.Getenv(path.TestEnvKey) != "1" {
		return nil
	}
	s.secureAuth = &auth
	return nil
}

func (s *GrafanaConfig) SecureLocation() string {
	if s.SecureLocationOverride == "" {
		return s.GetPath(SecureSecretsResource, "")
	}

	// if path starts with a slash assume it's an absolute path
	if s.SecureLocationOverride[0] == filepath.Separator {
		return s.SecureLocationOverride
	}
	fullPah := filepath.Join(s.OutputPath, s.SecureLocationOverride)
	fullPah = filepath.Clean(fullPah)
	return fullPah
}

func (s *GrafanaConfig) getFilter() *dashFilter {
	if s.filterFolder == nil {
		s.filterFolder = &dashFilter{}
	}
	return s.filterFolder
}

func (s *GrafanaConfig) SetFilterFolder(folderFilter string) {
	s.SetUseFilters()
	filter := s.getFilter()
	filter.Name = folderFilter
}

func (s *GrafanaConfig) ClearFilters() {
	filter := s.getFilter()
	filter.UseFilter = false
	filter.Name = ""
}

func (s *GrafanaConfig) SetUseFilters() {
	filter := s.getFilter()
	filter.UseFilter = true
	s.filterFolder = filter
}

func (s *GrafanaConfig) IsFilterSet() bool {
	return s.getFilter().UseFilter
}

// GetURL returns the URL for Grafana, trimming whitespace and adding a trailing slash if not already present.
func (s *GrafanaConfig) GetURL() string {
	if len(s.URL) == 0 {
		return s.URL
	}
	// remove white space
	s.URL = strings.TrimSpace(s.URL)
	// add trailing slash if present
	if s.URL[len(s.URL)-1] != '/' {
		s.URL = s.URL + "/"
	}

	return s.URL
}

// GetOrganizationName returns the id of the organization (defaults to 1 if unset)
func (s *GrafanaConfig) GetOrganizationName() string {
	if s.OrganizationName != "" {
		return s.OrganizationName
	}
	if s.IsBasicAuth() {
		return DefaultOrganizationName
	}
	return "unknown"
}

// SetGrafanaAdmin sets true if user has admin permissions
func (s *GrafanaConfig) SetGrafanaAdmin(admin bool) {
	s.grafanaAdminEnabled = admin
}

// IsBasicAuth returns true if user has basic auth enabled
func (s *GrafanaConfig) IsBasicAuth() bool {
	if s.UserName != "" && s.GetPassword() != "" {
		return true
	}

	return false
}

func NewGrafanaConfig(contextName string) *GrafanaConfig {
	s := &GrafanaConfig{
		contextName: contextName,
	}
	return s
}
