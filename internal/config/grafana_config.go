package config

// GrafanaConfig model wraps auth and watched list for grafana
type GrafanaConfig struct {
	Storage                        string                `mapstructure:"storage" yaml:"storage"`
	grafanaAdminEnabled            bool                  `mapstructure:"-" yaml:"-"`
	EnterpriseSupport              bool                  `mapstructure:"enterprise_support" yaml:"enterprise_support"`
	URL                            string                `mapstructure:"url" yaml:"url"`
	APIToken                       string                `mapstructure:"token" yaml:"token"`
	UserName                       string                `mapstructure:"user_name" yaml:"user_name"`
	Password                       string                `mapstructure:"password" yaml:"password"`
	OrganizationName               string                `mapstructure:"organization_name" yaml:"organization_name"`
	MonitoredFoldersOverride       []MonitoredOrgFolders `mapstructure:"watched_folders_override" yaml:"watched_folders_override"`
	MonitoredFolders               []string              `mapstructure:"watched" yaml:"watched"`
	ConnectionSettings             *ConnectionSettings   `mapstructure:"connections" yaml:"connections"`
	UserSettings                   *UserSettings         `mapstructure:"user" yaml:"user"`
	FilterOverrides                *FilterOverrides      `mapstructure:"filter_override" yaml:"filter_override"`
	OutputPath                     string                `mapstructure:"output_path" yaml:"output_path"`
	DownloadNestedDashboardFolders bool                  `mapstructure:"download_nested_dashboard_folders" yaml:"download_nested_dashboard_folders"`
}

type MonitoredOrgFolders struct {
	OrganizationName string   `json:"organization_name" yaml:"organization_name"`
	Folders          []string `json:"folders" yaml:"folders"`
}

// ConnectionSettings contains Filters and Matching Rules for Grafana
type ConnectionSettings struct {
	FilterRules   []MatchingRule     `mapstructure:"exclude_filters" yaml:"exclude_filters,omitempty"`
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
	return DefaultOrganizationName
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
