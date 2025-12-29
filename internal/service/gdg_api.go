package service

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/esnet/gdg/internal/config/domain"
	"github.com/esnet/gdg/pkg/test_tooling/common"

	"github.com/esnet/gdg/internal/api"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/storage"
)

var instance *DashNGoImpl

type DashNGoImpl struct {
	extended    *api.ExtendedApi
	gdgConfig   *domain.GDGAppConfiguration
	grafanaConf *domain.GrafanaConfig
	storage     storage.Storage
}

func (s *DashNGoImpl) GetGlobals() *domain.AppGlobals {
	if s.gdgConfig.Global == nil {
		s.gdgConfig.Global = &domain.AppGlobals{}
	}
	return s.gdgConfig.Global
}

func (s *DashNGoImpl) GetGdgConfig() *domain.GDGAppConfiguration {
	return s.gdgConfig
}

func setupConfigData(cfg *domain.GDGAppConfiguration, obj *DashNGoImpl) {
	obj.grafanaConf = cfg.GetDefaultGrafanaConfig()
	obj.gdgConfig = cfg
}

func newInstance(cfg *domain.GDGAppConfiguration) *DashNGoImpl {
	obj := &DashNGoImpl{
		gdgConfig: cfg,
	}
	setupConfigData(cfg, obj)

	if obj.GetGlobals().ApiDebug {
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
	storageEngine, err := ConfigureStorage(cfg)
	if err != nil {
		log.Fatal("Unable to configure a valid storage engine, %w", err)
	}
	obj.SetStorage(storageEngine)

	return obj
}

func (s *DashNGoImpl) SetStorage(v storage.Storage) {
	s.storage = v
}

func ConfigureStorage(cfg *domain.GDGAppConfiguration) (storage.Storage, error) {
	var (
		storageEngine storage.Storage
		err           error
	)

	// config
	storageType, appData := cfg.GetCloudConfiguration(cfg.GetDefaultGrafanaConfig().Storage)

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

func NewTestApiService(storageEngine storage.Storage, cfg *domain.GDGAppConfiguration) GrafanaService {
	if cfg == nil {
		cfg = config.InitGdgConfig(common.DefaultTestConfig)
	}
	ins := newInstance(cfg)
	ins.SetStorage(storageEngine)
	setupConfigData(cfg, ins)
	return ins
}

func NewDashNGoImpl(cfg *domain.GDGAppConfiguration) *DashNGoImpl {
	if instance == nil {
		instance = newInstance(cfg)
	}
	return instance
}
