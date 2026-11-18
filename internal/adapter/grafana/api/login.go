package api

import (
	"github.com/grafana/grafana-openapi-client-go/models"
)

// Login sets admin flag and provisions the Extended API for calls unsupported
// by the OpenAPI spec.
func (s *DashNGoImpl) Login() {
	if s.GdgConfig.PluginConfig.CipherEnabled() {
		s.GrafanaConf.UpdateSecureModel(s.Encoder.DecodeValue, s.LookupSvc)
	}

	if s.LookupResolver != nil {
		s.GrafanaConf.ResolveLookups(s.LookupResolver.Resolve, s.LookupSvc)
	}

	// Will only succeed for BasicAuth
	if s.GrafanaConf.IsBasicAuth() {
		var userInfo *models.UserProfileDTO
		var err error
		userInfo, err = s.GetUserInfo()
		if err == nil {
			s.GrafanaConf.SetGrafanaAdmin(userInfo.IsGrafanaAdmin)
		}
	}
}
