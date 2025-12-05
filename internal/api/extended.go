package api

import (
	"crypto/tls"
	"net/http"

	"github.com/esnet/gdg/internal/config/domain"

	"github.com/carlmjohnson/requests"
	"github.com/esnet/gdg/internal/config"
)

// ExtendedApi provides API request building for Grafana with optional debug mode.
type ExtendedApi struct {
	grafanaCfg *domain.GrafanaConfig
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

// getRequestBuilder returns a requests.Builder preconfigured with Grafana URL, auth, and optional TLS settings.
func (extended *ExtendedApi) getRequestBuilder() *requests.Builder {
	req := requests.URL(extended.grafanaCfg.GetURL())
	if config.Config().IgnoreSSL() {
		customTransport := http.DefaultTransport.(*http.Transport).Clone()
		customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // #nosec G402
		req = req.Transport(customTransport)
	}

	if extended.grafanaCfg.GetAPIToken() != "" {
		req.Header("Authorization", "Bearer "+extended.grafanaCfg.GetAPIToken())
	} else {
		req.BasicAuth(extended.grafanaCfg.UserName, extended.grafanaCfg.GetPassword())
	}

	return req
}
