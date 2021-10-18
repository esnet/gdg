package config

import (
	"errors"
	"path"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

//GrafanaConfig model wraps auth and watched list for grafana
type GrafanaConfig struct {
	AdminEnabled       bool
	IgnoreFilters      bool                          `yaml:"ignore_filters"`
	URL                string                        `yaml:"url"`
	APIToken           string                        `yaml:"token"`
	UserName           string                        `yaml:"user_name"`
	Password           string                        `yaml:"password"`
	Organization       string                        `yaml:"organization"`
	MonitoredFolders   []string                      `yaml:"watched"`
	DefaultDataSource  *GrafanaDataSource            `yaml:"-"`
	DataSourceSettings map[string]*GrafanaDataSource `yaml:"datasources"`
	OutputPath         string                        `yaml:"output_path"`
}

func (s *GrafanaConfig) GetDashboardOutput() string {
	return path.Join(s.OutputPath, "dashboards")
}

func (s *GrafanaConfig) GetDataSourceOutput() string {
	return path.Join(s.OutputPath, "datasources")
}

//GetMonitoredFolders return a list of the monitored folders alternatively returns the "General" folder.
func (s *GrafanaConfig) GetMonitoredFolders() []string {
	if len(s.MonitoredFolders) == 0 {
		return []string{"General"}
	}

	return s.MonitoredFolders
}

//GetCredentials return credentials for a given datasource or falls back on default value
func (s *GrafanaConfig) GetCredentials(dataSourceName string) (*GrafanaDataSource, error) {
	key := strings.ToLower(dataSourceName)
	if val, ok := s.DataSourceSettings[key]; ok {
		return val, nil
	} else {
		log.Infof("No datasource credentials found for '%s', falling back on default", dataSourceName)
		return s.GetDefaultCredentials(), errors.New("no valid configuration found, falling back on default")
	}
}

//GetCredentialByUrl attempts to match URL by regex
func (s *GrafanaConfig) GetCredentialByUrl(fullUrl string) (*GrafanaDataSource, error) {
	for key, val := range s.DataSourceSettings {
		if val.UrlRegex != "" {
			r, err := regexp.Compile(val.UrlRegex)
			if err != nil {
				log.Warnf("Invalid regex for DS: %s using regex: %s", key, val.UrlRegex)
				continue
			}
			match := r.MatchString(fullUrl)
			if match {
				return val, nil
			}
		}
	}
	log.Warn("No valid regex detected, falling back on default")
	return s.GetDefaultCredentials(), errors.New("no valid configuration found, falling back on default")

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
	UrlRegex string `yaml:"url_regex"`
}
