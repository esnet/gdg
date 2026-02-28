package api

import (
	"github.com/esnet/gdg/internal/adapter/grafana/extended"
	"github.com/esnet/gdg/internal/config/config_domain"
	"github.com/esnet/gdg/internal/ports"
)

type DashNGoImpl struct {
	extended    *extended.ExtendedApi
	gdgConfig   *config_domain.GDGAppConfiguration
	grafanaConf *config_domain.GrafanaConfig
	storage     ports.Storage
	encoder     ports.CipherEncoder
}

func (s *DashNGoImpl) GetGlobals() *config_domain.AppGlobals {
	if s.gdgConfig.Global == nil {
		s.gdgConfig.Global = &config_domain.AppGlobals{}
	}
	return s.gdgConfig.Global
}

func (s *DashNGoImpl) GetGdgConfig() *config_domain.GDGAppConfiguration {
	return s.gdgConfig
}

func (s *DashNGoImpl) SetStorage(v ports.Storage) {
	s.storage = v
}

func NewDashNGo(cfg *config_domain.GDGAppConfiguration, encoder ports.CipherEncoder, disk ports.Storage) ports.GrafanaService {
	obj := &DashNGoImpl{
		gdgConfig: cfg,
	}
	// Attach config
	obj.grafanaConf = cfg.GetDefaultGrafanaConfig()
	obj.gdgConfig = cfg
	obj.encoder = encoder
	obj.storage = disk

	return obj
}
