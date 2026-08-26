package v2

import (
	"context"

	"github.com/esnet/gdg/internal/domain"
	"github.com/esnet/gdg/internal/ports/outbound"
)

// TypedFilterAdapter wraps any Filter and adds typed convenience methods for the most
// common runtime call patterns. T is the primary entity type that this filter validates
// (e.g. *domain.NestedHit for folders, models.DataSourceListItemDTO for connections).
//
// TypedFilterAdapter[T] satisfies outbound.Filter via embedding, so it can be passed
// anywhere Filter is accepted without a type assertion. Construct one with NewTypedFilter.
type TypedFilterAdapter[T any] struct {
	outbound.Filter
}

// NewTypedFilter wraps an existing Filter in a TypedFilterAdapter for entity type T.
// The wrapper is a value type (stack-allocated) so construction is essentially free.
func NewTypedFilter[T any](f outbound.Filter) TypedFilterAdapter[T] {
	return TypedFilterAdapter[T]{Filter: f}
}

// ValidateEntity calls ValidateAll with the entity typed as T, making the entity type
// explicit at the call site. Prefer this over a bare filter.ValidateAll call when the
// entity type is statically known.
func (a TypedFilterAdapter[T]) ValidateEntity(ctx context.Context, entity T) bool {
	return a.ValidateAll(ctx, entity)
}

// GetEntityReaderValue calls GetReaderValue with the entity typed as T. This makes the
// entity type explicit and avoids the any cast that bare GetReaderValue requires.
func (a TypedFilterAdapter[T]) GetEntityReaderValue(ctx context.Context, ft domain.FilterType, entity T) any {
	return a.GetReaderValue(ctx, ft, entity)
}
