package api

import (
	"github.com/netsage-project/grafana-dashboard-manager/config"
	"github.com/grafana-tools/sdk"
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
	//Login
	Login() *sdk.Client
	AdminLogin() *sdk.Client
	//User
	ListUsers() []sdk.User
	PromoteUser(userLogin string) (*sdk.StatusMessage, error)
}

type DashNGoImpl struct {
	client      *sdk.Client
	adminClient *sdk.Client
	grafanaConf *config.GrafanaConfig
	configRef   *viper.Viper
	debug       bool
}

func (s *DashNGoImpl) init() {
	s.grafanaConf = config.GetDefaultGrafanaConfig()
	s.configRef = config.Config()
	s.client = s.Login()
	s.adminClient = s.AdminLogin()
	s.debug = s.configRef.GetBool("global.debug")

}

func NewApiService() ApiService {
	d := &DashNGoImpl{}
	d.init()
	return d

}
