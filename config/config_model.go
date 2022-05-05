package config

import (
	"errors"
	"github.com/thoas/go-funk"
	"path"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

type ResourceType string

const (
	UserResource              = "users"
	DashboardResource         = "dashboards"
	DataSourceResource        = "datasources"
	FolderResource            = "folders"
	AlertNotificationResource = "alertnotifications"
)

//GrafanaConfig model wraps auth and watched list for grafana
type GrafanaConfig struct {
	AdminEnabled       bool
	IgnoreFilters      bool                `yaml:"ignore_filters"`
	URL                string              `yaml:"url"`
	APIToken           string              `yaml:"token"`
	UserName           string              `yaml:"user_name"`
	Password           string              `yaml:"password"`
	Organization       string              `yaml:"organization"`
	MonitoredFolders   []string            `yaml:"watched"`
	DefaultDataSource  *GrafanaDataSource  `yaml:"-"`
	DataSourceSettings *DataSourceSettings `yaml:"datasources"`
	OutputPath         string              `yaml:"output_path"`
}

type DataSourceSettings struct {
	Filters     *DataSourceFilters            `yaml:"filters"`
	Credentials map[string]*GrafanaDataSource `yaml:"credentials"`
}

func (ds DataSourceSettings) FiltersEnabled() bool {
	return ds.Filters != nil
}

func (ds *DataSourceSettings) GetCredentials(dataSourceName string) (*GrafanaDataSource, error) {
	key := strings.ToLower(dataSourceName)
	if val, ok := ds.Credentials[key]; ok {
		return val, nil
	} else {
		return nil, errors.New("no valid configuration found, falling back on default")
	}
}

type DataSourceFilters struct {
	NameExclusions  string   `yaml:"name_exclusions"`
	DataSourceTypes []string `yaml:"valid_types"`
	pattern         *regexp.Regexp
}

func (filter DataSourceFilters) ValidDataType(dataType string) bool {
	if len(filter.DataSourceTypes) == 0 {
		return true
	}
	result := funk.Contains(filter.DataSourceTypes, dataType)
	log.Debugf("datatype %s is %t", dataType, result)
	return funk.Contains(filter.DataSourceTypes, dataType)
}

func (filter *DataSourceFilters) ValidName(name string) bool {
	if filter.pattern == nil {
		var err error
		filter.pattern, err = regexp.Compile(filter.NameExclusions)
		if err != nil {
			log.Warning("Could not compile datasource filter.  Aborting")
			filter.pattern = nil
			return false
		}
	}
	result := filter.pattern.Match([]byte(name))
	log.Debugf("%s was %t valid", name, !result)
	return !result
}

func (s *GrafanaConfig) GetDataSourceSettings() *DataSourceSettings {
	if s.DataSourceSettings == nil {
		s.DataSourceSettings = &DataSourceSettings{}
	}
	return s.DataSourceSettings
}

func (s *GrafanaConfig) GetDashboardOutput() string {
	return path.Join(s.OutputPath, DashboardResource)
}

func (s *GrafanaConfig) GetDataSourceOutput() string {
	return path.Join(s.OutputPath, DataSourceResource)
}

func (s *GrafanaConfig) GetAlertNotificationOutput() string {
	return path.Join(s.OutputPath, AlertNotificationResource)
}

func (s *GrafanaConfig) GetUserOutput() string {
	return path.Join(s.OutputPath, UserResource)
}

func (s *GrafanaConfig) GetFolderOutput() string {
	return path.Join(s.OutputPath, FolderResource)
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
	source, err := s.GetDataSourceSettings().GetCredentials(dataSourceName)
	if err == nil {
		return source, nil
	}

	log.Infof("No datasource credentials found for '%s', falling back on default", dataSourceName)
	return s.GetDefaultCredentials(), nil
}

//GetCredentialByUrl attempts to match URL by regex
func (s *GrafanaConfig) GetCredentialByUrl(fullUrl string) (*GrafanaDataSource, error) {
	for key, val := range s.GetDataSourceSettings().Credentials {
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
		if val, ok := s.GetDataSourceSettings().Credentials["default"]; ok {
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
