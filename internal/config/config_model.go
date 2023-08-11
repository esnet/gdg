package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/esnet/grafana-swagger-api-golang/goclient/models"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"path"
	"regexp"
)

type ResourceType string

const (
	AlertNotificationResource    = "alertnotifications"
	ConnectionPermissionResource = "connections-permissions"
	ConnectionResource           = "connections"
	LegacyConnections            = "datasources"
	DashboardResource            = "dashboards"
	FolderPermissionResource     = "folders-permissions"
	FolderResource               = "folders"
	LibraryElementResource       = "libraryelements"
	OrganizationResource         = "organizations"
	TeamResource                 = "teams"
	UserResource                 = "users"
)

func (s *ResourceType) String() string {
	return string(*s)
}

func (s *ResourceType) GetPath(basePath string) string {
	return path.Join(basePath, s.String())
}

func (ds *ConnectionSettings) FiltersEnabled() bool {
	return ds.FilterRules != nil
}

func (ds *ConnectionSettings) GetCredentials(dataSourceName models.AddDataSourceCommand) (*GrafanaConnection, error) {
	data, err := json.Marshal(dataSourceName)
	if err != nil {
		log.Warn("Unable to marshall Connection, unable to fetch credentials")
		return nil, fmt.Errorf("unable to marshall Connection, unable to fetch credentials")
	}
	//Get Auth based on New Matching Rules
	parser := gjson.ParseBytes(data)
	for _, entry := range ds.MatchingRules {
		//Check Rules
		valid := true
		for _, rule := range entry.Rules {
			fieldObject := parser.Get(rule.Field)
			if !fieldObject.Exists() {
				log.Warnf("Unable to find a field titled: %s in datasource, skipping validation rule", rule.Field)
				valid = false
				continue
			}
			fieldValue := fieldObject.String()
			p, err := regexp.Compile(rule.Regex)
			if err != nil {
				log.Warnf("Unable to compile regex: %s to match against field %s, skipping validation", rule.Regex, rule.Field)
				valid = false
			}
			if !p.Match([]byte(fieldValue)) {
				valid = false
				break
			}
		}
		if valid {
			return entry.Auth, nil
		}

	}

	return nil, errors.New("no valid configuration found, falling back on default")
}

func (ds *ConnectionSettings) IsExcluded(item interface{}) bool {
	data, err := json.Marshal(item)
	if err != nil {
		log.Warn("Unable to serialize object, cannot validate")
		return true
	}

	//Since filters are always converted only check we need should be this one.
	if ds.FilterRules != nil {
		for _, field := range ds.FilterRules {

			fieldParse := gjson.GetBytes(data, field.Field)
			if !fieldParse.Exists() || field.Regex == "" {
				continue
			}

			fieldValue := fieldParse.String()
			p, err := regexp.Compile(field.Regex)
			if err != nil {
				log.Warnf("Invalid regex for filter rule with field: %s", field.Field)
				return true
			}
			match := p.Match([]byte(fieldValue))
			//If inclusive, then the boolean is flipped
			if field.Inclusive {
				match = !match
			}
			if match {
				return match
			}
		}
	}

	return false

}

func (s *GrafanaConfig) GetFilterOverrides() *FilterOverrides {
	if s.FilterOverrides == nil {
		s.FilterOverrides = &FilterOverrides{IgnoreDashboardFilters: false}
	}
	return s.FilterOverrides
}

func (s *GrafanaConfig) GetDataSourceSettings() *ConnectionSettings {
	if s.DataSourceSettings == nil {
		s.DataSourceSettings = &ConnectionSettings{}
	}
	return s.DataSourceSettings
}

func (s *GrafanaConfig) GetPath(r ResourceType) string {
	return r.GetPath(s.OutputPath)
}

func (s *GrafanaConfig) GetDashboardOutput() string {
	return path.Join(s.OutputPath, DashboardResource)
}

func (s *GrafanaConfig) GetDataSourceOutput() string {
	return path.Join(s.OutputPath, ConnectionResource)
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

func (s *GrafanaConfig) GetTeamOutput() string {
	return path.Join(s.OutputPath, TeamResource)
}

// GetMonitoredFolders return a list of the monitored folders alternatively returns the "General" folder.
func (s *GrafanaConfig) GetMonitoredFolders() []string {
	if len(s.MonitoredFolders) == 0 {
		return []string{"General"}
	}

	return s.MonitoredFolders
}

// GetCredentials return credentials for a given datasource or falls back on default value
func (s *GrafanaConfig) GetCredentials(dataSourceName models.AddDataSourceCommand) (*GrafanaConnection, error) {
	source, err := s.GetDataSourceSettings().GetCredentials(dataSourceName)
	if err == nil {
		return source, nil
	}

	return nil, fmt.Errorf("no datasource credentials found for '%s', falling back on default", dataSourceName.Name)
}

// IsEnterprise Returns true when enterprise is enabled
func (s *GrafanaConfig) IsEnterprise() bool {
	return s.EnterpriseSupport
}
