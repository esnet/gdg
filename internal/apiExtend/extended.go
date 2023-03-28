package apiExtend

import (
	"github.com/carlmjohnson/requests"
	"github.com/esnet/gdg/apphelpers"
	"github.com/esnet/gdg/config"
)

//Most of these methods are here due to limitations in existing libraries being used.

type ExtendedApi struct {
	grafanaCfg *config.GrafanaConfig
	req        *requests.Builder
}

func NewExtendedApi() *ExtendedApi {
	cfg := apphelpers.GetCtxDefaultGrafanaConfig()
	o := ExtendedApi{
		grafanaCfg: cfg,
	}

	req := requests.URL(cfg.URL)

	if cfg.UserName != "" && cfg.Password != "" {
		req.BasicAuth(cfg.UserName, cfg.Password)
	} else {
		req.Header("Authorization", "Bearer "+cfg.APIToken)
	}

	o.req = req

	return &o
}
