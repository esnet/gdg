package filters

import "regexp"

type Filter interface {
	// Regex Tooling
	AddRegex(FilterType, *regexp.Regexp)
	// Entity filterMap
	GetEntity(FilterType) []string   // Returns slice of filter values or default value from Config
	GetFilter(key FilterType) string // Get the Filter value
	// FilterValid(key FilterType, value string) bool //true if filter match
	AddFilter(key FilterType, value string) // Add a filter to match against for a given type

	ValidateAll(any) bool // ValidateAll if Entry is valid
	InvokeValidation(FilterType, any) bool
	AddValidation(FilterType, func(any) bool)
}

// Adding an interface to avoid cyclic dependencies lines up to service.ServerInfoApi
type IGrafanaConfig interface {
	GetMonitoredFolders() []string
}
