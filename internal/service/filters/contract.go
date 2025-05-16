package filters

import (
	"fmt"
	"reflect"
)

const (
	TagsFilter     FilterType = "TagsFilter"
	DashFilter     FilterType = "DashFilter"
	FolderFilter   FilterType = "FolderFilter"
	Name           FilterType = "Name"
	ConnectionName FilterType = "ConnectionName" // used for Connection name
	AuthLabel      FilterType = "AuthLabel"
	OrgFilter      FilterType = "OrgFilter"
)

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
	GetExpectedValue(filterType FilterType) any
	GetExpectedString(filterType FilterType) string
	GetExpectedStringSlice(filterType FilterType) ([]string, error)
}

// FilterType Currently supported filters
type FilterType string

func (s FilterType) String() string {
	return string(s)
}
