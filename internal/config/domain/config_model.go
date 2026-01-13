package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"regexp"

	resourceTypes "github.com/esnet/gdg/pkg/config/domain"
	"github.com/esnet/gdg/pkg/plugins/secure/contract"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/tidwall/gjson"
)

// FiltersEnabled returns true if the filters are enabled for the resource type
func (ds *ConnectionSettings) FiltersEnabled() bool {
	return ds.FilterRules != nil
}

// GetCredentials returns the credentials for the connection
func (ds *ConnectionSettings) GetCredentials(connectionEntity models.AddDataSourceCommand, path string, encoder contract.CipherEncoder) (*GrafanaConnection, error) {
	data, err := json.Marshal(connectionEntity)
	if err != nil {
		slog.Warn("Unable to marshall Connection, unable to fetch credentials")
		return nil, fmt.Errorf("unable to marshall Connection, unable to fetch credentials")
	}
	// Get SecureData based on New Matching Rules
	parser := gjson.ParseBytes(data)
	for _, entry := range ds.MatchingRules {
		// Check Rules
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
			return entry.GetConnectionAuth(path, encoder)
		}

	}

	return nil, errors.New("no valid configuration found, falling back on default")
}

// IsExcluded returns true if the item should be excluded from the connection List
func (ds *ConnectionSettings) IsExcluded(item any) bool {
	data, err := json.Marshal(item)
	if err != nil {
		slog.Warn("Unable to serialize object, cannot validate")
		return true
	}

	// Since filters are always converted only check we need should be this one.
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
			// If inclusive, then the boolean is flipped
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

func (s *GrafanaConfig) GetDashboardSettings() *DashboardSettings {
	if s.DashboardSettings == nil {
		s.DashboardSettings = new(DashboardSettings)
	}
	return s.DashboardSettings
}

// GetConnectionSettings returns the settings for the connection
func (s *GrafanaConfig) GetConnectionSettings() *ConnectionSettings {
	if s.ConnectionSettings == nil {
		s.ConnectionSettings = &ConnectionSettings{}
	}
	return s.ConnectionSettings
}

// GetPath returns the path of the resource type
func (s *GrafanaConfig) GetPath(r resourceTypes.ResourceType, orgName string) string {
	return r.GetPath(s.OutputPath, orgName)
}

// GetUserSettings returns configured UserSettings
func (s *GrafanaConfig) GetUserSettings() *UserSettings {
	if s.UserSettings == nil {
		return &UserSettings{
			RandomPassword: false,
		}
	}
	// Set default values if none are set
	if s.UserSettings.MinLength == 0 {
		s.UserSettings.MinLength = minPasswordLength
	}
	if s.UserSettings.MaxLength == 0 {
		s.UserSettings.MaxLength = maxPasswordLength
	}

	return s.UserSettings
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
func (s *GrafanaConfig) GetMonitoredFolders(ignoreFilterVal bool) []string {
	if s.IsFilterSet() && s.getFilter().Name != "" && !ignoreFilterVal {
		return []string{s.filterFolder.Name}
	}
	orgFolders := s.GetOrgMonitoredFolders(s.GetOrganizationName())
	if len(orgFolders) > 0 {
		return orgFolders
	}
	if len(s.MonitoredFolders) == 0 {
		return []string{"General"}
	}

	return s.MonitoredFolders
}

func (s *GrafanaConfig) Validate() {
	// No-Op at the moment
}

// IsGrafanaAdmin returns true if the admin is set, represents a GrafanaAdmin
func (s *GrafanaConfig) IsGrafanaAdmin() bool {
	return s.grafanaAdminEnabled
}

// GetCredentials return credentials for a given datasource or falls back on default value
func (s *GrafanaConfig) GetCredentials(dataSourceName models.AddDataSourceCommand, location string, encoder contract.CipherEncoder) (*GrafanaConnection, error) {
	source, err := s.GetConnectionSettings().GetCredentials(dataSourceName, location, encoder)
	if err == nil {
		return source, nil
	}

	return nil, fmt.Errorf("no datasource credentials found for '%s', falling back on default", dataSourceName.Name)
}
