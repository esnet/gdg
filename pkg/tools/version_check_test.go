package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type moo struct {
	s string
}

func (v moo) GetServerInfo() map[string]any {
	return map[string]any{"Version": v.s}
}

func TestGrafanaRange(t *testing.T) {
	m := moo{s: "10.5.4"}
	assert.False(t, InRange([]VersionRange{{MinVersion: "v10.2.1", MaxVersion: "v10.2.2"}}, m))
	assert.True(t, InRange([]VersionRange{{MinVersion: "v10.5.1", MaxVersion: "v10.6.0"}}, m))
	// Inclusive tests
	assert.True(t, InRange([]VersionRange{{MinVersion: "v10.5.4", MaxVersion: "v10.6.0"}}, m))
	assert.True(t, InRange([]VersionRange{{MinVersion: "v10.5.1", MaxVersion: "v10.5.4"}}, m))
	assert.False(t, InRange([]VersionRange{{MinVersion: "v10.2.1", MaxVersion: "v10.2.2"}}, m))
	m.s = "10.2.0"
	assert.False(t, InRange([]VersionRange{{MinVersion: "v10.2.1", MaxVersion: "v10.2.2"}}, m))
	m.s = "10.2.1"
	assert.True(t, InRange([]VersionRange{
		{MinVersion: "v10.2.1", MaxVersion: "v10.2.2"},
		{MinVersion: "v10.1.0", MaxVersion: "v10.5.2"},
	}, m))
}

func TestGrafanaMinVersion(t *testing.T) {
	m := moo{s: "10.5.4"}
	assert.True(t, ValidateMinimumVersion("v10.3.2", m))
	assert.False(t, ValidateMinimumVersion("v10.7.2", m))
}

// ---------------------------------------------------------------------------
// Additional tests covering the rewritten semver library behaviour
// ---------------------------------------------------------------------------

// TestVersionRangeValidate exercises all branches of VersionRange.Validate.
func TestVersionRangeValidate(t *testing.T) {
	tests := []struct {
		name  string
		vr    VersionRange
		valid bool
	}{
		{"both empty is valid", VersionRange{}, true},
		{"valid min only", VersionRange{MinVersion: "v10.0.0"}, true},
		{"valid max only", VersionRange{MaxVersion: "11.2.3"}, true},
		{"valid min and max with v prefix", VersionRange{MinVersion: "v10.0.0", MaxVersion: "v11.0.0"}, true},
		{"valid min and max bare", VersionRange{MinVersion: "10.0.0", MaxVersion: "11.0.0"}, true},
		{"invalid min string", VersionRange{MinVersion: "not-a-version"}, false},
		{"invalid max string", VersionRange{MaxVersion: "not-a-version"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.valid, tt.vr.Validate())
		})
	}
}

// TestInRange_BareVersionNoPrefix verifies that InRange works when neither the
// server version nor the range strings carry a "v" prefix (Masterminds/semver
// accepts both forms).
func TestInRange_BareVersionNoPrefix(t *testing.T) {
	m := moo{s: "12.0.0"}
	assert.True(t, InRange([]VersionRange{{MinVersion: "11.0.0", MaxVersion: "13.0.0"}}, m))
	assert.False(t, InRange([]VersionRange{{MinVersion: "13.0.0", MaxVersion: "14.0.0"}}, m))
}

// TestInRange_MixedPrefixes tests that mixing v-prefixed and bare strings works.
func TestInRange_MixedPrefixes(t *testing.T) {
	m := moo{s: "12.0.0"}
	assert.True(t, InRange([]VersionRange{{MinVersion: "v11.0.0", MaxVersion: "13.0.0"}}, m))
}

// TestInRange_EmptyRangesAlwaysTrue confirms that an empty slice always returns true.
func TestInRange_EmptyRangesAlwaysTrue(t *testing.T) {
	m := moo{s: "9.0.0"}
	assert.True(t, InRange([]VersionRange{}, m))
}

// TestInRange_OnlyMinVersion confirms behaviour when only MinVersion is set.
func TestInRange_OnlyMinVersion(t *testing.T) {
	m := moo{s: "12.0.0"}
	assert.True(t, InRange([]VersionRange{{MinVersion: "v10.0.0"}}, m))
	assert.False(t, InRange([]VersionRange{{MinVersion: "v13.0.0"}}, m))
}

// TestInRange_OnlyMaxVersion confirms behaviour when only MaxVersion is set.
func TestInRange_OnlyMaxVersion(t *testing.T) {
	m := moo{s: "12.0.0"}
	assert.True(t, InRange([]VersionRange{{MaxVersion: "v13.0.0"}}, m))
	assert.False(t, InRange([]VersionRange{{MaxVersion: "v11.0.0"}}, m))
}

// TestInRange_InvalidRange confirms that an invalid range string causes InRange to return false.
func TestInRange_InvalidRange(t *testing.T) {
	m := moo{s: "12.0.0"}
	assert.False(t, InRange([]VersionRange{{MinVersion: "not-valid", MaxVersion: "also-bad"}}, m))
}

// TestInRange_MultipleRangesFirstFails confirms false is returned if the first
// range fails even if others would pass.
func TestInRange_MultipleRangesFirstFails(t *testing.T) {
	m := moo{s: "10.2.1"}
	// First range passes, second range is outside (v11+) — all must pass.
	assert.False(t, InRange([]VersionRange{
		{MinVersion: "v10.0.0", MaxVersion: "v10.5.0"},
		{MinVersion: "v11.0.0", MaxVersion: "v12.0.0"},
	}, m))
}

// TestValidateMinimumVersion_BareNoPrefix verifies that the server version
// returned without a "v" prefix is still correctly compared.
func TestValidateMinimumVersion_BareNoPrefix(t *testing.T) {
	m := moo{s: "13.0.0"}
	assert.True(t, ValidateMinimumVersion("v13.0.0", m))
	assert.True(t, ValidateMinimumVersion("13.0.0", m))
	assert.False(t, ValidateMinimumVersion("v14.0.0", m))
}

// TestValidateMinimumVersion_ExactEquality verifies that equal versions pass.
func TestValidateMinimumVersion_ExactEquality(t *testing.T) {
	m := moo{s: "10.5.4"}
	assert.True(t, ValidateMinimumVersion("v10.5.4", m))
}

// TestValidateMinimumVersion_InvalidMinVersion verifies that an unparseable
// minVersion argument returns false.
func TestValidateMinimumVersion_InvalidMinVersion(t *testing.T) {
	m := moo{s: "10.5.4"}
	assert.False(t, ValidateMinimumVersion("not-a-version", m))
}
