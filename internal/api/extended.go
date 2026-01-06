package api

import (
	"crypto/tls"
	"log/slog"
	"net/http"

	"github.com/esnet/gdg/internal/config/domain"
	"github.com/esnet/gdg/pkg/plugins/secure"
	"github.com/esnet/gdg/pkg/plugins/secure/contract"

	"github.com/carlmjohnson/requests"
)

// ExtendedApi provides API request building for Grafana with optional debug mode.
type ExtendedApi struct {
	appCfg  *domain.GDGAppConfiguration
	debug   bool
	encoder contract.CipherEncoder
}

func NewExtendedApi(cfg *domain.GDGAppConfiguration) *ExtendedApi {
	o := ExtendedApi{
		appCfg: cfg,
		debug:  cfg.IsApiDebug(),
	}
	if !cfg.PluginConfig.Disabled && cfg.PluginConfig.CipherPlugin != nil {
		o.encoder = secure.NewPluginCipherEncoder(cfg.PluginConfig.CipherPlugin, cfg.SecureConfig)
	} else {
		o.encoder = secure.NoOpEncoder{}
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
	token := extended.appCfg.GetDefaultGrafanaConfig().GetAPIToken()

	if token != "" {
		decodedValue, err := extended.encoder.DecodeValue(token)
		if err != nil {
			slog.Warn("Unable to decode Token using cipher plugin, trying string value")
		} else {
			token = decodedValue
		}
		req.Header("Authorization", "Bearer "+token)
	} else {
		password := extended.appCfg.GetDefaultGrafanaConfig().GetPassword()
		decodedValue, err := extended.encoder.DecodeValue(password)
		if err != nil {
			slog.Warn("Unable to decode Token using cipher plugin, trying string value")
		} else {
			password = decodedValue
		}
		req.BasicAuth(extended.appCfg.GetDefaultGrafanaConfig().UserName, password)
	}

	return req
}
