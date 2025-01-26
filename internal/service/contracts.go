package service

import (
	"github.com/esnet/gdg/internal/service/filters"
	"github.com/esnet/gdg/internal/service/types"
	customModels "github.com/esnet/gdg/internal/types"
	gdgType "github.com/esnet/gdg/internal/types"
	"github.com/grafana/grafana-openapi-client-go/models"
)

type ServerInfoApi interface {
	GetServerInfo() map[string]interface{}
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
	ListConnections(filter filters.Filter) []models.DataSourceListItemDTO
	DownloadConnections(filter filters.Filter) []string
	UploadConnections(filter filters.Filter) []string
	DeleteAllConnections(filter filters.Filter) []string
	ConnectionPermissions
}

type ConnectionPermissions interface {
	// Permissions Enterprise only
	ListConnectionPermissions(filter filters.Filter) []customModels.ConnectionPermissionItem
	DownloadConnectionPermissions(filter filters.Filter) []string
	UploadConnectionPermissions(filter filters.Filter) []string
	DeleteAllConnectionPermissions(filter filters.Filter) []string
}

// DashboardsApi Contract definition
type DashboardsApi interface {
	ListDashboards(filter filters.Filter) []*models.Hit
	DownloadDashboards(filter filters.Filter) []string
	UploadDashboards(filter filters.Filter) error
	DeleteAllDashboards(filter filters.Filter) []string
	LintDashboards(req types.LintRequest) []string
}

type AlertingApi interface {
	ListContactPoints() ([]*models.EmbeddedContactPoint, error)
	DownloadContactPoints() (string, error)
	ClearContactPoints() ([]string, error)
	UploadContactPoints() ([]string, error)
}

type DashboardPermissionsApi interface {
	ListDashboardPermissions(filterReq filters.Filter) ([]gdgType.DashboardAndPermissions, error)
	DownloadDashboardPermissions(filterReq filters.Filter) ([]string, error)
	ClearDashboardPermissions(filterReq filters.Filter) error
	UploadDashboardPermissions(filterReq filters.Filter) ([]string, error)
}

// FoldersApi Contract definition
type FoldersApi interface {
	ListFolders(filter filters.Filter) []*customModels.FolderDetails
	DownloadFolders(filter filters.Filter) []string
	UploadFolders(filter filters.Filter) []string
	DeleteAllFolders(filter filters.Filter) []string
	// Permissions
	ListFolderPermissions(filter filters.Filter) map[*customModels.FolderDetails][]*models.DashboardACLInfoDTO
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
	// TokenApi
	ServiceAccountApi
	Login()
}

// OrgPreferencesApi Contract definition
type OrgPreferencesApi interface {
	GetOrgPreferences(orgName string) (*models.Preferences, error)
	UploadOrgPreferences(orgName string, pref *models.Preferences) error
}

type organizationCrudApi interface {
	ListOrganizations(filter filters.Filter, withPreferences bool) []*gdgType.OrgsDTOWithPreferences
	DownloadOrganizations(filter filters.Filter) []string
	UploadOrganizations(filter filters.Filter) []string
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
	ListServiceAccounts() []*gdgType.ServiceAccountDTOWithTokens
	ListServiceAccountsTokens(id int64) ([]*models.TokenDTO, error)
	DeleteServiceAccount(accountId int64) error
	DeleteAllServiceAccounts() []string
	DeleteServiceAccountTokens(serviceId int64) []string
	CreateServiceAccountToken(serviceAccountId int64, role string, expiration int64) (*models.NewAPIKeyResult, error)
	CreateServiceAccount(name, role string, expiration int64) (*models.ServiceAccountDTO, error)
}

type TeamsApi interface {
	// Team
	DownloadTeams(filter filters.Filter) map[*models.TeamDTO][]*models.TeamMemberDTO
	UploadTeams(filter filters.Filter) map[*models.TeamDTO][]*models.TeamMemberDTO
	ListTeams(filter filters.Filter) map[*models.TeamDTO][]*models.TeamMemberDTO
	DeleteTeam(filter filters.Filter) ([]*models.TeamDTO, error)
}

// UsersApi Contract definition
type UsersApi interface {
	// UserApi
	ListUsers(filter filters.Filter) []*models.UserSearchHitDTO
	DownloadUsers(filter filters.Filter) []string
	UploadUsers(filter filters.Filter) []gdgType.UserProfileWithAuth
	DeleteAllUsers(filter filters.Filter) []string
	// Tools
	PromoteUser(userLogin string) (string, error)
	GetUserInfo() (*models.UserProfileDTO, error)
}
