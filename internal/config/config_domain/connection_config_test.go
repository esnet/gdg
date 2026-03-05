package config_domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ── ConnectionSettings.FiltersEnabled ────────────────────────────────────────

func TestFiltersEnabled_NilRulesReturnsFalse(t *testing.T) {
	cs := &ConnectionSettings{}
	assert.False(t, cs.FiltersEnabled())
}

func TestFiltersEnabled_EmptySliceReturnsTrue(t *testing.T) {
	// FilterRules is initialised (non-nil) even if empty
	cs := &ConnectionSettings{FilterRules: []MatchingRule{}}
	assert.True(t, cs.FiltersEnabled())
}

func TestFiltersEnabled_NonEmptySliceReturnsTrue(t *testing.T) {
	cs := &ConnectionSettings{
		FilterRules: []MatchingRule{{Field: "name", Regex: "prod-.*"}},
	}
	assert.True(t, cs.FiltersEnabled())
}

// ── ConnectionSettings.IsExcluded ────────────────────────────────────────────

// helper: a simple JSON-serialisable struct representing a connection entity
type testConn struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func TestIsExcluded_NoFiltersNeverExcludes(t *testing.T) {
	cs := &ConnectionSettings{}
	excluded := cs.IsExcluded(testConn{Name: "prod-db", Type: "postgres"})
	assert.False(t, excluded)
}

func TestIsExcluded_ExclusiveMatchExcludes(t *testing.T) {
	cs := &ConnectionSettings{
		FilterRules: []MatchingRule{
			{Field: "name", Regex: "dev-.*", Inclusive: false},
		},
	}
	// "dev-db" matches the exclusive rule → should be excluded
	assert.True(t, cs.IsExcluded(testConn{Name: "dev-db", Type: "mysql"}))
}

func TestIsExcluded_ExclusiveNoMatchDoesNotExclude(t *testing.T) {
	cs := &ConnectionSettings{
		FilterRules: []MatchingRule{
			{Field: "name", Regex: "dev-.*", Inclusive: false},
		},
	}
	// "prod-db" does NOT match → should NOT be excluded
	assert.False(t, cs.IsExcluded(testConn{Name: "prod-db", Type: "mysql"}))
}

func TestIsExcluded_InclusiveMatchDoesNotExclude(t *testing.T) {
	cs := &ConnectionSettings{
		FilterRules: []MatchingRule{
			{Field: "name", Regex: "prod-.*", Inclusive: true},
		},
	}
	// "prod-db" matches the inclusive rule → match is flipped → NOT excluded
	assert.False(t, cs.IsExcluded(testConn{Name: "prod-db", Type: "mysql"}))
}

func TestIsExcluded_InclusiveNoMatchExcludes(t *testing.T) {
	cs := &ConnectionSettings{
		FilterRules: []MatchingRule{
			{Field: "name", Regex: "prod-.*", Inclusive: true},
		},
	}
	// "dev-db" does NOT match the inclusive rule → flipped → IS excluded
	assert.True(t, cs.IsExcluded(testConn{Name: "dev-db", Type: "mysql"}))
}

func TestIsExcluded_InvalidRegexExcludes(t *testing.T) {
	cs := &ConnectionSettings{
		FilterRules: []MatchingRule{
			{Field: "name", Regex: "[bad", Inclusive: false},
		},
	}
	// Invalid regex → IsExcluded returns true (safe default)
	assert.True(t, cs.IsExcluded(testConn{Name: "anything", Type: "mysql"}))
}

func TestIsExcluded_MissingFieldIsSkipped(t *testing.T) {
	cs := &ConnectionSettings{
		FilterRules: []MatchingRule{
			{Field: "nonexistent_field", Regex: ".*", Inclusive: false},
		},
	}
	// Field doesn't exist in the JSON → rule is skipped → not excluded
	assert.False(t, cs.IsExcluded(testConn{Name: "anything", Type: "mysql"}))
}

func TestIsExcluded_EmptyRegexSkipsRule(t *testing.T) {
	cs := &ConnectionSettings{
		FilterRules: []MatchingRule{
			{Field: "name", Regex: "", Inclusive: false},
		},
	}
	// Empty regex → rule is skipped
	assert.False(t, cs.IsExcluded(testConn{Name: "anything"}))
}

func TestIsExcluded_UnmarshalFailureExcludes(t *testing.T) {
	cs := &ConnectionSettings{
		FilterRules: []MatchingRule{{Field: "name", Regex: ".*"}},
	}
	// A channel cannot be JSON-marshalled → should return true
	assert.True(t, cs.IsExcluded(make(chan int)))
}

// ── GDGAppConfiguration helpers ───────────────────────────────────────────────

func TestGetSecureEntities_InitialisesNilMap(t *testing.T) {
	app := &GDGAppConfiguration{}
	entities := app.GetSecureEntities()
	assert.NotNil(t, entities)
	assert.Empty(t, entities)
}

func TestGetSecureEntities_ReturnsExistingMap(t *testing.T) {
	app := &GDGAppConfiguration{
		SecureConfig: map[string][]string{"key": {"val"}},
	}
	entities := app.GetSecureEntities()
	assert.Equal(t, []string{"val"}, entities["key"])
}

func TestGetAppGlobals_InitialisesNilGlobal(t *testing.T) {
	app := &GDGAppConfiguration{}
	g := app.GetAppGlobals()
	assert.NotNil(t, g)
}

func TestGetAppGlobals_ReturnsExisting(t *testing.T) {
	existing := &AppGlobals{Debug: true}
	app := &GDGAppConfiguration{Global: existing}
	g := app.GetAppGlobals()
	assert.True(t, g.Debug)
}

func TestGetContext_LowerCase(t *testing.T) {
	app := &GDGAppConfiguration{ContextName: "Staging"}
	assert.Equal(t, "staging", app.GetContext())
}

func TestGetContexts_ReturnsContextMap(t *testing.T) {
	cfg := NewGrafanaConfig()
	app := &GDGAppConfiguration{
		Contexts: map[string]*GrafanaConfig{"default": cfg},
	}
	assert.Equal(t, cfg, app.GetContexts()["default"])
}

func TestUpdateContextNames_SlugifiesKeys(t *testing.T) {
	app := &GDGAppConfiguration{
		Contexts: map[string]*GrafanaConfig{
			"My Org": NewGrafanaConfig(),
		},
	}
	app.UpdateContextNames()
	// slug.Make("My Org") = "my-org"
	assert.Equal(t, "my-org", app.Contexts["My Org"].contextName)
}
