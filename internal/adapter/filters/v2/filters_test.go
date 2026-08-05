package v2

import (
	"context"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/esnet/gdg/internal/domain"
	"github.com/esnet/gdg/internal/ports/outbound"
	"github.com/gosimple/slug"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func setupReaders(t *testing.T, v outbound.Filter) {
	err := v.RegisterReader(reflect.TypeFor[*domain.NestedHit](), func(ctx context.Context, filterType domain.FilterType, a any) (any, error) {
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

	err = v.RegisterReader(reflect.TypeFor[[]byte](), func(ctx context.Context, filterType domain.FilterType, a any) (any, error) {
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
	var v outbound.Filter = NewBaseFilter()
	setupReaders(t, v)

	v.AddValidation(domain.TagsFilter, func(ctx context.Context, item any, expected any) error {
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
		Processor: func(ctx context.Context, item any) (any, error) {
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

	ctx := context.Background()
	assert.True(t, v.Validate(ctx, domain.TagsFilter, obj))
	assert.True(t, v.ValidateAll(ctx, obj))

	strVal := v.GetExpectedString(ctx, domain.TagsFilter)
	assert.Equal(t, "[netsage Ho]", strVal)
	// no data
	strVal = v.GetExpectedString(ctx, domain.DashFilter)
	assert.Equal(t, "", strVal)
	//
	assert.Nil(t, v.GetExpectedValue(ctx, domain.DashFilter))
	anyVal := v.GetExpectedValue(ctx, domain.TagsFilter)
	anyArr, ok := anyVal.([]string)
	assert.True(t, ok)
	assert.Equal(t, []string{"netsage", "Ho"}, anyArr)
}

// ---------------------------------------------------------------------------
// GetExpectedStringSlice
// ---------------------------------------------------------------------------

func TestGetExpectedStringSlice_NoExpectedValue_ReturnsError(t *testing.T) {
	v := NewBaseFilter()
	ctx := context.Background()
	// Nothing registered for DashFilter → GetExpectedValue returns nil → error.
	result, err := v.GetExpectedStringSlice(ctx, domain.DashFilter)
	assert.Nil(t, result)
	assert.Error(t, err)
}

func TestGetExpectedStringSlice_StringSliceValue_ReturnsSlice(t *testing.T) {
	v := NewBaseFilter()
	ctx := context.Background()
	v.AddValidation(domain.TagsFilter, func(ctx context.Context, value any, expected any) error {
		return nil
	}, []string{"alpha", "beta"})

	result, err := v.GetExpectedStringSlice(ctx, domain.TagsFilter)
	assert.NoError(t, err)
	assert.Equal(t, []string{"alpha", "beta"}, result)
}

func TestGetExpectedStringSlice_SingleStringValue_WrapsInSlice(t *testing.T) {
	v := NewBaseFilter()
	ctx := context.Background()
	v.AddValidation(domain.DashFilter, func(ctx context.Context, value any, expected any) error {
		return nil
	}, "single-value")

	result, err := v.GetExpectedStringSlice(ctx, domain.DashFilter)
	assert.NoError(t, err)
	assert.Equal(t, []string{"single-value"}, result)
}

func TestGetExpectedStringSlice_UnsupportedType_ReturnsError(t *testing.T) {
	v := NewBaseFilter()
	ctx := context.Background()
	// Register an int — neither []string nor string.
	v.AddValidation(domain.FolderFilter, func(ctx context.Context, value any, expected any) error {
		return nil
	}, 42)

	result, err := v.GetExpectedStringSlice(ctx, domain.FolderFilter)
	assert.Nil(t, result)
	assert.Error(t, err)
}
