package api

import (
	"context"
	"fmt"
	"sync"

	"github.com/esnet/gdg/apphelpers"
	"github.com/esnet/gdg/config"
	"github.com/grafana-tools/sdk"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type ApiService interface {
	//Organizations
	ListOrganizations() []sdk.Org
	//Dashboard
	ListDashboards(filter Filter) []sdk.FoundBoard
	ImportDashboards(filter Filter) []string
	ExportDashboards(filter Filter)
	DeleteAllDashboards(filter Filter) []string
	//DataSources
	ListDataSources(filter Filter) []sdk.Datasource
	ImportDataSources(filter Filter) []string
	ExportDataSources(filter Filter) []string
	DeleteAllDataSources(filter Filter) []string
	//AlertNotifications
	ListAlertNotifications() []sdk.AlertNotification
	ImportAlertNotifications() []string
	ExportAlertNotifications() []string
	DeleteAllAlertNotifications() []string
	//Login
	Login() *sdk.Client
	AdminLogin() *sdk.Client
	//User
	ListUsers() []sdk.User
	ImportUsers() []string
	ExportUsers() []sdk.User
	PromoteUser(userLogin string) (*sdk.StatusMessage, error)
	DeleteAllUsers() []string
	//Team
	ImportTeams() map[sdk.Team][]sdk.TeamMember
	ExportTeams() map[sdk.Team][]sdk.TeamMember
	ListTeams() []sdk.Team
	DeleteTeam(teamName string) (*sdk.StatusMessage, error)
	//TeamMembers
	ListTeamMembers(teamName string) []sdk.TeamMember
	AddTeamMember(teamName string, userLogin string) (*sdk.StatusMessage, error)
	DeleteTeamMember(teamName string, userLogin string) (*sdk.StatusMessage, error)
	//MetaData
	GetServerInfo() map[string]interface{}
	//Folder
	ListFolder(filter Filter) []sdk.Folder
	ImportFolder(filter Filter) []string
	ExportFolder(filter Filter) []string
	DeleteAllFolder(filter Filter) []string
}

var (
	instance *DashNGoImpl
	once     sync.Once
)

type DashNGoImpl struct {
	client      *sdk.Client
	adminClient *sdk.Client
	grafanaConf *config.GrafanaConfig
	configRef   *viper.Viper
	debug       bool
	storage     Storage
}

func (s *DashNGoImpl) GetAdminClient() *sdk.Client {
	if s.adminClient == nil {
		log.Fatal("Requested API requires admin to have basic http auth (username/password) configured. Token access is not supported")
	}
	return s.adminClient
}

func NewDashNGoImpl() *DashNGoImpl {
	once.Do(func() {
		instance = newInstance()
	})
	return instance
}

func newInstance() *DashNGoImpl {
	obj := &DashNGoImpl{}
	obj.grafanaConf = apphelpers.GetCtxDefaultGrafanaConfig()
	obj.configRef = config.Config().ViperConfig()
	obj.client = obj.Login()
	obj.adminClient = obj.AdminLogin()
	obj.debug = config.Config().IsDebug()
	configureStorage(obj)

	return obj
}

// Testing Only
func (s *DashNGoImpl) SetStorage(v Storage) {
	s.storage = v
}

func configureStorage(obj *DashNGoImpl) {
	//config
	appData := config.Config().ViperConfig().GetStringMap(fmt.Sprintf("storage_engine.%s", obj.grafanaConf.Storage))

	storageType := "local"
	if len(appData) != 0 {
		storageType = appData["kind"].(string)
	}
	var err error
	ctx := context.Background()
	ctx = context.WithValue(ctx, StorageContext, appData)
	switch storageType {
	case "cloud":
		{
			obj.storage, err = NewCloudStorage(ctx)
			if err != nil {
				log.Warn("falling back on Local Storage, Cloud storage configuration error")
				obj.storage = NewLocalStorage(ctx)
			}

		}
	default:
		obj.storage = NewLocalStorage(ctx)
	}
}

func NewApiService(override ...string) ApiService {
	//Used for Testing purposes
	if len(override) > 0 {
		return newInstance()
	}
	return NewDashNGoImpl()
}
