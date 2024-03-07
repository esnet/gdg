package filters

import (
	"encoding/json"
	"github.com/esnet/gdg/internal/config"
	"log/slog"

	"github.com/thoas/go-funk"
	"regexp"
	"strings"
)

// FilterType Currently supported filters
type FilterType string

const (
	TagsFilter    FilterType = "TagsFilter"
	DashFilter    FilterType = "DashFilter"
	FolderFilter  FilterType = "FolderFilter"
	DefaultFilter FilterType = "default"
	Name          FilterType = "Name"
	AuthLabel     FilterType = "AuthLabel"
	OrgFilter     FilterType = "OrgFilter"
)

func (s FilterType) String() string {
	return string(s)
}

type Filter interface {
	//Regex Tooling
	AddRegex(FilterType, *regexp.Regexp)
	//Entity filterMap
	GetEntity(FilterType) []string   // Returns slice of filter values or default value from Config
	GetFilter(key FilterType) string //Get the Filter value
	//FilterValid(key FilterType, value string) bool //true if filter match
	AddFilter(key FilterType, value string) //Add a filter to match against for a given type

	ValidateAll(interface{}) bool //ValidateAll if Entry is valid
	InvokeValidation(FilterType, interface{}) bool
	AddValidation(FilterType, func(interface{}) bool)
}

// BaseFilter is designed to be fairly generic, there shouldn't be any reason to extend it, but if you have a specialized
// use case feel free to do so.
type BaseFilter struct {
	filterMap          map[FilterType]string                 // Matches given field against a given value
	validationMethods  map[FilterType]func(interface{}) bool // Invokes a function to validate a certain entity type
	validationPatterns map[FilterType]*regexp.Regexp
}

func NewBaseFilter() *BaseFilter {
	b := &BaseFilter{}
	b.Init()
	return b
}

// Returns the entity filter
func (s *BaseFilter) getRegex(name FilterType) *regexp.Regexp {
	return s.validationPatterns[name]
}

func (s *BaseFilter) AddRegex(name FilterType, pattern *regexp.Regexp) {
	if name == "" {
		name = DefaultFilter
	}
	if pattern == nil {
		slog.Warn("invalid pattern received, cannot set filter pattern for entity name", "entityName", name)
		return
	}
	s.validationPatterns[name] = pattern
}

func (s *BaseFilter) getEntities(name FilterType, defaultVal []string) []string {
	if s.GetFilter(name) == "" {
		return defaultVal
	}
	entityFilter := s.GetFilter(name)
	//regex
	if s.getRegex(name) != nil {
		entityFilter = s.getRegex(name).ReplaceAllString(entityFilter, "")
	}
	s.AddFilter(name, entityFilter)

	return strings.Split(entityFilter, ",")
}

func (s *BaseFilter) GetEntity(name FilterType) []string {
	var defaultResponse []string
	if name == "" {
		return defaultResponse
	}
	switch name {
	case TagsFilter:
		entityFilter := s.GetFilter(name)
		var result []string
		err := json.Unmarshal([]byte(entityFilter), &result)
		if err == nil {
			return result
		}
		return s.getEntities(TagsFilter, []string{})
	case FolderFilter:
		return s.getEntities(FolderFilter, config.Config().GetDefaultGrafanaConfig().GetMonitoredFolders())
	default:
		return defaultResponse
	}

}

func (s *BaseFilter) AddValidation(name FilterType, f func(interface{}) bool) {
	if name == "" {
		name = DefaultFilter
	}

	s.validationMethods[name] = f

}

func (s *BaseFilter) InvokeValidation(name FilterType, i interface{}) bool {
	if name == "" {
		name = "default"
	}
	if s.validationMethods != nil && s.validationMethods[name] != nil {
		return s.validationMethods[name](i)
	}

	return false
}

// Validate Iterates through all validation checks
func (s *BaseFilter) ValidateAll(items interface{}) bool {
	for _, val := range s.validationMethods {
		ok := val(items)
		if !ok {
			return ok
		}
	}

	return true
}

// GetTypes returns all the current keys for the configured Filter
func (s *BaseFilter) GetTypes() []string {
	keys := funk.Keys(s.filterMap)
	return keys.([]string)
}

// GetFilter returns the value of the filter
func (s *BaseFilter) GetFilter(key FilterType) string {
	if val, ok := s.filterMap[key]; ok {
		return val
	}
	return ""
}

// AddFilter adds a filter and the corresponding value
func (s *BaseFilter) AddFilter(key FilterType, value string) {
	s.filterMap[key] = value
}

func (s *BaseFilter) Init() {
	s.filterMap = make(map[FilterType]string)
	s.validationMethods = make(map[FilterType]func(interface{}) bool, 0)
	s.validationPatterns = make(map[FilterType]*regexp.Regexp)
}
