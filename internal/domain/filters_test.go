package domain

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilterType_String(t *testing.T) {
	cases := []struct {
		ft   FilterType
		want string
	}{
		{TagsFilter, "TagsFilter"},
		{DashFilter, "DashFilter"},
		{FolderFilter, "FolderFilter"},
		{Name, "Name"},
		{UID, "UID"},
		{ConnectionName, "ConnectionName"},
		{AuthLabel, "AuthLabel"},
		{OrgFilter, "OrgFilter"},
		{AlertRuleFilterType, "AlertRuleFilter"},
	}
	for _, tc := range cases {
		t.Run(tc.want, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.ft.String())
		})
	}
}

func TestProcessorEntity_Validate_NilProcessorReturnsError(t *testing.T) {
	p := ProcessorEntity{Name: "no-op", Processor: nil}
	err := p.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no valid processor defined")
}

func TestProcessorEntity_Validate_NonNilProcessorReturnsNil(t *testing.T) {
	p := ProcessorEntity{
		Name: "identity",
		Processor: func(_ context.Context, item any) (any, error) {
			return item, nil
		},
	}
	err := p.Validate()
	require.NoError(t, err)
}
