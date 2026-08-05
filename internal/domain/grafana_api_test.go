package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// Permission version constants
// ---------------------------------------------------------------------------

func TestGdgPermApiVersionConstants(t *testing.T) {
	assert.Equal(t, "v1", GdgPermApiVersionV1)
	assert.Equal(t, "dashboard-permissions.grafana/v2", GdgPermApiVersionV2)
	// Ensure the two values are distinct.
	assert.NotEqual(t, GdgPermApiVersionV1, GdgPermApiVersionV2)
}

// ---------------------------------------------------------------------------
// DashboardPermissionsGdg struct initialisation
// ---------------------------------------------------------------------------

func TestDashboardPermissionsGdg_ZeroValue(t *testing.T) {
	var gdg DashboardPermissionsGdg
	assert.Empty(t, gdg.DashboardUID)
	assert.Empty(t, gdg.DashboardName)
	assert.Empty(t, gdg.GdgApiVersion)
	assert.Nil(t, gdg.Permissions)
}

func TestDashboardPermissionsGdg_FieldAssignment(t *testing.T) {
	gdg := DashboardPermissionsGdg{
		DashboardUID:  "uid-abc",
		DashboardName: "My Dashboard",
		GdgApiVersion: GdgPermApiVersionV2,
	}
	assert.Equal(t, "uid-abc", gdg.DashboardUID)
	assert.Equal(t, "My Dashboard", gdg.DashboardName)
	assert.Equal(t, GdgPermApiVersionV2, gdg.GdgApiVersion)
}

// ---------------------------------------------------------------------------
// DashboardAndPermissions — verify the unified permission type
// ---------------------------------------------------------------------------

func TestDashboardAndPermissions_ZeroValue(t *testing.T) {
	var dp DashboardAndPermissions
	assert.Nil(t, dp.Dashboard)
	assert.Nil(t, dp.Permissions)
}

func TestDashboardAndPermissions_WithDashboard(t *testing.T) {
	hit := &NestedHit{NestedPath: "General"}
	dp := DashboardAndPermissions{
		Dashboard:   hit,
		Permissions: nil,
	}
	assert.Equal(t, "General", dp.Dashboard.NestedPath)
}
