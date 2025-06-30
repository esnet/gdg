package domain

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
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
	APIToken                 string                `mapstructure:"token" yaml:"token"`
	ConnectionSettings       *ConnectionSettings   `mapstructure:"connections" yaml:"connections"`
	DashboardSettings        *DashboardSettings    `mapstructure:"dashboard_settings" yaml:"dashboard_settings"`
	MonitoredFolders         []string              `mapstructure:"watched" yaml:"watched"`
	filterFolder             *dashFilter           `mapstructure:"-" yaml:"-"`
	MonitoredFoldersOverride []MonitoredOrgFolders `mapstructure:"watched_folders_override" yaml:"watched_folders_override"`
	OrganizationName         string                `mapstructure:"organization_name" yaml:"organization_name"`
	SecureLocationOverride   string                `mapstructure:"secure_location" yaml:"secure_location"`
	OutputPath               string                `mapstructure:"output_path" yaml:"output_path"`
	Password                 string                `mapstructure:"password" yaml:"password"`
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

func (s *GrafanaConfig) getSecureAuth() *SecureModel {
	if s.secureAuth != nil {
		return s.secureAuth
	}

	securePath := s.SecureLocation()
	name := fmt.Sprintf("%s_auth.json", s.contextName)
	authFile := filepath.Join(securePath, name)
	_, err := os.Stat(authFile)
	if os.IsNotExist(err) {
		return nil
	}
	data, err := os.ReadFile(authFile)
	if err != nil {
		slog.Error("unable to read auth file, falling back on config/env values", "file", authFile, "err", err)
		return nil
	}
	obj := SecureModel{}
	err = json.Unmarshal(data, &obj)
	if err != nil {
		slog.Error("unable to unmarshal auth file", "file", authFile)
	}
	s.secureAuth = &obj
	return s.secureAuth
}

func (s *GrafanaConfig) GetPassword() string {
	secureAuth := s.getSecureAuth()
	if secureAuth == nil || secureAuth.Password == "" {
		return s.Password
	}
	return secureAuth.Password
}

func (s *GrafanaConfig) GetAPIToken() string {
	secureAuth := s.getSecureAuth()
	if secureAuth == nil || secureAuth.Token == "" {
		return s.APIToken
	}
	return secureAuth.Token
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
