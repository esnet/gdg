package api

import (
	"github.com/grafana-tools/sdk"
	"github.com/netsage-project/gdg/apphelpers"
	"github.com/netsage-project/gdg/config"
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
	ListAlertNotifications(filter Filter) []sdk.AlertNotification
	ImportAlertNotifications(filter Filter) []string
	ExportAlertNotifications(filter Filter) []string
	DeleteAllAlertNotifications(filter Filter) []string
	//Login
	Login() *sdk.Client
	AdminLogin() *sdk.Client
	//User
	ListUsers() []sdk.User
	PromoteUser(userLogin string) (*sdk.StatusMessage, error)
	//MetaData
	GetServerInfo() map[string]interface{}
}

type DashNGoImpl struct {
	client      *sdk.Client
	adminClient *sdk.Client
	grafanaConf *config.GrafanaConfig
	configRef   *viper.Viper
	debug       bool
}

func (s *DashNGoImpl) init() {
	s.grafanaConf = apphelpers.GetCtxDefaultGrafanaConfig()
	s.configRef = config.Config().ViperConfig()
	s.client = s.Login()
	s.adminClient = s.AdminLogin()
	s.debug = config.Config().IsDebug()

}

func NewApiService() ApiService {
	d := &DashNGoImpl{}
	d.init()
	return d

}
