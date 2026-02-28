package ports

import (
	customModels "github.com/esnet/gdg/internal/domain"
	"github.com/grafana/grafana-openapi-client-go/models"
)

type ServerInfoApi interface {
	GetServerInfo() map[string]any
}

type GrafanaService interface {
	OrganizationsApi
	DashboardsApi
	DashboardPermissionsApi
	ConnectionsApi
	UsersApi
	FoldersApi
	LibraryElementsApi
	TeamsApi
	AlertingApi

	AuthenticationApi
	// MetaData
	ServerInfoApi
	LicenseApi
}

type LicenseApi interface {
	IsEnterprise() bool
}

// ConnectionsApi Contract definition
type ConnectionsApi interface {
	ListConnections(filter V2Filter) []models.DataSourceListItemDTO
	DownloadConnections(filter V2Filter) []string
	UploadConnections(filter V2Filter) []string
	DeleteAllConnections(filter V2Filter) []string
	ConnectionPermissions
}

type ConnectionPermissions interface {
	// Permissions Enterprise only
	ListConnectionPermissions(filter V2Filter) []customModels.ConnectionPermissionItem
	DownloadConnectionPermissions(filter V2Filter) []string
	UploadConnectionPermissions(filter V2Filter) []string
	DeleteAllConnectionPermissions(filter V2Filter) []string
}

// DashboardsApi Contract definition
type DashboardsApi interface {
	ListDashboards(filter V2Filter) []*customModels.NestedHit
	DownloadDashboards(filter V2Filter) []string
	UploadDashboards(filterReq V2Filter) ([]string, error)
	DeleteAllDashboards(filter V2Filter) []string
}

type DashboardPermissionsApi interface {
	ListDashboardPermissions(filterReq V2Filter) ([]customModels.DashboardAndPermissions, error)
	DownloadDashboardPermissions(filterReq V2Filter) ([]string, error)
	ClearDashboardPermissions(filterReq V2Filter) error
	UploadDashboardPermissions(filterReq V2Filter) ([]string, error)
}

// FoldersApi Contract definition
type FoldersApi interface {
	ListFolders(filter V2Filter) []*customModels.NestedHit
	DownloadFolders(filter V2Filter) []string
	UploadFolders(filter V2Filter) []string
	DeleteAllFolders(filter V2Filter) []string
	// Permissions
	ListFolderPermissions(filter V2Filter) map[*customModels.NestedHit][]*models.DashboardACLInfoDTO
	DownloadFolderPermissions(filter V2Filter) []string
	UploadFolderPermissions(filter V2Filter) []string
}

type LibraryElementsApi interface {
	ListLibraryElements(filter V2Filter) []*customModels.WithNested[models.LibraryElementDTO]
	ListLibraryElementsConnections(filter V2Filter, connectionID string) []*models.DashboardFullWithMeta
	DownloadLibraryElements(filter V2Filter) []string
	UploadLibraryElements(filter V2Filter) []string
	DeleteAllLibraryElements(filter V2Filter) []string
}

// AuthenticationApi Contract definition
type AuthenticationApi interface {
	// TokenApi
	ServiceAccountApi
	Login()
	EncodeValue(in string) string
	DecodeValue(in string) string
}

// OrgPreferencesApi Contract definition
type OrgPreferencesApi interface {
	GetOrgPreferences(orgName string) (*models.PreferencesSpec, error)
	UploadOrgPreferences(orgName string, pref *models.PreferencesSpec) error
}

type organizationCrudApi interface {
	ListOrganizations(filter V2Filter, withPreferences bool) []*customModels.OrgsDTOWithPreferences
	DownloadOrganizations(filter V2Filter) []string
	UploadOrganizations(filter V2Filter) []string
}

type organizationToolsApi interface {
	// Manage Active Organization
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
	ListServiceAccounts() []*customModels.ServiceAccountDTOWithTokens
	ListServiceAccountsTokens(id int64) ([]*models.TokenDTO, error)
	DeleteServiceAccount(accountId int64) error
	DeleteAllServiceAccounts() []string
	DeleteServiceAccountTokens(serviceId int64) []string
	CreateServiceAccountToken(serviceAccountId int64, name string, expiration int64) (*models.NewAPIKeyResult, error)
	CreateServiceAccount(name, role string, expiration int64) (*models.ServiceAccountDTO, error)
}

type TeamsApi interface {
	// Team
	DownloadTeams(filter V2Filter) map[*models.TeamDTO][]*models.TeamMemberDTO
	UploadTeams(filter V2Filter) map[*models.TeamDTO][]*models.TeamMemberDTO
	ListTeams(filter V2Filter) map[*models.TeamDTO][]*models.TeamMemberDTO
	DeleteTeam(filter V2Filter) ([]*models.TeamDTO, error)
}

// UsersApi Contract definition
type UsersApi interface {
	// UserApi
	ListUsers(filter V2Filter) []*models.UserSearchHitDTO
	DownloadUsers(filter V2Filter) []string
	UploadUsers(filter V2Filter) []customModels.UserProfileWithAuth
	DeleteAllUsers(filter V2Filter) []string
	// Tools
	PromoteUser(userLogin string) (string, error)
	GetUserInfo() (*models.UserProfileDTO, error)
}
