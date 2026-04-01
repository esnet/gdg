package v2

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"

	"github.com/esnet/gdg/internal/domain"
	"github.com/esnet/gdg/internal/ports/outbound"
)

type BaseFilter struct {
	readerMap           map[reflect.Type]domain.FilterReader
	validationMethods   map[domain.FilterType]domain.InputValidation   // Invokes a function to validate a certain entity type
	preProcessMethods   map[domain.FilterType][]domain.ProcessorEntity // Invokes a function to validate a certain entity type
	expectedValueLookup map[domain.FilterType]any
}

func (b BaseFilter) RegisterReader(entityType reflect.Type, fn domain.FilterReader) error {
	b.readerMap[entityType] = fn
	return nil
}

func (b BaseFilter) RegisterDataProcessor(entityType domain.FilterType, entity domain.ProcessorEntity) error {
	if err := entity.Validate(); err != nil {
		return err
	}
	val, ok := b.preProcessMethods[entityType]
	if !ok {
		val = make([]domain.ProcessorEntity, 0)
	}
	val = append(val, entity)

	b.preProcessMethods[entityType] = val
	return nil
}

func (b BaseFilter) readInputValue(ctx context.Context, filterType domain.FilterType, obj any) (any, error) {
	t := reflect.TypeOf(obj)
	if val, ok := b.readerMap[t]; ok {
		return val(ctx, filterType, obj)
	}

	return nil, fmt.Errorf("no reader registered for type %v", filterType)
}

func (b BaseFilter) readExpectedValue(filterType domain.FilterType) (any, error) {
	if val, ok := b.expectedValueLookup[filterType]; ok {
		return val, nil
	}

	return nil, fmt.Errorf("no expected valua available for type %v", filterType)
}

func (b BaseFilter) AddValidation(filterType domain.FilterType, validation domain.InputValidation, expected any) {
	b.validationMethods[filterType] = validation
	b.expectedValueLookup[filterType] = expected
}

func (b BaseFilter) ValidateAll(ctx context.Context, data any) bool {
	valid := true
	for k := range b.expectedValueLookup {
		valid = b.Validate(ctx, k, data)
		if !valid {
			break
		}
	}

	return valid
}

func (b BaseFilter) applyPreProcessor(ctx context.Context, filterType domain.FilterType, val any) (any, error) {
	var allErrors []error
	var err error
	preProcList := b.preProcessMethods[filterType]
	for _, fn := range preProcList {
		slog.Debug("Running pre-processing function", "name", fn.Name, slog.Any("value", val))
		val, err = fn.Processor(ctx, val)
		if err != nil {
			allErrors = append(allErrors, fmt.Errorf("pre-processing method %s failed: %w", fn.Name, err))
			slog.Error("unable to run pre-processing method on input", "name", fn.Name, slog.Any("value", val))
		}
	}
	if len(allErrors) > 0 {
		return val, fmt.Errorf("pre-processing errors: %v", allErrors)
	}

	return val, nil
}

func (b BaseFilter) Validate(ctx context.Context, filterType domain.FilterType, obj any) bool {
	// get Data
	val, err := b.readInputValue(ctx, filterType, obj)
	if err != nil {
		slog.Warn("unable to read obj value", "err", err, "filter", filterType)
		return false
	}
	val, err = b.applyPreProcessor(ctx, filterType, val)
	if err != nil {
		slog.Error("unable to run pre-processing method on input", "err", err)
		return false
	}

	expectedVal, err := b.readExpectedValue(filterType)
	if err != nil {
		slog.Warn("unable to filter expected value", "err", err, "filter", filterType)
		return false
	}
	if validationMethod, okMethod := b.validationMethods[filterType]; okMethod {
		return validationMethod(ctx, val, expectedVal) == nil
	}
	return false
}

func (b BaseFilter) GetReaderValue(ctx context.Context, filterType domain.FilterType, obj any) any {
	val, err := b.readInputValue(ctx, filterType, obj)
	if err != nil {
		slog.Warn("unable to read obj value", "err", err, "filter", filterType)
		return nil
	}
	return val
}

func (b BaseFilter) GetExpectedValue(ctx context.Context, filterType domain.FilterType) any {
	val := b.expectedValueLookup[filterType]
	if val == nil {
		return nil
	}
	val, err := b.applyPreProcessor(ctx, filterType, val)
	if err != nil {
		return nil
	}

	return val
}

func (b BaseFilter) GetExpectedString(ctx context.Context, filterType domain.FilterType) string {
	val := b.GetExpectedValue(ctx, filterType)
	if val == nil {
		return ""
	}
	return fmt.Sprintf("%v", val)
}

func (b BaseFilter) GetExpectedStringSlice(ctx context.Context, filterType domain.FilterType) ([]string, error) {
	val := b.GetExpectedValue(ctx, filterType)
	if val == nil {
		return nil, fmt.Errorf("expected value is not found")
	}
	switch v := val.(type) {
	case []string:
		return v, nil
	case string:
		return []string{v}, nil
	default:
		return nil, fmt.Errorf("expected value is not a string  slice")
	}
}

func NewBaseFilter() outbound.Filter {
	return &BaseFilter{
		readerMap:           make(map[reflect.Type]domain.FilterReader),
		validationMethods:   make(map[domain.FilterType]domain.InputValidation),
		expectedValueLookup: make(map[domain.FilterType]any),
		preProcessMethods:   make(map[domain.FilterType][]domain.ProcessorEntity),
	}
}
