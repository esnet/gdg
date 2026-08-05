package domain

import (
	"github.com/grafana/grafana-openapi-client-go/models"
)

const (
	// GdgPermApiVersionV1 tags a stored dashboard-permissions file as originating
	// from the legacy ACL API (GET /api/dashboards/uid/{uid}/permissions).
	// Used by Grafana v12 and earlier, or v13 as a fallback.
	GdgPermApiVersionV1 = "v1"

	// GdgPermApiVersionV2 tags a stored dashboard-permissions file as originating
	// from the RBAC access-control API (GET /api/access-control/dashboards/{uid}).
	// Used by Grafana v13+.
	GdgPermApiVersionV2 = "dashboard-permissions.grafana/v2"
)

// DashboardPermissionsGdg is the GDG on-disk wrapper for a dashboard's permission set.
// It is written by DownloadDashboardPermissions and read by UploadDashboardPermissions.
//
//   - DashboardUID    is the Grafana dashboard UID — used to target the correct
//     dashboard when uploading permissions.
//   - DashboardName   is the human-readable title, stored for readability only.
//   - GdgApiVersion   records the API version at download time (GdgPermApiVersionV1
//     or GdgPermApiVersionV2). At upload time this is compared against the target
//     server version to route to the correct API and detect format mismatches.
//   - Permissions     is always []*models.ResourcePermissionDTO regardless of API
//     version — the v1 path maps DashboardACLInfoDTO → ResourcePermissionDTO before
//     writing, keeping the on-disk format unified.
type DashboardPermissionsGdg struct {
	DashboardUID  string                          `json:"dashboard_uid"`
	DashboardName string                          `json:"dashboard_name"`
	GdgApiVersion string                          `json:"gdg_api_version"`
	Permissions   []*models.ResourcePermissionDTO `json:"permissions"`
}

// ServiceAccountDTOWithTokens represents a service account and its associated tokens.
type ServiceAccountDTOWithTokens struct {
	ServiceAccount *models.ServiceAccountDTO
	Tokens         []*models.TokenDTO
}

// WithNested represents an entity with a nested path for filtering purposes.
type WithNested[T any] struct {
	Entity     *T
	NestedPath string
}

// NestedHit represents a Dashboard or Folder with an associated nested path in dashboard filtering.
type NestedHit struct {
	*models.Hit
	NestedPath string
}

// UserProfileWithAuth embeds UserProfileDTO and adds a Password field for authentication.
type UserProfileWithAuth struct {
	models.UserProfileDTO
	Password string // #nosec G117 is the user's temporary generated password
}

// OrgsDTOWithPreferences represents an organization and its preferences.
type OrgsDTOWithPreferences struct {
	Organization *models.OrgDTO          `json:"organization"`
	Preferences  *models.PreferencesSpec `json:"preferences"` // Preferences are preferences associated with a given org.  theme, dashboard, timezone, etc
}

// ConnectionPermissionItem holds a connection and its associated permissions.
type ConnectionPermissionItem struct {
	Connection  *models.DataSourceListItemDTO
	Permissions []*models.ResourcePermissionDTO
}

// DashboardAndPermissions holds a dashboard reference and its unified permission list.
// Permissions is always []*models.ResourcePermissionDTO regardless of the Grafana API
// version in use — the v1 path maps DashboardACLInfoDTO to ResourcePermissionDTO
// internally before returning, so callers always work with a single type.
type DashboardAndPermissions struct {
	Dashboard   *NestedHit
	Permissions []*models.ResourcePermissionDTO
}

type AlertRuleWithNestedFolder struct {
	*models.ProvisionedAlertRule
	NestedPath string
}
