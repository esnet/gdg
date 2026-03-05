package config_domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultUserPassword_AdminReturnsEmpty(t *testing.T) {
	u := &UserSettings{}
	assert.Equal(t, "", u.defaultUserPassword("admin"))
}

func TestDefaultUserPassword_ConsistentForSameUsername(t *testing.T) {
	u := &UserSettings{}
	pw1 := u.defaultUserPassword("alice")
	pw2 := u.defaultUserPassword("alice")
	assert.Equal(t, pw1, pw2, "same username must always produce the same password")
}

func TestDefaultUserPassword_DifferentForDifferentUsers(t *testing.T) {
	u := &UserSettings{}
	pw1 := u.defaultUserPassword("alice")
	pw2 := u.defaultUserPassword("bob")
	assert.NotEqual(t, pw1, pw2)
}

func TestDefaultUserPassword_NonEmpty(t *testing.T) {
	u := &UserSettings{}
	pw := u.defaultUserPassword("testuser")
	assert.NotEmpty(t, pw)
}

func TestGetPassword_NoRandomReturnsDeterministic(t *testing.T) {
	u := &UserSettings{RandomPassword: false}
	pw := u.GetPassword("alice")
	assert.Equal(t, u.defaultUserPassword("alice"), pw)
}

func TestGetPassword_RandomPasswordAtLeastMinLength(t *testing.T) {
	u := &UserSettings{
		RandomPassword: true,
		MinLength:      8,
		MaxLength:      16,
	}
	// The algorithm picks rand in [0, MaxLength) and adds MinLength,
	// so length is in [MinLength, MinLength+MaxLength-1].
	for range 10 {
		pw := u.GetPassword("alice")
		require.NotEmpty(t, pw)
		assert.GreaterOrEqual(t, len(pw), 8, "password must be at least MinLength characters")
	}
}

func TestGetPassword_MinGreaterThanMaxFallsBack(t *testing.T) {
	u := &UserSettings{
		RandomPassword: true,
		MinLength:      20,
		MaxLength:      5,
	}
	// Falls back to deterministic password
	pw := u.GetPassword("bob")
	assert.Equal(t, u.defaultUserPassword("bob"), pw)
}
