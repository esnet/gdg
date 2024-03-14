package service

import (
	"github.com/esnet/gdg/internal/service/filters"
	"github.com/esnet/gdg/internal/service/types"
	gdgType "github.com/esnet/gdg/internal/types"
	"github.com/grafana/grafana-openapi-client-go/models"
)

type ServerInfoApi interface {
	GetServerInfo() map[string]interface{}
}

type GrafanaService interface {
	OrganizationsApi
	DashboardsApi
	ConnectionsApi
	UsersApi
	FoldersApi
	LibraryElementsApi
	TeamsApi

	AuthenticationApi
	//MetaData
	ServerInfoApi
}

// ConnectionsApi Contract definition
type ConnectionsApi interface {
	ListConnections(filter filters.Filter) []models.DataSourceListItemDTO
	DownloadConnections(filter filters.Filter) []string
	UploadConnections(filter filters.Filter) []string
	DeleteAllConnections(filter filters.Filter) []string
	ConnectionPermissions
}

type ConnectionPermissions interface {
	// Permissions Enterprise only
	ListConnectionPermissions(filter filters.Filter) map[*models.DataSourceListItemDTO]*models.DataSourcePermissionsDTO
	DownloadConnectionPermissions(filter filters.Filter) []string
	UploadConnectionPermissions(filter filters.Filter) []string
	DeleteAllConnectionPermissions(filter filters.Filter) []string
}

// DashboardsApi Contract definition
type DashboardsApi interface {
	ListDashboards(filter filters.Filter) []*models.Hit
	DownloadDashboards(filter filters.Filter) []string
	UploadDashboards(filter filters.Filter)
	DeleteAllDashboards(filter filters.Filter) []string
	LintDashboards(req types.LintRequest) []string
}

// FoldersApi Contract definition
type FoldersApi interface {
	ListFolder(filter filters.Filter) []*models.Hit
	DownloadFolders(filter filters.Filter) []string
	UploadFolders(filter filters.Filter) []string
	DeleteAllFolders(filter filters.Filter) []string
	//Permissions
	ListFolderPermissions(filter filters.Filter) map[*models.Hit][]*models.DashboardACLInfoDTO
	DownloadFolderPermissions(filter filters.Filter) []string
	UploadFolderPermissions(filter filters.Filter) []string
}

type LibraryElementsApi interface {
	ListLibraryElements(filter filters.Filter) []*models.LibraryElementDTO
	ListLibraryElementsConnections(filter filters.Filter, connectionID string) []*models.DashboardFullWithMeta
	DownloadLibraryElements(filter filters.Filter) []string
	UploadLibraryElements(filter filters.Filter) []string
	DeleteAllLibraryElements(filter filters.Filter) []string
}

// AuthenticationApi Contract definition
type AuthenticationApi interface {
	TokenApi
	ServiceAccountApi
	Login()
}

// OrgPreferencesApi Contract definition
type OrgPreferencesApi interface {
	GetOrgPreferences(orgName string) (*models.Preferences, error)
	UploadOrgPreferences(orgName string, pref *models.Preferences) error
}

type organizationCrudApi interface {
	ListOrganizations(filter filters.Filter) []*gdgType.OrgsDTOWithPreferences
	DownloadOrganizations(filter filters.Filter) []string
	UploadOrganizations(filter filters.Filter) []string
}

type organizationToolsApi interface {
	//Manage Active Organization
	SetOrganizationByName(name string, useSlug bool) error
	GetUserOrganization() *models.OrgDetailsDTO
	GetTokenOrganization() *models.OrgDetailsDTO
	SetUserOrganizations(id int64) error
	ListUserOrganizations() ([]*models.UserOrgDTO, error)
}

// organizationUserCrudApi  Manages user memberships to an org
type organizationUserCrudApi interface {
	ListOrgUsers(orgId int64) []*models.OrgUserDTO
	AddUserToOrg(role, orgSlug string, userId int64) error
	DeleteUserFromOrg(orgId string, userId int64) error
	UpdateUserInOrg(role, orgSlug string, userId int64) error
}

// OrganizationsApi Contract definition
type OrganizationsApi interface {
	organizationCrudApi
	organizationToolsApi
	organizationUserCrudApi
	OrgPreferencesApi
	InitOrganizations()
}

type ServiceAccountApi interface {
	ListServiceAccounts() []*gdgType.ServiceAccountDTOWithTokens
	ListServiceAccountsTokens(id int64) ([]*models.TokenDTO, error)
	DeleteAllServiceAccounts() []string
	DeleteServiceAccountTokens(serviceId int64) []string
	CreateServiceAccountToken(name int64, role string, expiration int64) (*models.NewAPIKeyResult, error)
	CreateServiceAccount(name, role string, expiration int64) (*models.ServiceAccountDTO, error)
}

type TeamsApi interface {
	//Team
	DownloadTeams(filter filters.Filter) map[*models.TeamDTO][]*models.TeamMemberDTO
	UploadTeams(filter filters.Filter) map[*models.TeamDTO][]*models.TeamMemberDTO
	ListTeams(filter filters.Filter) map[*models.TeamDTO][]*models.TeamMemberDTO
	DeleteTeam(filter filters.Filter) ([]*models.TeamDTO, error)
}

type TokenApi interface {
	ListAPIKeys() []*models.APIKeyDTO
	DeleteAllTokens() []string
	CreateAPIKey(name, role string, expiration int64) (*models.NewAPIKeyResult, error)
}

// UsersApi Contract definition
type UsersApi interface {
	//UserApi
	ListUsers(filter filters.Filter) []*models.UserSearchHitDTO
	DownloadUsers(filter filters.Filter) []string
	UploadUsers(filter filters.Filter) []gdgType.UserProfileWithAuth
	DeleteAllUsers(filter filters.Filter) []string
	// Tools
	PromoteUser(userLogin string) (string, error)
	GetUserInfo() (*models.UserProfileDTO, error)
}
