package api

import (
	"github.com/carlmjohnson/requests"
	"github.com/esnet/gdg/internal/apphelpers"
	"github.com/esnet/gdg/internal/config"
)

//Most of these methods are here due to limitations in existing libraries being used.

type ExtendedApi struct {
	grafanaCfg *config.GrafanaConfig
}

func NewExtendedApi() *ExtendedApi {
	cfg := apphelpers.GetCtxDefaultGrafanaConfig()
	o := ExtendedApi{
		grafanaCfg: cfg,
	}

	return &o
}

func (s *ExtendedApi) getRequestBuilder() *requests.Builder {
	req := requests.URL(s.grafanaCfg.URL)

	if s.grafanaCfg.UserName != "" && s.grafanaCfg.Password != "" {
		req.BasicAuth(s.grafanaCfg.UserName, s.grafanaCfg.Password)
	} else {
		req.Header("Authorization", "Bearer "+s.grafanaCfg.APIToken)
	}

	return req
}
