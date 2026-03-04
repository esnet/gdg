package domain

import (
	"context"
	"fmt"
)

const (
	TagsFilter          FilterType = "TagsFilter"
	DashFilter          FilterType = "DashFilter"
	FolderFilter        FilterType = "FolderFilter"
	AlertRuleFilterType            = "AlertRuleFilter"
	Name                FilterType = "Name"
	UID                 FilterType = "UID"
	ConnectionName      FilterType = "ConnectionName" // used for Connection name
	AuthLabel           FilterType = "AuthLabel"
	OrgFilter           FilterType = "OrgFilter"
)

type (
	InputValidation func(ctx context.Context, value any, expected any) error
	Processor       func(ctx context.Context, item any) (any, error)
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

type FilterReader func(context.Context, FilterType, any) (any, error)

// FilterType Currently supported filters
type FilterType string

func (s FilterType) String() string {
	return string(s)
}
