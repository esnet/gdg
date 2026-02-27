package ports

import (
	"reflect"

	"github.com/esnet/gdg/internal/domain"
)

type V2Filter interface {
	RegisterReader(entityType reflect.Type, fn domain.FilterReader) error
	RegisterDataProcessor(entityType domain.FilterType, entity domain.ProcessorEntity) error
	AddValidation(f domain.FilterType, validation domain.InputValidation, expected any)
	Validate(domain.FilterType, any) bool
	ValidateAll(any) bool // ValidateAll if Entry is valid
	GetExpectedValue(filterType domain.FilterType) any
	GetExpectedString(filterType domain.FilterType) string
	GetExpectedStringSlice(filterType domain.FilterType) ([]string, error)
}
