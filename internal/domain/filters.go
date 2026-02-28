package domain

import (
	"fmt"
)

const (
	TagsFilter          FilterType = "TagsFilter"
	DashFilter          FilterType = "DashFilter"
	FolderFilter        FilterType = "FolderFilter"
	AlertRuleFilterType            = "AlertRuleFilter"
	Name                FilterType = "Name"
	ConnectionName      FilterType = "ConnectionName" // used for Connection name
	AuthLabel           FilterType = "AuthLabel"
	OrgFilter           FilterType = "OrgFilter"
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

// FilterType Currently supported filters
type FilterType string

func (s FilterType) String() string {
	return string(s)
}
