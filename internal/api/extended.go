package api

import (
	"crypto/tls"
	"net/http"

	"github.com/carlmjohnson/requests"
	"github.com/esnet/gdg/internal/config"
)

// Most of these methods are here due to limitations in existing libraries being used.
type ExtendedApi struct {
	grafanaCfg *config.GrafanaConfig
	debug      bool
}

func NewExtendedApi() *ExtendedApi {
	cfg := config.Config()
	o := ExtendedApi{
		grafanaCfg: cfg.GetDefaultGrafanaConfig(),
		debug:      cfg.IsApiDebug(),
	}
	return &o
}

func (extended *ExtendedApi) getRequestBuilder() *requests.Builder {
	req := requests.URL(extended.grafanaCfg.URL)
	if config.Config().IgnoreSSL() {
		customTransport := http.DefaultTransport.(*http.Transport).Clone()
		customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // #nosec G402
		req = req.Transport(customTransport)
	}

	if extended.grafanaCfg.APIToken != "" {
		req.Header("Authorization", "Bearer "+extended.grafanaCfg.APIToken)
	} else {
		req.BasicAuth(extended.grafanaCfg.UserName, extended.grafanaCfg.Password)
	}

	return req
}
