package config_domain

import (
	"testing"

	"github.com/go-viper/mapstructure/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// ── ContactPointSettings.FiltersEnabled ──────────────────────────────────────

func TestContactPointFiltersEnabled_NilRulesReturnsFalse(t *testing.T) {
	cp := &ContactPointSettings{}
	assert.False(t, cp.FiltersEnabled())
}

func TestContactPointFiltersEnabled_EmptySliceReturnsFalse(t *testing.T) {
	cp := &ContactPointSettings{FilterRules: []MatchingRule{}}
	assert.False(t, cp.FiltersEnabled())
}

func TestContactPointFiltersEnabled_NonEmptySliceReturnsTrue(t *testing.T) {
	cp := &ContactPointSettings{
		FilterRules: []MatchingRule{{Field: "name", Regex: "discord"}},
	}
	assert.True(t, cp.FiltersEnabled())
}

// ── YAML round-trip ───────────────────────────────────────────────────────────

const alertSettingsYAML = `
contact_points:
  filters:
    - field: name
      regex: discord
    - field: orgId
      regex: "1"
      inclusive: true
`

func TestAlertSettings_YAML_RoundTrip(t *testing.T) {
	var as AlertSettings
	require.NoError(t, yaml.Unmarshal([]byte(alertSettingsYAML), &as))

	rules := as.ContactSettings.FilterRules
	require.Len(t, rules, 2)

	assert.Equal(t, "name", rules[0].Field)
	assert.Equal(t, "discord", rules[0].Regex)
	assert.False(t, rules[0].Inclusive)

	assert.Equal(t, "orgId", rules[1].Field)
	assert.Equal(t, "1", rules[1].Regex)
	assert.True(t, rules[1].Inclusive)
}

func TestAlertSettings_YAML_EmptyFilters(t *testing.T) {
	const src = `
contact_points:
  filters: []
`
	var as AlertSettings
	require.NoError(t, yaml.Unmarshal([]byte(src), &as))
	assert.Empty(t, as.ContactSettings.FilterRules)
}

func TestAlertSettings_YAML_NoContactPoints(t *testing.T) {
	const src = `{}` // alert_settings with no contact_points key at all
	var as AlertSettings
	require.NoError(t, yaml.Unmarshal([]byte(src), &as))
	assert.Nil(t, as.ContactSettings.FilterRules)
}

// ── mapstructure round-trip (viper path) ─────────────────────────────────────

func TestAlertSettings_Mapstructure_RoundTrip(t *testing.T) {
	raw := map[string]any{
		"contact_points": map[string]any{
			"filters": []any{
				map[string]any{"field": "name", "regex": "discord"},
				map[string]any{"field": "orgId", "regex": "1", "inclusive": true},
			},
		},
	}

	var as AlertSettings
	require.NoError(t, mapstructure.Decode(raw, &as))

	rules := as.ContactSettings.FilterRules
	require.Len(t, rules, 2)

	assert.Equal(t, "name", rules[0].Field)
	assert.Equal(t, "discord", rules[0].Regex)
	assert.False(t, rules[0].Inclusive)

	assert.Equal(t, "orgId", rules[1].Field)
	assert.Equal(t, "1", rules[1].Regex)
	assert.True(t, rules[1].Inclusive)
}

func TestAlertSettings_Mapstructure_EmptyFilters(t *testing.T) {
	raw := map[string]any{
		"contact_points": map[string]any{
			"filters": []any{},
		},
	}
	var as AlertSettings
	require.NoError(t, mapstructure.Decode(raw, &as))
	assert.Empty(t, as.ContactSettings.FilterRules)
}

func TestAlertSettings_Mapstructure_NoContactPoints(t *testing.T) {
	raw := map[string]any{}
	var as AlertSettings
	require.NoError(t, mapstructure.Decode(raw, &as))
	assert.Nil(t, as.ContactSettings.FilterRules)
}
