package api

import (
	"regexp"
	"strings"

	"github.com/netsage-project/grafana-dashboard-manager/apphelpers"
	"github.com/thoas/go-funk"
)

//Currently supported filters
const (
	DashFilter   = "DashFilter"
	FolderFilter = "FolderFilter"
	Name         = "Name"
)

type Filter interface {
	GetTypes() []string                    //Return list of active Filter Types
	GetFilter(key string) string           //Get the Filter value
	AddFilter(key, value string)           //Add a filter query
	Validate(items map[string]string) bool //Validate if Entry is valid
	GetFolders() []string                  //List of supported folders if Any
}

type BaseFilter struct {
	Filters map[string]string
}

//GetTypes returns all the current keys for the configured Filter
func (s BaseFilter) GetTypes() []string {
	keys := funk.Keys(s.Filters)
	return keys.([]string)
}

//GetFilter returns the value of the filter
func (s BaseFilter) GetFilter(key string) string {
	if val, ok := s.Filters[key]; ok {
		return val
	}
	return ""
}

//AddFilter adds a filter and the corresponding value
func (s BaseFilter) AddFilter(key, value string) {
	s.Filters[key] = value
}

func (s *BaseFilter) Init() {
	s.Filters = make(map[string]string)
}

type DashboardFilter struct {
	quoteRegex *regexp.Regexp
	BaseFilter
}

//NewDashboardFilter creates a new dashboard filter
func NewDashboardFilter() *DashboardFilter {
	s := DashboardFilter{}
	s.init()
	return &s

}

func (s *DashboardFilter) init() {
	s.BaseFilter.Init()
	s.quoteRegex, _ = regexp.Compile("['\"]+")
}

//GetFolders splits the comma delimited folder list and returns a slice
func (s *DashboardFilter) GetFolders() []string {
	if s.GetFilter(FolderFilter) == "" {
		return apphelpers.GetCtxDefaultGrafanaConfig().GetMonitoredFolders()
	}
	folderFilter := s.GetFilter(FolderFilter)
	folderFilter = s.quoteRegex.ReplaceAllString(folderFilter, "")
	s.AddFilter(FolderFilter, folderFilter)

	return strings.Split(folderFilter, ",")
}

func (s DashboardFilter) validateDashboard(dashUid string) bool {
	if s.GetFilter(DashFilter) == "" {
		return true
	}
	return dashUid == s.GetFilter(DashFilter)
}

func (s DashboardFilter) Validate(items map[string]string) bool {
	var folderCheck, dashboardCheck bool
	//Check folder
	if folderFilter, ok := items[FolderFilter]; ok {
		folderCheck = s.validateFolder(folderFilter)
	} else {
		folderCheck = true
	}

	//check Dash
	if dashFilter, ok := items[DashFilter]; ok {
		dashboardCheck = s.validateDashboard(dashFilter)
	} else {
		dashboardCheck = true
	}
	return folderCheck && dashboardCheck
}

func (s DashboardFilter) validateFolder(folder string) bool {
	if s.GetFilter(FolderFilter) == "" {
		return true
	}
	return folder == s.GetFilter(FolderFilter)
}

type DatasourceFilter struct {
	BaseFilter
}

//GetFolders return empty list since Folders NA for datasources
func (s DatasourceFilter) GetFolders() []string {
	return []string{}
}

//Validate returns true if mapped values are valid
func (s DatasourceFilter) Validate(items map[string]string) bool {
	if s.GetFilter(Name) == "" {
		return true
	}
	return items[Name] == s.GetFilter(Name)

}
