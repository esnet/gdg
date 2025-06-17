package v2

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/esnet/gdg/internal/service/filters"
	"github.com/esnet/gdg/internal/types"
	"github.com/gosimple/slug"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func setupReaders(t *testing.T, v filters.V2Filter) {
	obj := types.NestedHit{}

	err := v.RegisterReader(reflect.TypeOf(&obj), func(filterType filters.FilterType, a any) (any, error) {
		val, ok := a.(*types.NestedHit)
		if !ok {
			return nil, fmt.Errorf("unsupported data type")
		}
		switch filterType {
		case filters.FolderFilter:
			return val.FolderTitle, nil
		case filters.TagsFilter:
			return val.Tags, nil
		case filters.DashFilter:
			return slug.Make(val.Title), nil

		default:
			return nil, fmt.Errorf("unsupported data type")
		}
	})
	assert.NoError(t, err)

	err = v.RegisterReader(reflect.TypeOf([]byte{}), func(filterType filters.FilterType, a any) (any, error) {
		val, ok := a.([]byte)
		if !ok {
			return nil, fmt.Errorf("unsupported data type")
		}
		switch filterType {
		case filters.FolderFilter:
			{
				r := gjson.GetBytes(val, "folderTitle")
				if !r.Exists() {
					return "General", nil
				}
				return r.String(), nil
			}
		case filters.TagsFilter:
			{
				r := gjson.GetBytes(val, "tags")
				if !r.Exists() || !r.IsArray() {
					return nil, fmt.Errorf("no valid title found")
				}
				ar := r.Array()
				data := lo.Map(ar, func(item gjson.Result, index int) string {
					return item.String()
				})
				return data, nil

			}
			// return val.Tags, nil
		case filters.DashFilter:
			{
				r := gjson.GetBytes(val, "title")
				if !r.Exists() || r.String() == "" {
					return nil, fmt.Errorf("no valid title found")
				}
				return r.String(), nil
			}
		default:
			return nil, fmt.Errorf("unsupported data type")
		}
	})

	assert.NoError(t, err)
}

func TestFilters(t *testing.T) {
	var v filters.V2Filter = NewBaseFilter()
	setupReaders(t, v)

	v.AddValidation(filters.TagsFilter, func(item any, expected any) error {
		itemObj, itemOk := item.([]string)
		if !itemOk {
			return fmt.Errorf("item was not a slice")
		}
		expectedVal, expectedOk := expected.([]string)
		if !expectedOk {
			return fmt.Errorf("expecred value was not a slice")
		}
		for _, expectedTag := range expectedVal {
			if slices.Contains(itemObj, expectedTag) {
				return nil
			}
		}

		return fmt.Errorf("tag was not found")
	}, []string{"netsage", "Ho"})

	obj := &types.NestedHit{
		Hit: &models.Hit{
			Tags: []string{"Ho  "},
		},
	}

	err := v.RegisterDataProcessor(filters.TagsFilter, filters.ProcessorEntity{
		Name: "Space Remover",
		Processor: func(item any) (any, error) {
			val, ok := item.([]string)
			if !ok {
				return val, fmt.Errorf("invalid data format received")
			}
			for ndx, i := range val {
				val[ndx] = strings.ReplaceAll(i, " ", "")
			}

			return val, nil
		},
	})
	assert.NoError(t, err)

	assert.True(t, v.Validate(filters.TagsFilter, obj))
	assert.True(t, v.ValidateAll(obj))

	strVal := v.GetExpectedString(filters.TagsFilter)
	assert.Equal(t, "[netsage Ho]", strVal)
	// no data
	strVal = v.GetExpectedString(filters.DashFilter)
	assert.Equal(t, "", strVal)
	//
	assert.Nil(t, v.GetExpectedValue(filters.DashFilter))
	anyVal := v.GetExpectedValue(filters.TagsFilter)
	anyArr, ok := anyVal.([]string)
	assert.True(t, ok)
	assert.Equal(t, []string{"netsage", "Ho"}, anyArr)
}
