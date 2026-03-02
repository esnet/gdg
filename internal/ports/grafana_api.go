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
	ListConnections(filter Filter) []models.DataSourceListItemDTO
	DownloadConnections(filter Filter) []string
	UploadConnections(filter Filter) []string
	DeleteAllConnections(filter Filter) []string
	ConnectionPermissions
}

type ConnectionPermissions interface {
	// Permissions Enterprise only
	ListConnectionPermissions(filter Filter) []customModels.ConnectionPermissionItem
	DownloadConnectionPermissions(filter Filter) []string
	UploadConnectionPermissions(filter Filter) []string
	DeleteAllConnectionPermissions(filter Filter) []string
}

// DashboardsApi Contract definition
type DashboardsApi interface {
	ListDashboards(filter Filter) []*customModels.NestedHit
	DownloadDashboards(filter Filter) []string
	UploadDashboards(filterReq Filter) ([]string, error)
	DeleteAllDashboards(filter Filter) []string
}

type DashboardPermissionsApi interface {
	ListDashboardPermissions(filterReq Filter) ([]customModels.DashboardAndPermissions, error)
	DownloadDashboardPermissions(filterReq Filter) ([]string, error)
	ClearDashboardPermissions(filterReq Filter) error
	UploadDashboardPermissions(filterReq Filter) ([]string, error)
}

// FoldersApi Contract definition
type FoldersApi interface {
	ListFolders(filter Filter) []*customModels.NestedHit
	DownloadFolders(filter Filter) []string
	UploadFolders(filter Filter) []string
	DeleteAllFolders(filter Filter) []string
	// Permissions
	ListFolderPermissions(filter Filter) map[*customModels.NestedHit][]*models.DashboardACLInfoDTO
	DownloadFolderPermissions(filter Filter) []string
	UploadFolderPermissions(filter Filter) []string
}

type LibraryElementsApi interface {
	ListLibraryElements(filter Filter) []*customModels.WithNested[models.LibraryElementDTO]
	ListLibraryElementsConnections(filter Filter, connectionID string) []*models.DashboardFullWithMeta
	DownloadLibraryElements(filter Filter) []string
	UploadLibraryElements(filter Filter) []string
	DeleteAllLibraryElements(filter Filter) []string
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
	ListOrganizations(filter Filter, withPreferences bool) []*customModels.OrgsDTOWithPreferences
	DownloadOrganizations(filter Filter) []string
	UploadOrganizations(filter Filter) []string
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
	DownloadTeams(filter Filter) map[*models.TeamDTO][]*models.TeamMemberDTO
	UploadTeams(filter Filter) map[*models.TeamDTO][]*models.TeamMemberDTO
	ListTeams(filter Filter) map[*models.TeamDTO][]*models.TeamMemberDTO
	DeleteTeam(filter Filter) ([]*models.TeamDTO, error)
}

// UsersApi Contract definition
type UsersApi interface {
	// UserApi
	ListUsers(filter Filter) []*models.UserSearchHitDTO
	DownloadUsers(filter Filter) []string
	UploadUsers(filter Filter) []customModels.UserProfileWithAuth
	DeleteAllUsers(filter Filter) []string
	// Tools
	PromoteUser(userLogin string) (string, error)
	GetUserInfo() (*models.UserProfileDTO, error)
}
