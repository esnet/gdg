package api

import (
	"context"
	"fmt"
	"github.com/esnet/gdg/apphelpers"
	"github.com/esnet/gdg/config"
	"github.com/grafana-tools/sdk"
	"github.com/spf13/viper"
	"sync"
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
	ctx := context.Background()
	ctx = context.WithValue(ctx, StorageContext, appData)
	switch storageType {
	case "cloud":
		obj.storage = NewCloudStorage(ctx)
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
