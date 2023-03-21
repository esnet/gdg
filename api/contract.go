package api

import (
	"context"
	"fmt"
	"github.com/esnet/gdg/apphelpers"
	"github.com/esnet/gdg/config"
	"github.com/grafana-tools/sdk"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"sync"

	gclient "github.com/grafana/grafana-api-golang-client"
)

type OrganizationsApi interface {
	//Organizations
	ListOrganizations() []sdk.Org
}

type DashboardsApi interface {
	//Dashboard
	ListDashboards(filter Filter) []sdk.FoundBoard
	ImportDashboards(filter Filter) []string
	ExportDashboards(filter Filter)
	DeleteAllDashboards(filter Filter) []string
}

type DataSourcesApi interface {
	//DataSources
	ListDataSources(filter Filter) []sdk.Datasource
	ImportDataSources(filter Filter) []string
	ExportDataSources(filter Filter) []string
	DeleteAllDataSources(filter Filter) []string
}

// AlertNotificationsApi
// Deprecated: Marked as Deprecated as of Grafana 9.0, Moving to ContactPoints is recommended
type AlertNotificationsApi interface {
	//AlertNotifications
	ListAlertNotifications() []sdk.AlertNotification
	ImportAlertNotifications() []string
	ExportAlertNotifications() []string
	DeleteAllAlertNotifications() []string
}

type AuthenticationApi interface {
	//Auth
	Login()
	AdminLogin()
}

type UsersApi interface {
	//User
	ListUsers() []sdk.User
	ImportUsers() []string
	ExportUsers() []sdk.User
	PromoteUser(userLogin string) (*sdk.StatusMessage, error)
	DeleteAllUsers() []string
}

type FoldersApi interface {
	//Folder
	ListFolder(filter Filter) []gclient.Folder
	ImportFolder(filter Filter) []string
	ExportFolder(filter Filter) []string
	DeleteAllFolder(filter Filter) []string
}

type ApiService interface {
	OrganizationsApi
	DashboardsApi
	DataSourcesApi
	AlertNotificationsApi
	UsersApi
	FoldersApi

	//MetaData
	GetServerInfo() map[string]interface{}
}

var (
	instance *DashNGoImpl
	once     sync.Once
)

type DashNGoImpl struct {
	legacyClient      *sdk.Client
	legacyAdminClient *sdk.Client
	client            *gclient.Client
	adminClient       *gclient.Client
	grafanaConf       *config.GrafanaConfig
	configRef         *viper.Viper
	debug             bool
	storage           Storage
}

func (s *DashNGoImpl) GetLegacyAdminClient() *sdk.Client {
	if s.legacyAdminClient == nil {
		log.Fatal("Requested API requires admin to have basic http auth (username/password) configured. Token access is not supported")
	}
	return s.legacyAdminClient
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
	obj.Login()
	obj.AdminLogin()

	obj.debug = config.Config().IsDebug()
	configureStorage(obj)

	return obj
}

//Testing Only
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
