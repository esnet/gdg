package v2

import (
	"fmt"
	"reflect"

	"github.com/esnet/gdg/internal/service/filters"
)

func GetMismatchParams[T any, V any](value, expected any, filterType filters.FilterType) (T, V, error) {
	var allErrors []error
	val, ok := value.(T)
	if !ok {
		allErrors = append(allErrors, fmt.Errorf("invalid input data type. Cannot filter %s for type: %v", filterType, reflect.TypeOf(value)))
	}
	// Check folder
	exp, ok := expected.(V)
	if !ok {
		allErrors = append(allErrors, fmt.Errorf("invalid expected data type. Cannot filter %s for type: %v", filterType, reflect.TypeOf(expected)))
	}
	if len(allErrors) > 0 {
		return val, exp, fmt.Errorf("GetParams errors: %v", allErrors)
	}
	return val, exp, nil
}

func GetParams[T any](value, expected any, filterType filters.FilterType) (T, T, error) {
	var allErrors []error
	val, ok := value.(T)
	if !ok {
		allErrors = append(allErrors, fmt.Errorf("invalid input data type. Cannot filter %s for type: %v", filterType, reflect.TypeOf(value)))
	}
	// Check folder
	exp, ok := expected.(T)
	if !ok {
		allErrors = append(allErrors, fmt.Errorf("invalid expected data type. Cannot filter %s for type: %v", filterType, reflect.TypeOf(expected)))
	}
	if len(allErrors) > 0 {
		return val, exp, fmt.Errorf("GetParams errors: %v", allErrors)
	}
	return val, exp, nil
}
