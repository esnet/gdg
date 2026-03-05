package config_domain

import (
	"testing"

	resourceTypes "github.com/esnet/gdg/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── NewGrafanaConfig / Apply / options ────────────────────────────────────────

func TestNewGrafanaConfig_Defaults(t *testing.T) {
	cfg := NewGrafanaConfig()
	require.NotNil(t, cfg)
	assert.Empty(t, cfg.URL)
	assert.Empty(t, cfg.UserName)
}

func TestNewGrafanaConfig_WithContextName(t *testing.T) {
	cfg := NewGrafanaConfig(WithContextName("prod"))
	assert.Equal(t, "prod", cfg.contextName)
}

func TestNewGrafanaConfig_WithSecureAuth(t *testing.T) {
	auth := SecureModel{Password: "secret", Token: "tok"}
	cfg := NewGrafanaConfig(WithSecureAuth(auth))
	require.NotNil(t, cfg.secureAuth)
	assert.Equal(t, "secret", cfg.secureAuth.Password)
	assert.Equal(t, "tok", cfg.secureAuth.Token)
}

func TestGrafanaConfig_Apply(t *testing.T) {
	cfg := NewGrafanaConfig()
	cfg.Apply(WithContextName("staging"))
	assert.Equal(t, "staging", cfg.contextName)
}

// ── GetURL ────────────────────────────────────────────────────────────────────

func TestGetURL_EmptyReturnEmpty(t *testing.T) {
	cfg := NewGrafanaConfig()
	assert.Equal(t, "", cfg.GetURL())
}

func TestGetURL_AppendsTrailingSlash(t *testing.T) {
	cfg := NewGrafanaConfig()
	cfg.URL = "http://localhost:3000"
	assert.Equal(t, "http://localhost:3000/", cfg.GetURL())
}

func TestGetURL_DoesNotDoubleSlash(t *testing.T) {
	cfg := NewGrafanaConfig()
	cfg.URL = "http://localhost:3000/"
	assert.Equal(t, "http://localhost:3000/", cfg.GetURL())
}

func TestGetURL_TrimsWhitespace(t *testing.T) {
	cfg := NewGrafanaConfig()
	cfg.URL = "  http://grafana.example.com  "
	result := cfg.GetURL()
	assert.Equal(t, "http://grafana.example.com/", result)
}

// ── IsBasicAuth ───────────────────────────────────────────────────────────────

func TestIsBasicAuth_BothSetReturnsTrue(t *testing.T) {
	cfg := NewGrafanaConfig(WithSecureAuth(SecureModel{Password: "pass"}))
	cfg.UserName = "admin"
	assert.True(t, cfg.IsBasicAuth())
}

func TestIsBasicAuth_NoPasswordReturnsFalse(t *testing.T) {
	cfg := NewGrafanaConfig()
	cfg.UserName = "admin"
	// no password → secureAuth.Password == ""
	assert.False(t, cfg.IsBasicAuth())
}

func TestIsBasicAuth_NoUsernameReturnsFalse(t *testing.T) {
	cfg := NewGrafanaConfig(WithSecureAuth(SecureModel{Password: "pass"}))
	assert.False(t, cfg.IsBasicAuth())
}

// ── GetOrganizationName ───────────────────────────────────────────────────────

func TestGetOrganizationName_ExplicitName(t *testing.T) {
	cfg := NewGrafanaConfig()
	cfg.OrganizationName = "MyOrg"
	assert.Equal(t, "MyOrg", cfg.GetOrganizationName())
}

func TestGetOrganizationName_DefaultWhenBasicAuth(t *testing.T) {
	cfg := NewGrafanaConfig(WithSecureAuth(SecureModel{Password: "pass"}))
	cfg.UserName = "admin"
	// No explicit org name + basic auth → DefaultOrganizationName
	assert.Equal(t, DefaultOrganizationName, cfg.GetOrganizationName())
}

func TestGetOrganizationName_UnknownWhenTokenAuth(t *testing.T) {
	cfg := NewGrafanaConfig(WithSecureAuth(SecureModel{Token: "mytoken"}))
	// no username, no explicit org → "unknown"
	assert.Equal(t, "unknown", cfg.GetOrganizationName())
}

// ── SetGrafanaAdmin / IsGrafanaAdmin ──────────────────────────────────────────

func TestIsGrafanaAdmin_DefaultFalse(t *testing.T) {
	cfg := NewGrafanaConfig()
	assert.False(t, cfg.IsGrafanaAdmin())
}

func TestSetGrafanaAdmin_SetsTrue(t *testing.T) {
	cfg := NewGrafanaConfig()
	cfg.SetGrafanaAdmin(true)
	assert.True(t, cfg.IsGrafanaAdmin())
}

func TestSetGrafanaAdmin_CanSetFalse(t *testing.T) {
	cfg := NewGrafanaConfig()
	cfg.SetGrafanaAdmin(true)
	cfg.SetGrafanaAdmin(false)
	assert.False(t, cfg.IsGrafanaAdmin())
}

// ── Filter management ─────────────────────────────────────────────────────────

func TestIsFilterSet_DefaultFalse(t *testing.T) {
	cfg := NewGrafanaConfig()
	assert.False(t, cfg.IsFilterSet())
}

func TestSetUseFilters_SetsFilterTrue(t *testing.T) {
	cfg := NewGrafanaConfig()
	cfg.SetUseFilters()
	assert.True(t, cfg.IsFilterSet())
}

func TestSetFilterFolder_SetsNameAndEnablesFilter(t *testing.T) {
	cfg := NewGrafanaConfig()
	cfg.SetFilterFolder("MyFolder")
	assert.True(t, cfg.IsFilterSet())
	assert.Equal(t, "MyFolder", cfg.getFilter().Name)
}

func TestClearFilters_ResetsFilter(t *testing.T) {
	cfg := NewGrafanaConfig()
	cfg.SetFilterFolder("MyFolder")
	cfg.ClearFilters()
	assert.False(t, cfg.IsFilterSet())
	assert.Empty(t, cfg.getFilter().Name)
}

// ── GetMonitoredFolders ───────────────────────────────────────────────────────

func TestGetMonitoredFolders_EmptyDefaultsToGeneral(t *testing.T) {
	cfg := NewGrafanaConfig()
	folders := cfg.GetMonitoredFolders(false)
	assert.Equal(t, []string{"General"}, folders)
}

func TestGetMonitoredFolders_ReturnsConfiguredFolders(t *testing.T) {
	cfg := NewGrafanaConfig()
	cfg.MonitoredFolders = []string{"Dashboards", "Alerts"}
	folders := cfg.GetMonitoredFolders(false)
	assert.Equal(t, []string{"Dashboards", "Alerts"}, folders)
}

func TestGetMonitoredFolders_FilterOverridesWhenSet(t *testing.T) {
	cfg := NewGrafanaConfig()
	cfg.MonitoredFolders = []string{"Dashboards"}
	cfg.SetFilterFolder("SpecificFolder")
	// When a folder filter is set and ignoreFilterVal=false → return filter name
	folders := cfg.GetMonitoredFolders(false)
	assert.Equal(t, []string{"SpecificFolder"}, folders)
}

func TestGetMonitoredFolders_IgnoreFilterValBypassesFilter(t *testing.T) {
	cfg := NewGrafanaConfig()
	cfg.MonitoredFolders = []string{"Dashboards"}
	cfg.SetFilterFolder("SpecificFolder")
	// ignoreFilterVal=true → return MonitoredFolders, not the filter
	folders := cfg.GetMonitoredFolders(true)
	assert.Equal(t, []string{"Dashboards"}, folders)
}

// ── GetOrgMonitoredFolders ────────────────────────────────────────────────────

func TestGetOrgMonitoredFolders_NoOverrides(t *testing.T) {
	cfg := NewGrafanaConfig()
	result := cfg.GetOrgMonitoredFolders("Main Org.")
	assert.Nil(t, result)
}

func TestGetOrgMonitoredFolders_MatchingOrg(t *testing.T) {
	cfg := NewGrafanaConfig()
	cfg.MonitoredFoldersOverride = []MonitoredOrgFolders{
		{OrganizationName: "OrgA", Folders: []string{"Dash1", "Dash2"}},
	}
	result := cfg.GetOrgMonitoredFolders("OrgA")
	assert.Equal(t, []string{"Dash1", "Dash2"}, result)
}

func TestGetOrgMonitoredFolders_NonMatchingOrg(t *testing.T) {
	cfg := NewGrafanaConfig()
	cfg.MonitoredFoldersOverride = []MonitoredOrgFolders{
		{OrganizationName: "OrgA", Folders: []string{"Dash1"}},
	}
	result := cfg.GetOrgMonitoredFolders("OrgB")
	assert.Nil(t, result)
}

func TestGetOrgMonitoredFolders_EmptyFoldersIgnored(t *testing.T) {
	cfg := NewGrafanaConfig()
	cfg.MonitoredFoldersOverride = []MonitoredOrgFolders{
		{OrganizationName: "OrgA", Folders: []string{}},
	}
	result := cfg.GetOrgMonitoredFolders("OrgA")
	assert.Nil(t, result, "entries with empty folder list should be ignored")
}

// ── GetDashboardSettings ──────────────────────────────────────────────────────

func TestGetDashboardSettings_InitialisesWhenNil(t *testing.T) {
	cfg := NewGrafanaConfig()
	assert.Nil(t, cfg.DashboardSettings)
	ds := cfg.GetDashboardSettings()
	require.NotNil(t, ds)
	assert.False(t, ds.IgnoreFilters)
}

func TestGetDashboardSettings_ReturnsExistingWhenSet(t *testing.T) {
	cfg := NewGrafanaConfig()
	cfg.DashboardSettings = &DashboardSettings{IgnoreFilters: true}
	ds := cfg.GetDashboardSettings()
	assert.True(t, ds.IgnoreFilters)
}

// ── GetConnectionSettings ─────────────────────────────────────────────────────

func TestGetConnectionSettings_InitialisesWhenNil(t *testing.T) {
	cfg := NewGrafanaConfig()
	cs := cfg.GetConnectionSettings()
	require.NotNil(t, cs)
}

func TestGetConnectionSettings_ReturnsExisting(t *testing.T) {
	cfg := NewGrafanaConfig()
	existing := &ConnectionSettings{FilterRules: []MatchingRule{{Field: "name", Regex: ".*"}}}
	cfg.ConnectionSettings = existing
	cs := cfg.GetConnectionSettings()
	assert.Equal(t, existing, cs)
}

// ── SecureLocation ────────────────────────────────────────────────────────────

func TestSecureLocation_DefaultUsesOutputPath(t *testing.T) {
	cfg := NewGrafanaConfig(WithContextName("default"))
	cfg.OutputPath = "/backups"
	loc := cfg.SecureLocation()
	// SecureSecretsResource = "secure"; not namespaced → /backups/secure
	assert.Equal(t, "/backups/secure", loc)
}

func TestSecureLocation_AbsoluteOverride(t *testing.T) {
	cfg := NewGrafanaConfig()
	cfg.SecureLocationOverride = "/etc/gdg/secure"
	loc := cfg.SecureLocation()
	assert.Equal(t, "/etc/gdg/secure", loc)
}

func TestSecureLocation_RelativeOverride(t *testing.T) {
	cfg := NewGrafanaConfig()
	cfg.OutputPath = "/backups"
	cfg.SecureLocationOverride = "mysecure"
	loc := cfg.SecureLocation()
	assert.Equal(t, "/backups/mysecure", loc)
}

// ── GetAuthLocation ───────────────────────────────────────────────────────────

func TestGetAuthLocation_IncludesContextName(t *testing.T) {
	cfg := NewGrafanaConfig(WithContextName("staging"))
	cfg.OutputPath = "/backups"
	loc := cfg.GetAuthLocation()
	assert.Contains(t, loc, "auth_staging")
}

// ── GetCloudAuthLocation ──────────────────────────────────────────────────────

func TestGetCloudAuthLocation_EmptyStorageReturnsEmpty(t *testing.T) {
	cfg := NewGrafanaConfig()
	cfg.OutputPath = "/backups"
	assert.Empty(t, cfg.GetCloudAuthLocation())
}

func TestGetCloudAuthLocation_WithStorage(t *testing.T) {
	cfg := NewGrafanaConfig()
	cfg.OutputPath = "/backups"
	cfg.Storage = "minio"
	loc := cfg.GetCloudAuthLocation()
	assert.Contains(t, loc, "s3_minio")
}

// ── GetPath ───────────────────────────────────────────────────────────────────

func TestGetPath_DelegatesToResourceType(t *testing.T) {
	cfg := NewGrafanaConfig()
	cfg.OutputPath = "/backups"
	got := cfg.GetPath(resourceTypes.UserResource, "")
	assert.Equal(t, "/backups/users", got)
}

// ── TestGetSecureAuth / TestSetSecureAuth (test helpers) ──────────────────────

func TestTestGetSetSecureAuth_RoundTrip(t *testing.T) {
	cfg := NewGrafanaConfig()
	auth := SecureModel{Password: "pw", Token: "tok"}
	err := cfg.TestSetSecureAuth(auth)
	require.NoError(t, err)
	got := cfg.TestGetSecureAuth()
	require.NotNil(t, got)
	assert.Equal(t, "pw", got.Password)
	assert.Equal(t, "tok", got.Token)
}

// ── GetUserSettings ───────────────────────────────────────────────────────────

func TestGetUserSettings_NilReturnsDefault(t *testing.T) {
	cfg := NewGrafanaConfig()
	us := cfg.GetUserSettings()
	require.NotNil(t, us)
	assert.False(t, us.RandomPassword)
}

func TestGetUserSettings_AppliesMinMaxDefaults(t *testing.T) {
	cfg := NewGrafanaConfig()
	cfg.UserSettings = &UserSettings{RandomPassword: true}
	us := cfg.GetUserSettings()
	assert.True(t, us.RandomPassword)
	// MinLength and MaxLength should be set to non-zero defaults
	assert.Greater(t, us.MinLength, 0)
	assert.Greater(t, us.MaxLength, 0)
}

func TestGetUserSettings_DoesNotOverrideExplicitValues(t *testing.T) {
	cfg := NewGrafanaConfig()
	cfg.UserSettings = &UserSettings{MinLength: 8, MaxLength: 32}
	us := cfg.GetUserSettings()
	assert.Equal(t, 8, us.MinLength)
	assert.Equal(t, 32, us.MaxLength)
}
