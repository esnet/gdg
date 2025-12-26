package api

import (
	"crypto/tls"
	"net/http"

	"github.com/esnet/gdg/internal/config/domain"

	"github.com/carlmjohnson/requests"
)

// ExtendedApi provides API request building for Grafana with optional debug mode.
type ExtendedApi struct {
	appCfg *domain.GDGAppConfiguration
	debug  bool
}

func NewExtendedApi(cfg *domain.GDGAppConfiguration) *ExtendedApi {
	o := ExtendedApi{
		appCfg: cfg,
		debug:  cfg.IsApiDebug(),
	}
	return &o
}

// getRequestBuilder returns a requests.Builder preconfigured with Grafana URL, auth, and optional TLS settings.
func (extended *ExtendedApi) getRequestBuilder() *requests.Builder {
	req := requests.URL(extended.appCfg.GetDefaultGrafanaConfig().GetURL())
	if extended.appCfg.IgnoreSSL() {
		customTransport := http.DefaultTransport.(*http.Transport).Clone()
		customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // #nosec G402
		req = req.Transport(customTransport)
	}

	if extended.appCfg.GetDefaultGrafanaConfig().GetAPIToken() != "" {
		req.Header("Authorization", "Bearer "+extended.appCfg.GetDefaultGrafanaConfig().GetAPIToken())
	} else {
		req.BasicAuth(extended.appCfg.GetDefaultGrafanaConfig().UserName, extended.appCfg.GetDefaultGrafanaConfig().GetPassword())
	}

	return req
}
