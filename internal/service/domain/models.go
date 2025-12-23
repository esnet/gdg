package domain

import (
	"github.com/grafana/grafana-openapi-client-go/models"
)

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
	Password string
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

// DashboardAndPermissions holds a dashboard reference and its permission list.
type DashboardAndPermissions struct {
	Dashboard   *NestedHit
	Permissions []*models.DashboardACLInfoDTO
}

type AlertRuleWithNestedFolder struct {
	*models.ProvisionedAlertRule
	NestedPath string
}
