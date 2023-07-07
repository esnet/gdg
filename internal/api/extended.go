package api

import (
	"crypto/tls"
	"github.com/carlmjohnson/requests"
	"github.com/esnet/gdg/internal/config"
	"net/http"
)

//Most of these methods are here due to limitations in existing libraries being used.

type ExtendedApi struct {
	grafanaCfg *config.GrafanaConfig
}

func NewExtendedApi() *ExtendedApi {
	cfg := config.Config().GetDefaultGrafanaConfig()
	o := ExtendedApi{
		grafanaCfg: cfg,
	}

	return &o
}

func (extended *ExtendedApi) getRequestBuilder() *requests.Builder {

	req := requests.URL(extended.grafanaCfg.URL)
	if config.Config().IgnoreSSL() {
		customTransport := http.DefaultTransport.(*http.Transport).Clone()
		customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		req = req.Transport(customTransport)
	}

	if extended.grafanaCfg.APIToken != "" {
		req.Header("Authorization", "Bearer "+extended.grafanaCfg.APIToken)
	} else {
		req.BasicAuth(extended.grafanaCfg.UserName, extended.grafanaCfg.Password)
	}

	return req
}
