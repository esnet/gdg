package v2

import (
	"context"
	"fmt"
	"reflect"

	"github.com/esnet/gdg/internal/domain"
	"github.com/esnet/gdg/internal/ports/outbound"
)

// RegisterTypedReader eliminates the reflect.TypeFor + type-assertion boilerplate that every
// setup<Resource>Readers function previously repeated. The caller supplies a typed handler fn
// that receives a concrete T directly; this wrapper handles the reflect registration and the
// a.(T) type dispatch, producing a cleaner and more descriptive error on mismatch.
func RegisterTypedReader[T any](f outbound.Filter, fn func(context.Context, domain.FilterType, T) (any, error)) error {
	return f.RegisterReader(reflect.TypeFor[T](), func(ctx context.Context, ft domain.FilterType, a any) (any, error) {
		val, ok := a.(T)
		if !ok {
			var zero T
			return nil, fmt.Errorf("unsupported data type: expected %T, got %T", zero, a)
		}
		return fn(ctx, ft, val)
	})
}

// RegisterTypedValidation registers a type-safe validation function for a single filter
// dimension, eliminating the GetParams[T] boilerplate from every AddValidation callback.
// Use this when the value extracted by the reader and the expected value stored in the
// filter are the same type T. The argument order follows Go's convention of configuration
// before behavior: expected value is declared before the validation function.
//
// For mismatched types (e.g. string value, []string expected) continue using AddValidation
// directly with GetMismatchParams[T, V].
func RegisterTypedValidation[T any](
	f outbound.Filter,
	filterType domain.FilterType,
	expected T,
	fn func(ctx context.Context, val T, expected T) error,
) {
	f.AddValidation(filterType, func(ctx context.Context, value, exp any) error {
		v, e, err := GetParams[T](value, exp, filterType)
		if err != nil {
			return err
		}
		return fn(ctx, v, e)
	}, expected)
}

// ValidateEntity is a typed convenience wrapper for filter.ValidateAll that makes the
// entity type explicit at the call site. Prefer this over a bare ValidateAll call when
// the entity type is statically known.
func ValidateEntity[T any](ctx context.Context, f outbound.Filter, entity T) bool {
	return f.ValidateAll(ctx, entity)
}

func GetMismatchParams[T any, V any](value, expected any, filterType domain.FilterType) (T, V, error) {
	var (
		zero1 T
		zero2 V
	)
	val, ok := value.(T)
	if !ok {
		return zero1, zero2, fmt.Errorf("invalid input data type for filter %s: expected %T, got %T", filterType, zero1, value)
	}
	// Check folder
	exp, ok := expected.(V)
	if !ok {
		return zero1, zero2, fmt.Errorf("invalid expected data type for filter %s: expected %T, got %T", filterType, zero2, expected)
	}

	return val, exp, nil
}

func GetParams[T any](value, expected any, filterType domain.FilterType) (T, T, error) {
	var zero T
	val, ok := value.(T)
	if !ok {
		return zero, zero, fmt.Errorf("invalid input data type for filter %s: expected %T, got %T", filterType, zero, value)
	}
	// Check folder
	exp, ok := expected.(T)
	if !ok {
		return zero, zero, fmt.Errorf("invalid expected data type for filter %s: expected %T, got %T", filterType, zero, expected)
	}
	return val, exp, nil
}
