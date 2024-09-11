package filters

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleBaseGetTypes(t *testing.T) {
	b := BaseFilter{
		filterMap: map[FilterType]string{TagsFilter: "moo", Name: "Woot"},
	}
	result := b.GetTypes()
	assert.Equal(t, len(result), 2)
	assert.True(t, slices.Contains(result, string(TagsFilter)))
	assert.True(t, slices.Contains(result, string(Name)))
}
