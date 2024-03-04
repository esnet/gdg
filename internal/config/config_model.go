package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gosimple/slug"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/tidwall/gjson"
	"log/slog"
	"path"
	"regexp"
)

type ResourceType string

const (
	ConnectionPermissionResource = "connections-permissions"
	ConnectionResource           = "connections"
	DashboardResource            = "dashboards"
	FolderPermissionResource     = "folders-permissions"
	FolderResource               = "folders"
	LibraryElementResource       = "libraryelements"
	OrganizationResource         = "organizations"
	OrganizationMetaResource     = "org"
	TeamResource                 = "teams"
	UserResource                 = "users"
	TemplatesResource            = "templates"
	SecureSecretsResource        = "secure"
)

var orgNamespacedResource = map[ResourceType]bool{
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
		orgName := slug.Make(Config().GetDefaultGrafanaConfig().GetOrganizationName())
		return path.Join(basePath, fmt.Sprintf("%s_%s", OrganizationMetaResource, orgName), s.String())

	}
	return path.Join(basePath, s.String())
}

// FiltersEnabled returns true if the filters are enabled for the resource type
func (ds *ConnectionSettings) FiltersEnabled() bool {
	return ds.FilterRules != nil
}

// GetCredentials returns the credentials for the connection
func (ds *ConnectionSettings) GetCredentials(connectionEntity models.AddDataSourceCommand, path string) (*GrafanaConnection, error) {
	data, err := json.Marshal(connectionEntity)
	if err != nil {
		slog.Warn("Unable to marshall Connection, unable to fetch credentials")
		return nil, fmt.Errorf("unable to marshall Connection, unable to fetch credentials")
	}
	//Get SecureData based on New Matching Rules
	parser := gjson.ParseBytes(data)
	for _, entry := range ds.MatchingRules {
		//Check Rules
		valid := true
		for _, rule := range entry.Rules {
			fieldObject := parser.Get(rule.Field)
			if !fieldObject.Exists() {
				slog.Warn("Unable to find a matching field in datasource, skipping validation rule", "filedName", rule.Field)
				valid = false
				continue
			}
			fieldValue := fieldObject.String()
			p, err := regexp.Compile(rule.Regex)
			if err != nil {
				slog.Warn("Unable to compile regex to match against field, skipping validation", "regex", rule.Regex, "fieldName", rule.Field)
				valid = false
			}
			if !p.Match([]byte(fieldValue)) {
				valid = false
				break
			}
		}
		if valid {
			return entry.GetAuth(path)
		}

	}

	return nil, errors.New("no valid configuration found, falling back on default")
}

// IsExcluded returns true if the item should be excluded from the connection List
func (ds *ConnectionSettings) IsExcluded(item interface{}) bool {
	data, err := json.Marshal(item)
	if err != nil {
		slog.Warn("Unable to serialize object, cannot validate")
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
				slog.Warn("Invalid regex for filter rule", "field", field.Field)
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
	if s.ConnectionSettings == nil {
		s.ConnectionSettings = &ConnectionSettings{}
	}
	return s.ConnectionSettings
}

// GetPath returns the path of the resource type
func (s *GrafanaConfig) GetPath(r ResourceType) string {
	return r.GetPath(s.OutputPath)
}

// GetOrgMonitoredFolders return the OrganizationMonitoredFolders that override a given Org
func (s *GrafanaConfig) GetOrgMonitoredFolders(orgName string) []string {
	for _, item := range s.MonitoredFoldersOverride {
		if item.OrganizationName == orgName && len(item.Folders) > 0 {
			return item.Folders
		}
	}

	return nil
}

// GetMonitoredFolders return a list of the monitored folders alternatively returns the "General" folder.
func (s *GrafanaConfig) GetMonitoredFolders() []string {
	orgFolders := s.GetOrgMonitoredFolders(s.GetOrganizationName())
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

}

// IsAdminEnabled returns true if the admin is set, represents a GrafanaAdmin
func (s *GrafanaConfig) IsAdminEnabled() bool {
	return s.adminEnabled
}

// GetCredentials return credentials for a given datasource or falls back on default value
func (s *GrafanaConfig) GetCredentials(dataSourceName models.AddDataSourceCommand, location string) (*GrafanaConnection, error) {
	source, err := s.GetDataSourceSettings().GetCredentials(dataSourceName, location)
	if err == nil {
		return source, nil
	}

	return nil, fmt.Errorf("no datasource credentials found for '%s', falling back on default", dataSourceName.Name)
}

// IsEnterprise Returns true when enterprise is enabled
func (s *GrafanaConfig) IsEnterprise() bool {
	return s.EnterpriseSupport
}
