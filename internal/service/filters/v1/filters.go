package v1

import (
	"encoding/json"
	"log/slog"
	"regexp"
	"strings"

	"github.com/esnet/gdg/internal/service/filters"

	"github.com/esnet/gdg/internal/config"
	"github.com/samber/lo"
	"golang.org/x/exp/maps"
)

// BaseFilter is designed to be fairly generic, there shouldn't be any reason to extend it, but if you have a specialized
// use case feel free to do so.
type BaseFilter struct {
	filterMap          map[filters.FilterType]string         // Matches given field against a given value
	validationMethods  map[filters.FilterType]func(any) bool // Invokes a function to validate a certain entity type
	validationPatterns map[filters.FilterType]*regexp.Regexp
}

func NewBaseFilter() filters.Filter {
	b := &BaseFilter{}
	b.Init()
	return b
}

// Returns the entity filter
func (s *BaseFilter) getRegex(name filters.FilterType) *regexp.Regexp {
	return s.validationPatterns[name]
}

func (s *BaseFilter) AddRegex(name filters.FilterType, pattern *regexp.Regexp) {
	if name == "" {
		name = filters.DefaultFilter
	}
	if pattern == nil {
		slog.Warn("invalid pattern received, cannot set filter pattern for entity name", "entityName", name)
		return
	}
	s.validationPatterns[name] = pattern
}

func (s *BaseFilter) getEntities(name filters.FilterType, defaultVal []string) []string {
	if s.GetFilter(name) == "" {
		return defaultVal
	}
	entityFilter := s.GetFilter(name)
	// regex
	if s.getRegex(name) != nil {
		entityFilter = s.getRegex(name).ReplaceAllString(entityFilter, "")
	}
	s.AddFilter(name, entityFilter)

	return strings.Split(entityFilter, ",")
}

func (s *BaseFilter) GetEntity(name filters.FilterType) []string {
	var defaultResponse []string
	if name == "" {
		return defaultResponse
	}
	switch name {
	case filters.TagsFilter:
		entityFilter := s.GetFilter(name)
		var result []string
		err := json.Unmarshal([]byte(entityFilter), &result)
		if err == nil {
			return result
		}
		return s.getEntities(filters.TagsFilter, []string{})
	case filters.FolderFilter:
		return s.getEntities(filters.FolderFilter, config.Config().GetDefaultGrafanaConfig().GetMonitoredFolders())
	default:
		return defaultResponse
	}
}

func (s *BaseFilter) AddValidation(name filters.FilterType, f func(any) bool) {
	if name == "" {
		name = filters.DefaultFilter
	}

	s.validationMethods[name] = f
}

func (s *BaseFilter) InvokeValidation(name filters.FilterType, i any) bool {
	if name == "" {
		name = "default"
	}
	if s.validationMethods != nil && s.validationMethods[name] != nil {
		return s.validationMethods[name](i)
	}

	return false
}

// ValidateAll Iterates through all validation checks
func (s *BaseFilter) ValidateAll(items any) bool {
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
	keys := maps.Keys(s.filterMap)
	stringKeys := lo.Map(keys, func(item filters.FilterType, index int) string {
		return string(item)
	})
	return stringKeys
}

// GetFilter returns the value of the filter
func (s *BaseFilter) GetFilter(key filters.FilterType) string {
	if val, ok := s.filterMap[key]; ok {
		return val
	}
	return ""
}

// AddFilter adds a filter and the corresponding value
func (s *BaseFilter) AddFilter(key filters.FilterType, value string) {
	s.filterMap[key] = value
}

func (s *BaseFilter) Init() {
	s.filterMap = make(map[filters.FilterType]string)
	s.validationMethods = make(map[filters.FilterType]func(any) bool)
	s.validationPatterns = make(map[filters.FilterType]*regexp.Regexp)
}
