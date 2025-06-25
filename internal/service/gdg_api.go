package service

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"sync"

	"github.com/esnet/gdg/internal/config/domain"

	"github.com/esnet/gdg/internal/api"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/storage"
)

var (
	instance        *DashNGoImpl
	initServiceOnce sync.Once
)

type DashNGoImpl struct {
	extended    *api.ExtendedApi
	grafanaConf *domain.GrafanaConfig
	globalConf  *domain.AppGlobals
	storage     storage.Storage
}

var DefaultConfigProvider config.Provider = func() *config.Configuration {
	return config.Config()
}

func setupConfigData(cfg *config.Configuration, obj *DashNGoImpl) {
	obj.grafanaConf = cfg.GetDefaultGrafanaConfig()
	obj.globalConf = cfg.GetGDGConfig().GetAppGlobals()
}

func newInstance() *DashNGoImpl {
	obj := &DashNGoImpl{}
	setupConfigData(config.Config(), obj)

	if obj.globalConf.ApiDebug {
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
	storageEngine, err := ConfigureStorage(DefaultConfigProvider)
	if err != nil {
		log.Fatal("Unable to configure a valid storage engine, %w", err)
	}
	obj.SetStorage(storageEngine)

	return obj
}

func (s *DashNGoImpl) SetStorage(v storage.Storage) {
	s.storage = v
}

func ConfigureStorage(provider config.Provider) (storage.Storage, error) {
	var (
		storageEngine storage.Storage
		err           error
		cfg           *config.Configuration
	)
	if provider != nil {
		cfg = provider()
	} else {
		cfg = config.Config()
	}
	// config
	storageType, appData := cfg.GetCloudConfiguration(config.Config().GetDefaultGrafanaConfig().Storage)

	ctx := context.Background()
	ctx = context.WithValue(ctx, storage.Context, appData)
	switch storageType {
	case "cloud":
		{
			storageEngine, err = storage.NewCloudStorage(ctx)
			if err != nil {
				return nil, fmt.Errorf("unable to configure CloudStorage Engine:	%w", err)
			}
		}
	default:
		storageEngine = storage.NewLocalStorage(ctx)
	}
	return storageEngine, nil
}

func NewTestApiService(storageEngine storage.Storage, cfgProvider config.Provider) GrafanaService {
	ins := newInstance()
	ins.SetStorage(storageEngine)
	if cfgProvider == nil {
		cfgProvider = DefaultConfigProvider
	}
	setupConfigData(cfgProvider(), ins)
	return ins
}

func NewDashNGoImpl() *DashNGoImpl {
	initServiceOnce.Do(func() {
		instance = newInstance()
	})
	return instance
}
