package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/esnet/grafana-swagger-api-golang/goclient/models"
	"github.com/tidwall/gjson"
	"path"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

type ResourceType string

const (
	UserResource              = "users"
	TeamResource              = "teams"
	DashboardResource         = "dashboards"
	DataSourceResource        = "datasources"
	FolderResource            = "folders"
	AlertNotificationResource = "alertnotifications"
	LibraryElementResource    = "libraryelements"
)

func (s *ResourceType) String() string {
	return string(*s)
}

func (s *ResourceType) GetPath(basePath string) string {
	return path.Join(basePath, s.String())
}

// GrafanaConfig model wraps auth and watched list for grafana
type GrafanaConfig struct {
	Storage            string `yaml:"storage"`
	AdminEnabled       bool
	URL                string              `yaml:"url"`
	APIToken           string              `yaml:"token"`
	UserName           string              `yaml:"user_name"`
	Password           string              `yaml:"password"`
	Organization       string              `yaml:"organization"`
	MonitoredFolders   []string            `yaml:"watched"`
	DefaultDataSource  *GrafanaDataSource  `yaml:"-"`
	DataSourceSettings *DataSourceSettings `yaml:"datasources"`
	FilterOverrides    *FilterOverrides    `yaml:"filter_override"`
	OutputPath         string              `yaml:"output_path"`
}

type DataSourceSettings struct {
	Filters       *DataSourceFilters            `yaml:"filters,omitempty"`
	FilterRules   []MatchingRule                `yaml:"exclude_filters,omitempty"`
	Credentials   map[string]*GrafanaDataSource `yaml:"credentials,omitempty"`
	MatchingRules []RegexMatchesList            `yaml:"credential_rules,omitempty"`
}

type RegexMatchesList struct {
	Rules []MatchingRule     `yaml:"rules,omitempty"`
	Auth  *GrafanaDataSource `yaml:"auth,omitempty"`
}

type CredentialRule struct {
	RegexMatchesList
	Auth *GrafanaDataSource `yaml:"auth,omitempty"`
}

type MatchingRule struct {
	Field     string `yaml:"field,omitempty"`
	Regex     string `yaml:"regex,omitempty"`
	Inclusive bool   `yaml:"inclusive,omitempty"`
}

type FilterOverrides struct {
	IgnoreDashboardFilters bool `yaml:"ignore_dashboard_filters"`
}

func (ds *DataSourceSettings) FiltersEnabled() bool {
	//Build FilterRules if they not set and FilterRules are nil
	if ds.Filters != nil && ds.FilterRules == nil {
		ds.convertLegacyFilterEntity()
	}

	return ds.FilterRules != nil || ds.Filters != nil
}

// convertLegacyCredentialsEntity converts previous pattern to the new rule matching one
// Deprecated
func (ds *DataSourceSettings) convertLegacyCredentialsEntity() {

	log.Warn("WARNING: the configuration key 'credentials' is now deprecated, please migrate to using 'credential_rules'")
	log.Warn("A best effort is being done for backward compatibility, but if you use url regex, both name and URL need to match")
	var defaultAuth *RegexMatchesList
	for key, entry := range ds.Credentials {
		if key == "default" {
			defaultAuth = &RegexMatchesList{
				Rules: []MatchingRule{{
					Field: "name",
					Regex: ".*",
				}},
				Auth: &GrafanaDataSource{
					User:     entry.User,
					Password: entry.Password,
				},
			}
			continue
		}

		newEntry := RegexMatchesList{
			Rules: []MatchingRule{
				{
					Field: "name",
					Regex: "(?i)" + key, //make it case insensitive
				},
			},
			Auth: &GrafanaDataSource{
				User:     entry.User,
				Password: entry.Password,
			},
		}
		if entry.UrlRegex != "" {
			newEntry.Rules = append(newEntry.Rules, MatchingRule{
				Field: "url",
				Regex: entry.UrlRegex,
			})
		}

		ds.MatchingRules = append(ds.MatchingRules, newEntry)

	}

	//Add default entry last
	if defaultAuth != nil {
		ds.MatchingRules = append(ds.MatchingRules, *defaultAuth)
	}

}

func (ds *DataSourceSettings) GetCredentials(dataSourceName models.AddDataSourceCommand) (*GrafanaDataSource, error) {
	if len(ds.MatchingRules) == 0 && len(ds.Credentials) > 0 {
		ds.convertLegacyCredentialsEntity()
	}
	data, err := json.Marshal(dataSourceName)
	if err != nil {
		log.Warn("Unable to marshall Datasource, unable to fetch credentials")
		return nil, fmt.Errorf("unable to marshall Datasource, unable to fetch credentials")
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

// convertLegacyFilterEntity construct new matching rules from legacy entity
// Deprecated
func (ds *DataSourceSettings) convertLegacyFilterEntity() {

	log.Warn("WARNING: Filters are deprecated, please move to FilterRules.  This config option " +
		"will stop function in the next release")
	//ds.FilterRules = make([]MatchingRule, 0)
	//Add Name Match
	if ds.Filters.NameExclusions != "" {
		ds.FilterRules = append(ds.FilterRules, MatchingRule{
			Field: "name",
			Regex: ds.Filters.NameExclusions,
		})
		if len(ds.Filters.DataSourceTypes) > 0 {
			dataTypeRule := MatchingRule{
				Field:     "type",
				Inclusive: true,
			}
			bld := strings.Builder{}
			bld.WriteString("")
			for ndx, dataType := range ds.Filters.DataSourceTypes {
				bld.WriteString(dataType)
				if ndx != len(ds.Filters.DataSourceTypes)-1 {
					bld.WriteString("|")
				}
			}
			dataTypeRule.Regex = bld.String()
			ds.FilterRules = append(ds.FilterRules, dataTypeRule)
		}
	}

}

func (ds *DataSourceSettings) IsExcluded(item interface{}) bool {
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

type DataSourceFilters struct {
	NameExclusions  string   `yaml:"name_exclusions"`
	DataSourceTypes []string `yaml:"valid_types"`
	pattern         *regexp.Regexp
}

func (s *GrafanaConfig) GetFilterOverrides() *FilterOverrides {
	if s.FilterOverrides == nil {
		s.FilterOverrides = &FilterOverrides{IgnoreDashboardFilters: false}
	}
	return s.FilterOverrides
}

func (s *GrafanaConfig) GetDataSourceSettings() *DataSourceSettings {
	if s.DataSourceSettings == nil {
		s.DataSourceSettings = &DataSourceSettings{}
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
func (s *GrafanaConfig) GetCredentials(dataSourceName models.AddDataSourceCommand) (*GrafanaDataSource, error) {
	source, err := s.GetDataSourceSettings().GetCredentials(dataSourceName)
	if err == nil {
		return source, nil
	}

	return nil, fmt.Errorf("no datasource credentials found for '%s', falling back on default", dataSourceName.Name)
}

// Default datasource credentials
type GrafanaDataSource struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	UrlRegex string `yaml:"url_regex"`
}
