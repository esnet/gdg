package service

import (
	"context"
	"github.com/esnet/gdg/internal/api"
	"github.com/esnet/gdg/internal/config"
	"github.com/spf13/viper"
	"log/slog"
	"os"
	"sync"
)

var (
	instance        *DashNGoImpl
	initServiceOnce sync.Once
)

type DashNGoImpl struct {
	extended *api.ExtendedApi

	grafanaConf *config.GrafanaConfig
	configRef   *viper.Viper
	debug       bool
	apiDebug    bool
	storage     Storage
}

func NewDashNGoImpl() *DashNGoImpl {
	initServiceOnce.Do(func() {
		instance = newInstance()
	})
	return instance
}

func newInstance() *DashNGoImpl {
	obj := &DashNGoImpl{}
	obj.grafanaConf = config.Config().GetDefaultGrafanaConfig()
	obj.configRef = config.Config().GetViperConfig(config.ViperGdgConfig)
	obj.debug = config.Config().IsDebug()
	obj.apiDebug = config.Config().IsApiDebug()
	if obj.apiDebug {
		err := os.Setenv("DEBUG", "1")
		if err != nil {
			slog.Debug("unable to set debug env value", slog.Any("err", err))
		}
	} else {
		err := os.Setenv("DEBUG", "0")
		if err != nil {
			slog.Debug("unable to set debug env value", slog.Any("err", err))
		}
	}
	obj.Login()
	configureStorage(obj)

	return obj
}

// Testing Only
func (s *DashNGoImpl) SetStorage(v Storage) {
	s.storage = v
}

func configureStorage(obj *DashNGoImpl) {
	//config
	storageType, appData := config.Config().GetCloudConfiguration(config.Config().GetDefaultGrafanaConfig().Storage)

	var err error
	ctx := context.Background()
	ctx = context.WithValue(ctx, StorageContext, appData)
	switch storageType {
	case "cloud":
		{
			obj.storage, err = NewCloudStorage(ctx)
			if err != nil {
				slog.Warn("falling back on Local Storage, Cloud storage configuration error")
				obj.storage = NewLocalStorage(ctx)
			}
		}
	default:
		obj.storage = NewLocalStorage(ctx)
	}
}

func NewApiService(override ...string) GrafanaService {
	//Used for Testing purposes
	if len(override) > 0 {
		return newInstance()
	}
	return NewDashNGoImpl()
}
