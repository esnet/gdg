package service

import (
	"crypto/tls"
	"log"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/esnet/gdg/internal/config/domain"

	"github.com/esnet/gdg/internal/api"
	"github.com/go-openapi/strfmt"
	"github.com/grafana/grafana-openapi-client-go/models"

	"github.com/grafana/grafana-openapi-client-go/client"
)

// Login sets admin flag and provisions the Extended API for calls unsupported by the OpenAPI spec.
func (s *DashNGoImpl) Login() {
	var err error
	// Will only succeed for BasicAuth
	if s.grafanaConf.IsBasicAuth() {
		var userInfo *models.UserProfileDTO
		userInfo, err = s.GetUserInfo()
		// Sets state based on user permissions
		if err == nil {
			s.grafanaConf.SetGrafanaAdmin(userInfo.IsGrafanaAdmin)
		}
	}

	s.extended = api.NewExtendedApi(s.gdgConfig)
}

func ignoreSSL(transportConfig *client.TransportConfig) {
	_, clientTransport := ignoreSSLErrors()
	transportConfig.TLSConfig = clientTransport.TLSClientConfig
}

type NewClientOpts func(transportConfig *client.TransportConfig)

func GetOrgNameClientOpts(cfg *domain.GDGAppConfiguration) NewClientOpts {
	orgName := cfg.GetDefaultGrafanaConfig().OrganizationName
	if orgName != "" {
		return func(transportConfig *client.TransportConfig) {
			orgId, err := api.NewExtendedApi(cfg).GetConfiguredOrgId(orgName)
			if err != nil {
				slog.Error("unable to determine org ID, falling back", slog.Any("err", err))
				orgId = 1
			}

			transportConfig.OrgID = orgId
		}
	}

	return func(clientCfg *client.TransportConfig) {
		clientCfg.OrgID = domain.DefaultOrganizationId
	}
}

func (s *DashNGoImpl) getNewClient(opts ...NewClientOpts) (*client.GrafanaHTTPAPI, *client.TransportConfig) {
	var err error
	u, err := url.Parse(s.grafanaConf.GetURL())
	if err != nil {
		log.Fatal("invalid Grafana URL", s.grafanaConf.GetURL())
	}
	path, err := url.JoinPath(u.Path, "api")
	if err != nil {
		log.Fatal("invalid Grafana URL Path")
	}

	httpConfig := &client.TransportConfig{
		Host:         u.Host,
		BasePath:     path,
		Schemes:      []string{u.Scheme},
		NumRetries:   s.gdgConfig.GetAppGlobals().RetryCount,
		RetryTimeout: s.gdgConfig.GetAppGlobals().GetRetryTimeout(),
		Debug:        s.GetGlobals().ApiDebug,
	}

	// If more than one opts is passed, depend on the caller to setup his required configuration
	if s.grafanaConf.IsBasicAuth() && len(opts) == 1 {
		opts = append(opts, GetOrgNameClientOpts(s.gdgConfig))
	}
	for _, opt := range opts {
		if opt != nil {
			opt(httpConfig)
		}
	}
	if s.gdgConfig.IgnoreSSL() {
		ignoreSSL(httpConfig)
	}

	return client.NewHTTPClientWithConfig(strfmt.Default, httpConfig), httpConfig
}

// GetClient Returns a new defaultClient given token precedence over Basic Auth
func (s *DashNGoImpl) GetClient() *client.GrafanaHTTPAPI {
	if s.grafanaConf.GetAPIToken() != "" {
		token := s.grafanaConf.GetAPIToken()
		newToken, err := s.encoder.DecodeValue(token)
		if err != nil {
			slog.Warn("unable to decode token", slog.Any("err", err))
		} else {
			token = newToken
		}
		grafanaClient, _ := s.getNewClient(func(clientCfg *client.TransportConfig) {
			clientCfg.APIKey = token
			clientCfg.Debug = s.GetGlobals().ApiDebug
		})
		return grafanaClient
	}

	return s.GetBasicAuthClient()
}

func (s *DashNGoImpl) GetBasicClientWithOpts(opts ...NewClientOpts) *client.GrafanaHTTPAPI {
	allOpts := s.getDefaultBasicOpts()
	allOpts = append(allOpts, opts...)
	grafanaClient, _ := s.getNewClient(allOpts...)
	return grafanaClient
}

// GetAdminClient Returns the admin defaultClient if one is configured
func (s *DashNGoImpl) GetAdminClient() *client.GrafanaHTTPAPI {
	if !s.grafanaConf.IsGrafanaAdmin() || s.grafanaConf.UserName == "" {
		log.Fatal("Unable to get Grafana Admin SecureData. ")
	}
	return s.GetBasicClientWithOpts()
}

func (s *DashNGoImpl) getDefaultBasicOpts() []NewClientOpts {
	pass := s.grafanaConf.GetPassword()
	newPass, err := s.encoder.DecodeValue(pass)
	if err != nil {
		slog.Warn("Unable to decode password, falling back on string value", slog.String("password", pass))
	} else {
		pass = newPass
	}

	return []NewClientOpts{func(clientCfg *client.TransportConfig) {
		clientCfg.BasicAuth = url.UserPassword(s.grafanaConf.UserName, pass)
		clientCfg.Debug = s.GetGlobals().ApiDebug
	}}
}

// GetBasicAuthClient returns a basic auth grafana API Client
func (s *DashNGoImpl) GetBasicAuthClient() *client.GrafanaHTTPAPI {
	return s.GetBasicClientWithOpts()
}

// ignoreSSLErrors when called replaces the default http legacyClient to ignore invalid SSL issues.
// only to be used for testing, highly discouraged in production.
func ignoreSSLErrors() (*http.Client, *http.Transport) {
	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // #nosec G402
	httpclient := &http.Client{Transport: customTransport}
	return httpclient, customTransport
}
