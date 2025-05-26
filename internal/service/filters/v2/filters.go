package v2

import (
	"fmt"
	"log/slog"
	"reflect"

	"github.com/esnet/gdg/internal/service/filters"
)

type BaseFilter struct {
	readerMap           map[reflect.Type]filters.FilterReader
	validationMethods   map[filters.FilterType]filters.InputValidation   // Invokes a function to validate a certain entity type
	preProcessMethods   map[filters.FilterType][]filters.ProcessorEntity // Invokes a function to validate a certain entity type
	expectedValueLookup map[filters.FilterType]any
}

func (b BaseFilter) RegisterReader(entityType reflect.Type, fn filters.FilterReader) error {
	b.readerMap[entityType] = fn
	return nil
}

func (b BaseFilter) RegisterDataProcessor(entityType filters.FilterType, entity filters.ProcessorEntity) error {
	if err := entity.Validate(); err != nil {
		return err
	}
	val, ok := b.preProcessMethods[entityType]
	if !ok {
		val = make([]filters.ProcessorEntity, 0)
	}
	val = append(val, entity)

	b.preProcessMethods[entityType] = val
	return nil
}

func (b BaseFilter) readInputValue(filterType filters.FilterType, obj any) (any, error) {
	t := reflect.TypeOf(obj)
	if val, ok := b.readerMap[t]; ok {
		return val(filterType, obj)
	}

	return nil, fmt.Errorf("no reader registered for type %v", filterType)
}

func (b BaseFilter) readExpectedValue(filterType filters.FilterType) (any, error) {
	if val, ok := b.expectedValueLookup[filterType]; ok {
		return val, nil
	}

	return nil, fmt.Errorf("no expected valua available for type %v", filterType)
}

func (b BaseFilter) AddValidation(filterType filters.FilterType, validation filters.InputValidation, expected any) {
	b.validationMethods[filterType] = validation
	b.expectedValueLookup[filterType] = expected
}

func (b BaseFilter) ValidateAll(obj any) bool {
	valid := true
	for k := range b.expectedValueLookup {
		valid = b.Validate(k, obj)
		if !valid {
			break
		}
	}

	return valid
}

func (b BaseFilter) applyPreProcessor(filterType filters.FilterType, val any) (any, error) {
	var allErrors []error
	var err error
	preProcList := b.preProcessMethods[filterType]
	for _, fn := range preProcList {
		slog.Debug("Running pre-processing function", "name", fn.Name, slog.Any("value", val))
		val, err = fn.Processor(val)
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

func (b BaseFilter) Validate(filterType filters.FilterType, obj any) bool {
	// get Data
	val, err := b.readInputValue(filterType, obj)
	if err != nil {
		slog.Warn("unable to read obj value", "err", err, "filter", filterType)
		return false
	}
	val, err = b.applyPreProcessor(filterType, val)
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
		return validationMethod(val, expectedVal) == nil
	}
	return false
}

func (b BaseFilter) GetExpectedValue(filterType filters.FilterType) any {
	val := b.expectedValueLookup[filterType]
	if val == nil {
		return nil
	}
	val, err := b.applyPreProcessor(filterType, val)
	if err != nil {
		return nil
	}

	return val
}

func (b BaseFilter) GetExpectedString(filterType filters.FilterType) string {
	val := b.GetExpectedValue(filterType)
	if val == nil {
		return ""
	}
	return fmt.Sprintf("%v", val)
}

func (b BaseFilter) GetExpectedStringSlice(filterType filters.FilterType) ([]string, error) {
	val := b.GetExpectedValue(filterType)
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

func NewBaseFilter() filters.V2Filter {
	return &BaseFilter{
		readerMap:           make(map[reflect.Type]filters.FilterReader),
		validationMethods:   make(map[filters.FilterType]filters.InputValidation),
		expectedValueLookup: make(map[filters.FilterType]any),
		preProcessMethods:   make(map[filters.FilterType][]filters.ProcessorEntity),
	}
}
