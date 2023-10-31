package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/esnet/grafana-swagger-api-golang/goclient/models"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"os"
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
	OrganizationMetaResource     = "org"
	TeamResource                 = "teams"
	UserResource                 = "users"
)

var orgNamespacedResource = map[ResourceType]bool{
	AlertNotificationResource:    true,
	ConnectionPermissionResource: true,
	ConnectionResource:           true,
	DashboardResource:            true,
	FolderPermissionResource:     true,
	FolderResource:               true,
	LibraryElementResource:       true,
	TeamResource:                 true,
}

// isNamespaced returns true if the resource type is namespaced
func (s *ResourceType) isNamespaced() bool {
	return orgNamespacedResource[*s]
}

// String returns the string representation of the resource type
func (s *ResourceType) String() string {
	return string(*s)
}

// GetPath returns the path of the resource type, if Namespaced, will delimit the path by org Id
func (s *ResourceType) GetPath(basePath string) string {
	if s.isNamespaced() {
		orgId := Config().GetDefaultGrafanaConfig().GetOrganizationId()
		return path.Join(basePath, fmt.Sprintf("%s_%d", OrganizationMetaResource, orgId), s.String())

	}
	return path.Join(basePath, s.String())
}

// FiltersEnabled returns true if the filters are enabled for the resource type
func (ds *ConnectionSettings) FiltersEnabled() bool {
	return ds.FilterRules != nil
}

// GetCredentials returns the credentials for the connection
func (ds *ConnectionSettings) GetCredentials(connectionEntity models.AddDataSourceCommand) (*GrafanaConnection, error) {
	data, err := json.Marshal(connectionEntity)
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

// IsExcluded returns true if the item should be excluded from the connection List
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

// GetFilterOverrides returns the filter overrides for the connection
func (s *GrafanaConfig) GetFilterOverrides() *FilterOverrides {
	if s.FilterOverrides == nil {
		s.FilterOverrides = &FilterOverrides{IgnoreDashboardFilters: false}
	}
	return s.FilterOverrides
}

// GetDataSourceSettings returns the datasource settings for the connection
func (s *GrafanaConfig) GetDataSourceSettings() *ConnectionSettings {
	if s.DataSourceSettings == nil {
		s.DataSourceSettings = &ConnectionSettings{}
	}
	return s.DataSourceSettings
}

// GetPath returns the path of the resource type
func (s *GrafanaConfig) GetPath(r ResourceType) string {
	return r.GetPath(s.OutputPath)
}

// GetDashboardOutput returns the path of the dashboards output
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

// GetOrgMonitoredFolders return the OrganizationMonitoredFolders that override a given Org
func (s *GrafanaConfig) GetOrgMonitoredFolders(orgId int64) []string {
	for _, item := range s.MonitoredFoldersOverride {
		if item.OrganizationId == orgId && len(item.Folders) > 0 {
			return item.Folders
		}
	}

	return nil
}

// GetMonitoredFolders return a list of the monitored folders alternatively returns the "General" folder.
func (s *GrafanaConfig) GetMonitoredFolders() []string {
	orgFolders := s.GetOrgMonitoredFolders(s.OrganizationId)
	if len(orgFolders) > 0 {
		return orgFolders
	}
	if len(s.MonitoredFolders) == 0 {
		return []string{"General"}
	}

	return s.MonitoredFolders
}

// Validate will return terminate if any deprecated configuration is found.
func (s *GrafanaConfig) Validate() {
	if len(s.LegacyConnectionSettings) > 0 {
		log.Fatal("Using 'datasources' is now deprecated, please use 'connections' instead")
	}
	//Validate Connections
	//TODO: remove code after next release
	legacyCheck := s.GetPath(LegacyConnections)
	if _, err := os.Stat(legacyCheck); !os.IsNotExist(err) {
		log.Fatalf("Your export contains a datasource directry which is deprecated.  Please remove or "+
			"rename directory to '%s'", ConnectionResource)
	}

}

// IsAdminEnabled returns true if the admin is set, represents a GrafanaAdmin
func (s *GrafanaConfig) IsAdminEnabled() bool {
	return s.adminEnabled
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
