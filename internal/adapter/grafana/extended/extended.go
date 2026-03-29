package extended

import (
	"crypto/tls"
	"net/http"

	"github.com/carlmjohnson/requests"
	"github.com/esnet/gdg/internal/config/config_domain"
	"github.com/esnet/gdg/internal/ports/outbound"
)

// Api provides API request building for Grafana with optional debug mode.
type Api struct {
	appCfg *config_domain.GDGAppConfiguration
	debug  bool
}

func NewExtendedApi(cfg *config_domain.GDGAppConfiguration) outbound.ExtendedApi {
	o := Api{
		appCfg: cfg,
		debug:  cfg.IsApiDebug(),
	}
	return &o
}

// getRequestBuilder returns a requests.Builder preconfigured with Grafana URL, auth, and optional TLS settings.
func (extended *Api) getRequestBuilder() *requests.Builder {
	req := requests.URL(extended.appCfg.GetDefaultGrafanaConfig().GetURL())
	if extended.appCfg.IgnoreSSL() {
		customTransport := http.DefaultTransport.(*http.Transport).Clone()
		customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // #nosec G402
		req = req.Transport(customTransport)
	}
	token := extended.appCfg.GetDefaultGrafanaConfig().GetAPIToken()

	if token != "" {
		req.Header("Authorization", "Bearer "+token)
	} else {
		password := extended.appCfg.GetDefaultGrafanaConfig().GetPassword()
		req.BasicAuth(extended.appCfg.GetDefaultGrafanaConfig().UserName, password)
	}

	return req
}
