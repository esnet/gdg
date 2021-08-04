package config

import (
	"strings"

	log "github.com/sirupsen/logrus"
)

//GrafanaConfig model wraps auth and watched list for grafana
type GrafanaConfig struct {
	AdminEnabled       bool
	URL                string                        `yaml:"url"`
	APIToken           string                        `yaml:"token"`
	UserName           string                        `yaml:"user_name"`
	Password           string                        `yaml:"password"`
	Organization       string                        `yaml:"organization"`
	MonitoredFolders   []string                      `yaml:"watched"`
	DefaultDataSource  *GrafanaDataSource            `yaml:"-"`
	DataSourceSettings map[string]*GrafanaDataSource `yaml:"datasources"`
	OutputDashboard    string                        `yaml:"dashboards_output"`
	OutputDataSource   string                        `yaml:"datasources_output"`
}

//GetMonitoredFolders return a list of the monitored folders alternatively returns the "General" folder.
func (s *GrafanaConfig) GetMonitoredFolders() []string {
	if len(s.MonitoredFolders) == 0 {
		return []string{"General"}
	}

	return s.MonitoredFolders
}

//GetCredentials return credentials for a given datasource or falls back on default value
func (s *GrafanaConfig) GetCredentials(dataSourceName string) *GrafanaDataSource {
	key := strings.ToLower(dataSourceName)
	if val, ok := s.DataSourceSettings[key]; ok {
		return val
	} else {
		log.Infof("No datasource credentials found for '%s', falling back on default", dataSourceName)
		return s.GetDefaultCredentials()
	}

}

//GetDefaultCredentials returns the default credentials
func (s *GrafanaConfig) GetDefaultCredentials() *GrafanaDataSource {
	if s.DefaultDataSource == nil {
		if val, ok := s.DataSourceSettings["default"]; ok {
			s.DefaultDataSource = val
		} else {
			log.Warn("No default credentials set, assuming no auth required")
		}
	}

	return s.DefaultDataSource
}

//Default datasource credentials
type GrafanaDataSource struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}
