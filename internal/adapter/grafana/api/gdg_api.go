package api

import (
	"github.com/esnet/gdg/internal/config/config_domain"
	"github.com/esnet/gdg/internal/ports"
	"github.com/esnet/gdg/internal/ports/outbound"
)

type DashNGoImpl struct {
	extended    outbound.ExtendedApi
	gdgConfig   *config_domain.GDGAppConfiguration
	grafanaConf *config_domain.GrafanaConfig
	storage     outbound.Storage
	encoder     outbound.CipherEncoder
	resources   ports.Resources
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

func (s *DashNGoImpl) SetStorage(v outbound.Storage) {
	s.storage = v
}

func NewDashNGo(
	cfg *config_domain.GDGAppConfiguration,
	encoder outbound.CipherEncoder,
	disk outbound.Storage,
	extended outbound.ExtendedApi,
	resource ports.Resources) outbound.GrafanaService {
	obj := &DashNGoImpl{
		gdgConfig: cfg,
		extended:  extended,
		resources: resource,
	}
	// Attach config
	obj.grafanaConf = cfg.GetDefaultGrafanaConfig()
	obj.gdgConfig = cfg
	obj.encoder = encoder
	obj.storage = disk

	return obj
}
