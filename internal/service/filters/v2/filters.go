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

func (b BaseFilter) readInputValue(f filters.FilterType, obj any) (any, error) {
	t := reflect.TypeOf(obj)
	if val, ok := b.readerMap[t]; ok {
		return val(f, obj)
	}

	return nil, fmt.Errorf("no reader registered for type %v", f)
}

func (b BaseFilter) readExpectedValue(f filters.FilterType) (any, error) {
	if val, ok := b.expectedValueLookup[f]; ok {
		return val, nil
	}

	return nil, fmt.Errorf("no expected valua available for type %v", f)
}

func (b BaseFilter) AddValidation(f filters.FilterType, validation filters.InputValidation, expected any) {
	b.validationMethods[f] = validation
	b.expectedValueLookup[f] = expected
}

func (b BaseFilter) ValidateAll(obj any) bool {
	valid := true
	for k := range b.expectedValueLookup {
		valid = b.Validate(k, obj)
		if !valid {
			break
		}
	}

	return true
}

func (b BaseFilter) Validate(f filters.FilterType, obj any) bool {
	// get Data
	val, err := b.readInputValue(f, obj)
	if err != nil {
		slog.Warn("unable to read obj value", "err", err, "filter", f)
		return false
	}
	preProcList := b.preProcessMethods[f]
	for _, fn := range preProcList {
		slog.Debug("Running pre-processing function", "name", fn.Name, slog.Any("value", val))
		val, err = fn.Processor(val)
		if err != nil {
			slog.Error("unable to run pre-processing method on input", "name", fn.Name, slog.Any("value", val))
		}
	}
	expectedVal, err := b.readExpectedValue(f)
	if err != nil {
		slog.Warn("unable to filter expected value", "err", err, "filter", f)
		return false
	}
	return b.validationMethods[f](val, expectedVal) == nil
}

func (b BaseFilter) GetExpectedValue(filterType filters.FilterType) any {
	return b.expectedValueLookup[filterType]
}

func (b BaseFilter) GetStringValue(filterType filters.FilterType) string {
	if val, ok := b.expectedValueLookup[filterType]; ok {
		return fmt.Sprintf("%v", val)
	}
	return ""
}

func NewBaseFilter() filters.V2Filter {
	return &BaseFilter{
		readerMap:           make(map[reflect.Type]filters.FilterReader),
		validationMethods:   make(map[filters.FilterType]filters.InputValidation),
		expectedValueLookup: make(map[filters.FilterType]any),
		preProcessMethods:   make(map[filters.FilterType][]filters.ProcessorEntity),
	}
}
