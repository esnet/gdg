package api

import (
	"context"
	"fmt"

	"github.com/esnet/gdg/apphelpers"
	"github.com/esnet/gdg/config"
	"github.com/esnet/gdg/internal/apiExtend"
	"github.com/esnet/grafana-swagger-api-golang/goclient/client"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"sync"
)

type ApiService interface {
	OrganizationsApi
	DashboardsApi
	DataSourcesApi
	AlertNotificationsApi
	UsersApi
	FoldersApi
	LibraryElementsApi
	TeamsApi

	//MetaData
	GetServerInfo() map[string]interface{}
}

var (
	instance *DashNGoImpl
	once     sync.Once
)

type DashNGoImpl struct {
	client   *client.GrafanaHTTPAPI
	extended *apiExtend.ExtendedApi

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
	obj.Login()

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
