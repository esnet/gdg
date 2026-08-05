package config_domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTemplatingConfig(templates ...TemplateDashboards) *TemplatingConfig {
	return &TemplatingConfig{
		Entities: TemplateEntities{
			Dashboards: templates,
		},
	}
}

func TestGetTemplate_Found(t *testing.T) {
	cfg := newTemplatingConfig(
		TemplateDashboards{TemplateName: "alpha"},
		TemplateDashboards{TemplateName: "beta"},
		TemplateDashboards{TemplateName: "gamma"},
	)

	got, ok := cfg.GetTemplate("beta")

	require.True(t, ok)
	require.NotNil(t, got)
	assert.Equal(t, "beta", got.TemplateName)
}

func TestGetTemplate_NotFound(t *testing.T) {
	cfg := newTemplatingConfig(
		TemplateDashboards{TemplateName: "alpha"},
		TemplateDashboards{TemplateName: "beta"},
	)

	got, ok := cfg.GetTemplate("missing")

	assert.False(t, ok)
	assert.Nil(t, got)
}

func TestGetTemplate_EmptySlice(t *testing.T) {
	cfg := newTemplatingConfig()

	got, ok := cfg.GetTemplate("anything")

	assert.False(t, ok)
	assert.Nil(t, got)
}

func TestGetTemplate_FirstEntry(t *testing.T) {
	cfg := newTemplatingConfig(
		TemplateDashboards{TemplateName: "first"},
		TemplateDashboards{TemplateName: "second"},
	)

	got, ok := cfg.GetTemplate("first")

	require.True(t, ok)
	assert.Equal(t, "first", got.TemplateName)
}

func TestGetTemplate_LastEntry(t *testing.T) {
	cfg := newTemplatingConfig(
		TemplateDashboards{TemplateName: "first"},
		TemplateDashboards{TemplateName: "second"},
		TemplateDashboards{TemplateName: "last"},
	)

	got, ok := cfg.GetTemplate("last")

	require.True(t, ok)
	assert.Equal(t, "last", got.TemplateName)
}

// TestGetTemplate_ReturnedPointerIsAddressable verifies that the returned pointer
// points into the slice (not a copy), so callers can mutate it in place.
func TestGetTemplate_ReturnedPointerIsAddressable(t *testing.T) {
	cfg := newTemplatingConfig(
		TemplateDashboards{TemplateName: "mutable"},
	)

	got, ok := cfg.GetTemplate("mutable")
	require.True(t, ok)

	// Mutate via the returned pointer.
	got.TemplateName = "changed"

	// The slice entry should reflect the change.
	assert.Equal(t, "changed", cfg.Entities.Dashboards[0].TemplateName)
}

func TestGetTemplate_WithDashboardEntities(t *testing.T) {
	cfg := newTemplatingConfig(
		TemplateDashboards{
			TemplateName: "with-entities",
			DashboardEntities: []TemplateDashboardEntity{
				{
					Folder:           "General",
					OrganizationName: "Main Org",
					DashboardName:    "My Dashboard",
					TemplateData:     map[string]any{"env": "prod"},
				},
			},
		},
	)

	got, ok := cfg.GetTemplate("with-entities")

	require.True(t, ok)
	require.Len(t, got.DashboardEntities, 1)
	assert.Equal(t, "General", got.DashboardEntities[0].Folder)
	assert.Equal(t, "Main Org", got.DashboardEntities[0].OrganizationName)
	assert.Equal(t, "prod", got.DashboardEntities[0].TemplateData["env"])
}
