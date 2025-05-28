package v1

import (
	"slices"
	"testing"

	"github.com/esnet/gdg/internal/service/filters"

	"github.com/stretchr/testify/assert"
)

func TestSimpleBaseGetTypes(t *testing.T) {
	b := BaseFilter{
		filterMap: map[filters.FilterType]string{filters.TagsFilter: "moo", filters.Name: "Woot"},
	}
	result := b.GetTypes()
	assert.Equal(t, len(result), 2)
	assert.True(t, slices.Contains(result, string(filters.TagsFilter)))
	assert.True(t, slices.Contains(result, string(filters.Name)))
}
