package ports

import (
	"context"
	"reflect"

	"github.com/esnet/gdg/internal/domain"
)

type Filter interface {
	// Setup should not need ctx, but may revisit
	RegisterReader(entityType reflect.Type, fn domain.FilterReader) error
	RegisterDataProcessor(entityType domain.FilterType, entity domain.ProcessorEntity) error
	AddValidation(f domain.FilterType, validation domain.InputValidation, expected any)
	// Runtime validation all have context
	Validate(context.Context, domain.FilterType, any) bool
	ValidateAll(ctx context.Context, data any) bool // ValidateAll if Entry is valid
	GetExpectedValue(ctx context.Context, filterType domain.FilterType) any
	GetExpectedString(ctx context.Context, filterType domain.FilterType) string
	GetExpectedStringSlice(ctx context.Context, filterType domain.FilterType) ([]string, error)
	GetReaderValue(ctx context.Context, filterType domain.FilterType, obj any) any
}
