package config

type DashboardSettings struct {
	NestedFolders    bool `mapstructure:"nested_folders" yaml:"nested_folders"`
	IgnoreFilters    bool `yaml:"ignore_filters" mapstructure:"ignore_filters" `
	IgnoreBadFolders bool `yaml:"ignore_bad_folders" mapstructure:"ignore_bad_folders"`
}

// GrafanaConfig model wraps auth and watched list for grafana
type GrafanaConfig struct {
	Storage                  string                `mapstructure:"storage" yaml:"storage"`
	grafanaAdminEnabled      bool                  `mapstructure:"-" yaml:"-"`
	URL                      string                `mapstructure:"url" yaml:"url"`
	APIToken                 string                `mapstructure:"token" yaml:"token"`
	UserName                 string                `mapstructure:"user_name" yaml:"user_name"`
	Password                 string                `mapstructure:"password" yaml:"password"`
	OrganizationName         string                `mapstructure:"organization_name" yaml:"organization_name"`
	MonitoredFoldersOverride []MonitoredOrgFolders `mapstructure:"watched_folders_override" yaml:"watched_folders_override"`
	MonitoredFolders         []string              `mapstructure:"watched" yaml:"watched"`
	ConnectionSettings       *ConnectionSettings   `mapstructure:"connections" yaml:"connections"`
	UserSettings             *UserSettings         `mapstructure:"user" yaml:"user"`
	DashboardSettings        *DashboardSettings    `mapstructure:"dashboard_settings" yaml:"dashboard_settings"`
	OutputPath               string                `mapstructure:"output_path" yaml:"output_path"`
}

type MonitoredOrgFolders struct {
	OrganizationName string   `json:"organization_name" yaml:"organization_name"`
	Folders          []string `json:"folders" yaml:"folders"`
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
	if s.UserName != "" && s.Password != "" {
		return true
	}

	return false
}
