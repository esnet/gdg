package v2

import (
	"fmt"

	"github.com/esnet/gdg/internal/domain"
)

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
