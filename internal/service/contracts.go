package service

import (
	customModels "github.com/esnet/gdg/internal/service/domain"
	"github.com/esnet/gdg/internal/service/filters"
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
	ListConnections(filter filters.V2Filter) []models.DataSourceListItemDTO
	DownloadConnections(filter filters.V2Filter) []string
	UploadConnections(filter filters.V2Filter) []string
	DeleteAllConnections(filter filters.V2Filter) []string
	ConnectionPermissions
}

type ConnectionPermissions interface {
	// Permissions Enterprise only
	ListConnectionPermissions(filter filters.V2Filter) []customModels.ConnectionPermissionItem
	DownloadConnectionPermissions(filter filters.V2Filter) []string
	UploadConnectionPermissions(filter filters.V2Filter) []string
	DeleteAllConnectionPermissions(filter filters.V2Filter) []string
}

// DashboardsApi Contract definition
type DashboardsApi interface {
	ListDashboards(filter filters.V2Filter) []*customModels.NestedHit
	DownloadDashboards(filter filters.V2Filter) []string
	UploadDashboards(filterReq filters.V2Filter) ([]string, error)
	DeleteAllDashboards(filter filters.V2Filter) []string
}

type AlertContactPoints interface {
	ListContactPoints() ([]*models.EmbeddedContactPoint, error)
	DownloadContactPoints() (string, error)
	ClearContactPoints() ([]string, error)
	UploadContactPoints() ([]string, error)
}

type AlertRules interface {
	DownloadAlertRules(filter filters.V2Filter) ([]string, error)
	ListAlertRules(filter filters.V2Filter) ([]*customModels.AlertRuleWithNestedFolder, error)
	ClearAlertRules(filter filters.V2Filter) ([]string, error)
	UploadAlertRules(filter filters.V2Filter) error
}

type AlertTemplates interface {
	DownloadAlertTemplates() (string, error)
	ListAlertTemplates() ([]*models.NotificationTemplate, error)
	ClearAlertTemplates() ([]string, error)
	UploadAlertTemplates() ([]string, error)
}

type AlertPolicies interface {
	DownloadAlertNotifications() (string, error)
	ListAlertNotifications() (*models.Route, error)
	ClearAlertNotifications() error
	UploadAlertNotifications() (*models.Route, error)
}

type AlertTimings interface {
	DownloadAlertTimings() (string, error)
	ListAlertTimings() ([]*models.MuteTimeInterval, error)
	ClearAlertTimings() error
	UploadAlertTimings() ([]string, error)
}

type AlertingApi interface {
	AlertContactPoints
	AlertRules
	AlertTemplates
	AlertPolicies
	AlertTimings
}

type DashboardPermissionsApi interface {
	ListDashboardPermissions(filterReq filters.V2Filter) ([]customModels.DashboardAndPermissions, error)
	DownloadDashboardPermissions(filterReq filters.V2Filter) ([]string, error)
	ClearDashboardPermissions(filterReq filters.V2Filter) error
	UploadDashboardPermissions(filterReq filters.V2Filter) ([]string, error)
}

// FoldersApi Contract definition
type FoldersApi interface {
	ListFolders(filter filters.V2Filter) []*customModels.NestedHit
	DownloadFolders(filter filters.V2Filter) []string
	UploadFolders(filter filters.V2Filter) []string
	DeleteAllFolders(filter filters.V2Filter) []string
	// Permissions
	ListFolderPermissions(filter filters.V2Filter) map[*customModels.NestedHit][]*models.DashboardACLInfoDTO
	DownloadFolderPermissions(filter filters.V2Filter) []string
	UploadFolderPermissions(filter filters.V2Filter) []string
}

type LibraryElementsApi interface {
	ListLibraryElements(filter filters.V2Filter) []*customModels.WithNested[models.LibraryElementDTO]
	ListLibraryElementsConnections(filter filters.V2Filter, connectionID string) []*models.DashboardFullWithMeta
	DownloadLibraryElements(filter filters.V2Filter) []string
	UploadLibraryElements(filter filters.V2Filter) []string
	DeleteAllLibraryElements(filter filters.V2Filter) []string
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
	ListOrganizations(filter filters.V2Filter, withPreferences bool) []*customModels.OrgsDTOWithPreferences
	DownloadOrganizations(filter filters.V2Filter) []string
	UploadOrganizations(filter filters.V2Filter) []string
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
	DownloadTeams(filter filters.V2Filter) map[*models.TeamDTO][]*models.TeamMemberDTO
	UploadTeams(filter filters.V2Filter) map[*models.TeamDTO][]*models.TeamMemberDTO
	ListTeams(filter filters.V2Filter) map[*models.TeamDTO][]*models.TeamMemberDTO
	DeleteTeam(filter filters.V2Filter) ([]*models.TeamDTO, error)
}

// UsersApi Contract definition
type UsersApi interface {
	// UserApi
	ListUsers(filter filters.V2Filter) []*models.UserSearchHitDTO
	DownloadUsers(filter filters.V2Filter) []string
	UploadUsers(filter filters.V2Filter) []customModels.UserProfileWithAuth
	DeleteAllUsers(filter filters.V2Filter) []string
	// Tools
	PromoteUser(userLogin string) (string, error)
	GetUserInfo() (*models.UserProfileDTO, error)
}
