package domain

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/esnet/gdg/internal/storage"
	resourceTypes "github.com/esnet/gdg/pkg/config/domain"
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
	contextName              string
	secureAuth               *SecureModel
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
	UserName                 string                `mapstructure:"user_name" yaml:"user_name"`
	UserSettings             *UserSettings         `mapstructure:"user" yaml:"user"`
	grafanaAdminEnabled      bool                  `mapstructure:"-" yaml:"-"`
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

// Testing Functions

// TestGetSecureAuth returns a copy of the secure auth model if test env is enabled.
func (s *GrafanaConfig) TestGetSecureAuth() *SecureModel {
	if os.Getenv(path.TestEnvKey) != "1" {
		return nil
	}
	d := *s.secureAuth
	return &d
}

// TestSetSecureAuth sets the secure authentication model for testing purposes.
func (s *GrafanaConfig) TestSetSecureAuth(auth SecureModel) error {
	if os.Getenv(path.TestEnvKey) != "1" {
		return nil
	}
	s.secureAuth = &auth
	return nil
}

// End Testing Functions

// loadAuthData
func loadData[T any](securePath string, obj *T) (*T, error) {
	v := viper.New()
	v.SetConfigFile(securePath)
	formats := []string{".yaml", ".yml", ".json"}
	var outerErr error
	for _, ext := range formats {
		filename := securePath + ext
		if _, err := os.Stat(filename); err == nil {
			// File exists
			v.SetConfigFile(filename)
			if readError := v.ReadInConfig(); readError != nil {
				continue // Try next extension
			}
			marshErr := v.Unmarshal(&obj)
			if marshErr != nil {
				outerErr = fmt.Errorf("unable to unmarshal secure file %s, readError: %w", securePath, marshErr)
				continue
			}
			return obj, nil

		}
	}
	if outerErr != nil {
		return nil, fmt.Errorf("unable to find secure file %s, err: %w", securePath, outerErr)
	}
	return nil, fmt.Errorf("unable to find secure file %s", securePath)
}

// GetCloudAuth returns a map of cloud authentication credentials loaded from the configured file.
func (s *GrafanaConfig) GetCloudAuth() map[string]string {
	authFile := s.GetCloudAuthLocation()
	m := make(map[string]string)
	if authFile == "" {
		return m
	}
	_, err := loadData(authFile, &m)
	if err != nil {
		slog.Warn(fmt.Sprintf("%v, falling back on Env settings. Please set '%s' and '%s' if you haven't done so already",
			err, storage.CloudKey, storage.CloudSecret))
	}

	return m
}

// getSecureAuth returns the parsed secure authentication model,
// loading from YAML, YML or JSON files in order of precedence.
// It caches the result for subsequent calls.
func (s *GrafanaConfig) getSecureAuth() *SecureModel {
	if s.secureAuth != nil {
		return s.secureAuth
	}

	authFile := s.GetAuthLocation()
	obj, err := loadData(authFile, new(SecureModel))
	if err != nil {
		slog.Error(err.Error())
		return s.secureAuth
	}
	s.secureAuth = obj
	return s.secureAuth
}

// UpdateSecureModel updates the secure model using the supplied function, if secure auth is present.
func (s *GrafanaConfig) UpdateSecureModel(fn func(string) (string, error)) {
	secureAuth := s.getSecureAuth()
	if secureAuth == nil {
		return
	}
	secureAuth.UpdateSecureModel(fn)
}

// GetPassword returns the password, respecting environment variable override if set.
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

// GetAPIToken returns the API token, checking for an environment variable override before falling back to stored credentials.
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

// GetCloudAuthLocation returns the file path to the cloud auth credentials for this config.
func (s *GrafanaConfig) GetCloudAuthLocation() string {
	securePath := s.SecureLocation()
	if s.Storage == "" {
		return ""
	}
	name := fmt.Sprintf("%s_%s", CloudAuthPrefix, s.Storage)
	authFile := filepath.Join(securePath, name)
	return authFile
}

// GetAuthLocation returns the file path for the authentication token based on the
// secure location and context name.
func (s *GrafanaConfig) GetAuthLocation() string {
	securePath := s.SecureLocation()
	name := fmt.Sprintf("%s_%s", AuthPrefix, s.contextName)
	authFile := filepath.Join(securePath, name)
	return authFile
}

// SecureLocation returns the resolved path for secure resources, using override or default.
func (s *GrafanaConfig) SecureLocation() string {
	if s.SecureLocationOverride == "" {
		return s.GetPath(resourceTypes.SecureSecretsResource, "")
	}

	// if path starts with a slash assume it's an absolute path
	if s.SecureLocationOverride[0] == filepath.Separator {
		return s.SecureLocationOverride
	}
	fullPah := filepath.Join(s.OutputPath, s.SecureLocationOverride)
	fullPah = filepath.Clean(fullPah)
	return fullPah
}

// getFilter returns the dashFilter associated with this GrafanaConfig,
// creating it if necessary.
func (s *GrafanaConfig) getFilter() *dashFilter {
	if s.filterFolder == nil {
		s.filterFolder = &dashFilter{}
	}
	return s.filterFolder
}

// SetFilterFolder sets the name filter for folder queries in GrafanaConfig.
func (s *GrafanaConfig) SetFilterFolder(folderFilter string) {
	s.SetUseFilters()
	filter := s.getFilter()
	filter.Name = folderFilter
}

// ClearFilters disables any filter that may have been set on the GrafanaConfig,
// resetting UseFilter to false and clearing the Name field.
func (s *GrafanaConfig) ClearFilters() {
	filter := s.getFilter()
	filter.UseFilter = false
	filter.Name = ""
}

// SetUseFilters enables filter usage by setting UseFilter to true in GrafanaConfig.
func (s *GrafanaConfig) SetUseFilters() {
	filter := s.getFilter()
	filter.UseFilter = true
	s.filterFolder = filter
}

// IsFilterSet reports whether a filter is set for the configuration.
func (s *GrafanaConfig) IsFilterSet() bool {
	return s.getFilter().UseFilter
}

// GetURL returns the Grafana URL, trimmed of whitespace and guaranteed to end with a slash.
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
