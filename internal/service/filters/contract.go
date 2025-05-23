package filters

import (
	"fmt"
	"reflect"
	"regexp"
)

const (
	TagsFilter    FilterType = "TagsFilter"
	DashFilter    FilterType = "DashFilter"
	FolderFilter  FilterType = "FolderFilter"
	DefaultFilter FilterType = "default"
	Name          FilterType = "Name"
	AuthLabel     FilterType = "AuthLabel"
	OrgFilter     FilterType = "OrgFilter"
)

func DefaultLoader(t any) any {
	return t
}

type (
	InputValidation func(value any, expected any) error
	Processor       func(item any) (any, error)
)

type ProcessorEntity struct {
	Name string
	// priority  int8
	// postProcess bool
	Processor Processor
}

func (p ProcessorEntity) Validate() error {
	if p.Processor == nil {
		return fmt.Errorf("no valid processor defined")
	}

	return nil
}

type FilterReader func(FilterType, any) (any, error)

type V2Filter interface {
	RegisterReader(entityType reflect.Type, fn FilterReader) error
	RegisterDataProcessor(entityType FilterType, entity ProcessorEntity) error
	AddValidation(f FilterType, validation InputValidation, expected any)
	Validate(FilterType, any) bool
	ValidateAll(any) bool // ValidateAll if Entry is valid
	//
	GetExpectedValue(filterType FilterType) any
	GetExpectedString(filterType FilterType) string
	GetExpectedStringSlice(filterType FilterType) ([]string, error)
}

type Filter interface {
	// Regex Tooling
	AddRegex(FilterType, *regexp.Regexp)
	// FilterValid(key FilterType, value string) bool //true if filter match
	AddFilter(key FilterType, value string) // Add a filter to match against for a given type
	AddValidation(FilterType, func(any) bool)
	GetEntity(FilterType) []string   // Returns slice of filter values or default value from Config
	GetFilter(key FilterType) string // Get the Filter value

	ValidateAll(any) bool // ValidateAll if Entry is valid
	InvokeValidation(FilterType, any) bool
}

// FilterType Currently supported filters
type FilterType string

func (s FilterType) String() string {
	return string(s)
}
