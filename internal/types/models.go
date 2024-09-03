package types

import "github.com/grafana/grafana-openapi-client-go/models"

type ServiceAccountDTOWithTokens struct {
	ServiceAccount *models.ServiceAccountDTO
	Tokens         []*models.TokenDTO
}

type FolderDetails struct {
	*models.Hit
	NestedPath string
}

type UserProfileWithAuth struct {
	models.UserProfileDTO
	Password string
}

type OrgsDTOWithPreferences struct {
	Organization *models.OrgDTO      `json:"organization"`
	Preferences  *models.Preferences `json:"preferences"` // Preferences are preferences associated with a given org.  theme, dashboard, timezone, etc
}

type ConnectionPermissionItem struct {
	Connection  *models.DataSourceListItemDTO
	Permissions []*models.ResourcePermissionDTO
}
