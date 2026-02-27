package v2

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/esnet/gdg/internal/domain"
	"github.com/esnet/gdg/internal/ports"
	"github.com/gosimple/slug"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func setupReaders(t *testing.T, v ports.V2Filter) {
	err := v.RegisterReader(reflect.TypeFor[*domain.NestedHit](), func(filterType domain.FilterType, a any) (any, error) {
		val, ok := a.(*domain.NestedHit)
		if !ok {
			return nil, fmt.Errorf("unsupported data type")
		}
		switch filterType {
		case domain.FolderFilter:
			return val.FolderTitle, nil
		case domain.TagsFilter:
			return val.Tags, nil
		case domain.DashFilter:
			return slug.Make(val.Title), nil

		default:
			return nil, fmt.Errorf("unsupported data type")
		}
	})
	assert.NoError(t, err)

	err = v.RegisterReader(reflect.TypeFor[[]byte](), func(filterType domain.FilterType, a any) (any, error) {
		val, ok := a.([]byte)
		if !ok {
			return nil, fmt.Errorf("unsupported data type")
		}
		switch filterType {
		case domain.FolderFilter:
			{
				r := gjson.GetBytes(val, "folderTitle")
				if !r.Exists() {
					return "General", nil
				}
				return r.String(), nil
			}
		case domain.TagsFilter:
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
		case domain.DashFilter:
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
	var v ports.V2Filter = NewBaseFilter()
	setupReaders(t, v)

	v.AddValidation(domain.TagsFilter, func(item any, expected any) error {
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

	obj := &domain.NestedHit{
		Hit: &models.Hit{
			Tags: []string{"Ho  "},
		},
	}

	err := v.RegisterDataProcessor(domain.TagsFilter, domain.ProcessorEntity{
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

	assert.True(t, v.Validate(domain.TagsFilter, obj))
	assert.True(t, v.ValidateAll(obj))

	strVal := v.GetExpectedString(domain.TagsFilter)
	assert.Equal(t, "[netsage Ho]", strVal)
	// no data
	strVal = v.GetExpectedString(domain.DashFilter)
	assert.Equal(t, "", strVal)
	//
	assert.Nil(t, v.GetExpectedValue(domain.DashFilter))
	anyVal := v.GetExpectedValue(domain.TagsFilter)
	anyArr, ok := anyVal.([]string)
	assert.True(t, ok)
	assert.Equal(t, []string{"netsage", "Ho"}, anyArr)
}
