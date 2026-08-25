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

func TestFiltersEnabled_EmptySliceReturnsFalse(t *testing.T) {
	// An initialised but empty slice has no active rules, so FiltersEnabled is false
	cs := &ConnectionSettings{FilterRules: []MatchingRule{}}
	assert.False(t, cs.FiltersEnabled())
}

func TestFiltersEnabled_NonEmptySliceReturnsTrue(t *testing.T) {
	cs := &ConnectionSettings{
		FilterRules: []MatchingRule{{Field: "name", Regex: "prod-.*"}},
	}
	assert.True(t, cs.FiltersEnabled())
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
