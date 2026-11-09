package extended

import (
	"crypto/tls"
	"net/http"

	"github.com/carlmjohnson/requests"
	"github.com/esnet/gdg/internal/config/config_domain"
	"github.com/esnet/gdg/internal/ports/outbound"
)

// Api provides API request building for Grafana with custom HTTP client.
type Api struct {
	appCfg     *config_domain.GDGAppConfiguration
	httpClient *http.Client
}

func NewExtendedApi(cfg *config_domain.GDGAppConfiguration) outbound.ExtendedApi {
	c := *http.DefaultClient
	if cfg != nil {
		if cfg.IgnoreSSL() {
			t := http.DefaultTransport.(*http.Transport).Clone()
			t.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // #nosec G402
			c.Transport = t
		}

		if cfg.TransportWrapper != nil {
			c.Transport = cfg.TransportWrapper(c.Transport)
		}
	}

	o := Api{
		appCfg:     cfg,
		httpClient: &c,
	}
	return &o
}

// getRequestBuilder returns a requests.Builder preconfigured with Grafana URL, auth, and optional TLS settings.
func (extended *Api) getRequestBuilder() *requests.Builder {
	req := requests.URL(extended.appCfg.GetDefaultGrafanaConfig().GetURL())
	req = req.Client(extended.httpClient)

	token := extended.appCfg.GetDefaultGrafanaConfig().GetAPIToken()

	if token != "" {
		req.Header("Authorization", "Bearer "+token)
	} else {
		password := extended.appCfg.GetDefaultGrafanaConfig().GetPassword()
		req.BasicAuth(extended.appCfg.GetDefaultGrafanaConfig().UserName, password)
	}

	return req
}
