package v2

import (
	"testing"

	"github.com/esnet/gdg/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- GetParams ---

func TestGetParams_Success(t *testing.T) {
	val, exp, err := GetParams[string]("hello", "world", domain.FolderFilter)
	require.NoError(t, err)
	assert.Equal(t, "hello", val)
	assert.Equal(t, "world", exp)
}

func TestGetParams_InvalidInputType(t *testing.T) {
	val, exp, err := GetParams[string](123, "world", domain.FolderFilter)
	require.Error(t, err)
	assert.Empty(t, val)
	assert.Empty(t, exp)
	assert.Contains(t, err.Error(), "invalid input data type")
}

func TestGetParams_InvalidExpectedType(t *testing.T) {
	val, exp, err := GetParams[string]("hello", 456, domain.FolderFilter)
	require.Error(t, err)
	assert.Empty(t, val)
	assert.Empty(t, exp)
	assert.Contains(t, err.Error(), "invalid expected data type")
}

func TestGetParams_BothInvalidTypes(t *testing.T) {
	val, exp, err := GetParams[string](123, 456, domain.FolderFilter)
	require.Error(t, err)
	assert.Empty(t, val)
	assert.Empty(t, exp)
}

func TestGetParams_WithInt(t *testing.T) {
	val, exp, err := GetParams[int](42, 99, domain.FolderFilter)
	require.NoError(t, err)
	assert.Equal(t, 42, val)
	assert.Equal(t, 99, exp)
}

func TestGetParams_WithStruct(t *testing.T) {
	type MyStruct struct{ Name string }
	a := MyStruct{Name: "a"}
	b := MyStruct{Name: "b"}
	val, exp, err := GetParams[MyStruct](a, b, domain.FolderFilter)
	require.NoError(t, err)
	assert.Equal(t, a, val)
	assert.Equal(t, b, exp)
}

// --- GetMismatchParams ---

func TestGetMismatchParams_Success(t *testing.T) {
	val, exp, err := GetMismatchParams[string, int]("hello", 42, domain.FolderFilter)
	require.NoError(t, err)
	assert.Equal(t, "hello", val)
	assert.Equal(t, 42, exp)
}

func TestGetMismatchParams_InvalidInputType(t *testing.T) {
	val, exp, err := GetMismatchParams[string, int](123, 42, domain.FolderFilter)
	require.Error(t, err)
	assert.Empty(t, val)
	assert.Zero(t, exp)
	assert.Contains(t, err.Error(), "invalid input data type for filter FolderFilter: expected string, got int")
}

func TestGetMismatchParams_InvalidExpectedType(t *testing.T) {
	val, exp, err := GetMismatchParams[string, int]("hello", "not-an-int", domain.FolderFilter)
	require.Error(t, err)
	assert.Empty(t, val) // original returns partial val on error
	assert.Zero(t, exp)
	assert.Contains(t, err.Error(), "invalid expected data type for filter FolderFilter: expected int, got string")
}

func TestGetMismatchParams_BothInvalid_CollectsAllErrors(t *testing.T) {
	// Unlike GetParamsDifferentTypes, GetMismatchParams collects ALL errors
	// so both type failures should appear in the error message
	_, _, err := GetMismatchParams[string, int](true, true, domain.FolderFilter)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid input data type")
	assert.Contains(t, err.Error(), "invalid input data type for filter FolderFilter: expected string, got bool")
}

func TestGetMismatchParams_FilterTypeInError(t *testing.T) {
	_, _, err := GetMismatchParams[string, int](123, 42, domain.AlertRuleFilterType)
	require.Error(t, err)
	assert.Contains(t, err.Error(), string(domain.AlertRuleFilterType))
}
