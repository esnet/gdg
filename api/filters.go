package api

import (
	"regexp"
	"strings"

	"github.com/esnet/gdg/apphelpers"
	"github.com/thoas/go-funk"
)

// Currently supported filters
const (
	TagsFilter   = "TagsFilter"
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
	GetTags() []string                     //List of tags
}

type BaseFilter struct {
	Filters map[string]string
}

// GetTypes returns all the current keys for the configured Filter
func (s BaseFilter) GetTypes() []string {
	keys := funk.Keys(s.Filters)
	return keys.([]string)
}

// GetFilter returns the value of the filter
func (s BaseFilter) GetFilter(key string) string {
	if val, ok := s.Filters[key]; ok {
		return val
	}
	return ""
}

// AddFilter adds a filter and the corresponding value
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

// NewDashboardFilter creates a new dashboard filter
func NewDashboardFilter() *DashboardFilter {
	s := DashboardFilter{}
	s.init()
	return &s

}

func (s *DashboardFilter) init() {
	s.BaseFilter.Init()
	s.quoteRegex, _ = regexp.Compile("['\"]+")
}

// GetFolders splits the comma delimited folder list and returns a slice
func (s *DashboardFilter) GetFolders() []string {
	if s.GetFilter(FolderFilter) == "" {
		return apphelpers.GetCtxDefaultGrafanaConfig().GetMonitoredFolders()
	}
	folderFilter := s.GetFilter(FolderFilter)
	folderFilter = s.quoteRegex.ReplaceAllString(folderFilter, "")
	s.AddFilter(FolderFilter, folderFilter)

	return strings.Split(folderFilter, ",")
}

// GetTags returns a list of all tags to filter on
func (s *DashboardFilter) GetTags() []string {
	if s.GetFilter(TagsFilter) == "" {
		return []string{}
	}
	tagsFilter := s.GetFilter(TagsFilter)
	tagsFilter = s.quoteRegex.ReplaceAllString(tagsFilter, "")
	s.AddFilter(TagsFilter, tagsFilter)

	return strings.Split(tagsFilter, ",")
}

func (s *DashboardFilter) validateDashboard(dashUid string) bool {
	if s.GetFilter(DashFilter) == "" {
		return true
	}
	return dashUid == s.GetFilter(DashFilter)
}

func (s *DashboardFilter) Validate(items map[string]string) bool {
	var folderCheck, tagsCheck, dashboardCheck bool
	//Check folder
	if folderFilter, ok := items[FolderFilter]; ok {
		folderCheck = s.validateFolder(folderFilter)
	} else {
		folderCheck = true
	}

	//check tags
	if tagsFilter, ok := items[TagsFilter]; ok {
		tagsCheck = s.validateTags(tagsFilter)
	} else {
		tagsCheck = true
	}

	//check Dash
	if dashFilter, ok := items[DashFilter]; ok {
		dashboardCheck = s.validateDashboard(dashFilter)
	} else {
		dashboardCheck = true
	}
	return folderCheck && tagsCheck && dashboardCheck
}

func (s *DashboardFilter) validateFolder(folder string) bool {
	if s.GetFilter(FolderFilter) == "" {
		return true
	}
	return folder == s.GetFilter(FolderFilter)
}

func (s *DashboardFilter) validateTags(tags string) bool {
	if s.GetFilter(TagsFilter) == "" {
		return true
	}
	return tags == s.GetFilter(TagsFilter)
}

type DatasourceFilter struct {
	BaseFilter
}

// GetFolders return empty list since Folders NA for datasources
func (s DatasourceFilter) GetFolders() []string {
	return []string{}
}

// GetTags return empty list since Tags NA for datasources
func (s DatasourceFilter) GetTags() []string {
	return []string{}
}

// Validate returns true if mapped values are valid
func (s DatasourceFilter) Validate(items map[string]string) bool {
	if s.GetFilter(Name) == "" {
		return true
	}
	return items[Name] == s.GetFilter(Name)

}
