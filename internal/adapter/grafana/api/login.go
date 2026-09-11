package api

import (
	"github.com/grafana/grafana-openapi-client-go/client"
	"github.com/grafana/grafana-openapi-client-go/models"
)

// NewClientOpts is a functional option applied to a TransportConfig when
// constructing a Grafana HTTP client. Defined here so it is visible across
// the package; the actual client-building logic lives on baseService.
type NewClientOpts func(transportConfig *client.TransportConfig)

// Login sets admin flag and provisions the Extended API for calls unsupported
// by the OpenAPI spec.
func (s *DashNGoImpl) Login() {
	if s.gdgConfig.PluginConfig.CipherEnabled() {
		s.grafanaConf.UpdateSecureModel(s.encoder.DecodeValue, s.lookupSvc)
	}

	if s.lookupResolver != nil {
		s.grafanaConf.ResolveLookups(s.lookupResolver.Resolve, s.lookupSvc)
	}

	// Will only succeed for BasicAuth
	if s.grafanaConf.IsBasicAuth() {
		var userInfo *models.UserProfileDTO
		var err error
		userInfo, err = s.GetUserInfo()
		if err == nil {
			s.grafanaConf.SetGrafanaAdmin(userInfo.IsGrafanaAdmin)
		}
	}
}
