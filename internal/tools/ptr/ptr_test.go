package ptr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPtrTo(t *testing.T) {
	v := Of(false)
	assert.False(t, *v)
	v = Of(true)
	assert.True(t, *v)
}

func TestValueOrDefault(t *testing.T) {
	// Test with integers
	t.Run("Integer with nil value", func(t *testing.T) {
		var nilInt *int = nil
		fallback := 42
		result := ValueOrDefault(nilInt, fallback)

		if result != fallback {
			t.Errorf("Expected %d, got %d", fallback, result)
		}
	})

	t.Run("Integer with non-nil value", func(t *testing.T) {
		value := 99
		fallback := 42
		result := ValueOrDefault(&value, fallback)

		if result != value {
			t.Errorf("Expected %d, got %d", value, result)
		}
	})

	// Test with strings
	t.Run("String with nil value", func(t *testing.T) {
		var nilString *string = nil
		fallback := "default"
		result := ValueOrDefault(nilString, fallback)

		if result != fallback {
			t.Errorf("Expected %s, got %s", fallback, result)
		}
	})

	t.Run("String with non-nil value", func(t *testing.T) {
		value := "actual"
		fallback := "default"
		result := ValueOrDefault(&value, fallback)

		if result != value {
			t.Errorf("Expected %s, got %s", value, result)
		}
	})

	// Test with floats
	t.Run("Float with nil value", func(t *testing.T) {
		var nilFloat *float64 = nil
		fallback := 3.14
		result := ValueOrDefault(nilFloat, fallback)

		if result != fallback {
			t.Errorf("Expected %f, got %f", fallback, result)
		}
	})

	t.Run("Float with non-nil value", func(t *testing.T) {
		value := 2.71
		fallback := 3.14
		result := ValueOrDefault(&value, fallback)

		if result != value {
			t.Errorf("Expected %f, got %f", value, result)
		}
	})

	// Test with booleans
	t.Run("Boolean with nil value", func(t *testing.T) {
		var nilBool *bool = nil
		fallback := true
		result := ValueOrDefault(nilBool, fallback)

		if result != fallback {
			t.Errorf("Expected %t, got %t", fallback, result)
		}
	})

	t.Run("Boolean with non-nil value", func(t *testing.T) {
		value := false
		fallback := true
		result := ValueOrDefault(&value, fallback)

		if result != value {
			t.Errorf("Expected %t, got %t", value, result)
		}
	})

	// Test with slices
	t.Run("Slice with nil value", func(t *testing.T) {
		var nilSlice *[]int = nil
		fallback := []int{1, 2, 3}
		result := ValueOrDefault(nilSlice, fallback)

		if len(result) != len(fallback) {
			t.Errorf("Expected slice length %d, got %d", len(fallback), len(result))
		}
		for i := range fallback {
			if result[i] != fallback[i] {
				t.Errorf("At index %d: expected %d, got %d", i, fallback[i], result[i])
			}
		}
	})

	t.Run("Slice with non-nil value", func(t *testing.T) {
		value := []int{4, 5, 6}
		fallback := []int{1, 2, 3}
		result := ValueOrDefault(&value, fallback)

		if len(result) != len(value) {
			t.Errorf("Expected slice length %d, got %d", len(value), len(result))
		}
		for i := range value {
			if result[i] != value[i] {
				t.Errorf("At index %d: expected %d, got %d", i, value[i], result[i])
			}
		}
	})

	// Test with maps
	t.Run("Map with nil value", func(t *testing.T) {
		var nilMap *map[string]int = nil
		fallback := map[string]int{"one": 1, "two": 2}
		result := ValueOrDefault(nilMap, fallback)

		if len(result) != len(fallback) {
			t.Errorf("Expected map length %d, got %d", len(fallback), len(result))
		}
		for k, v := range fallback {
			if result[k] != v {
				t.Errorf("For key %s: expected %d, got %d", k, v, result[k])
			}
		}
	})

	t.Run("Map with non-nil value", func(t *testing.T) {
		value := map[string]int{"three": 3, "four": 4}
		fallback := map[string]int{"one": 1, "two": 2}
		result := ValueOrDefault(&value, fallback)

		if len(result) != len(value) {
			t.Errorf("Expected map length %d, got %d", len(value), len(result))
		}
		for k, v := range value {
			if result[k] != v {
				t.Errorf("For key %s: expected %d, got %d", k, v, result[k])
			}
		}
	})
}
