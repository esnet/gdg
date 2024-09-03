package service

import (
	"crypto/tls"
	"log"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/esnet/gdg/internal/api"
	"github.com/esnet/gdg/internal/config"
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

	s.extended = api.NewExtendedApi()
}

func ignoreSSL(transportConfig *client.TransportConfig) {
	_, clientTransport := ignoreSSLErrors()
	transportConfig.TLSConfig = clientTransport.TLSClientConfig
}

type NewClientOpts func(transportConfig *client.TransportConfig)

func (s *DashNGoImpl) getNewClient(opts ...NewClientOpts) (*client.GrafanaHTTPAPI, *client.TransportConfig) {
	var err error
	u, err := url.Parse(s.grafanaConf.URL)
	if err != nil {
		log.Fatal("invalid Grafana URL", s.grafanaConf.URL)
	}
	path, err := url.JoinPath(u.Path, "api")
	if err != nil {
		log.Fatal("invalid Grafana URL Path")
	}

	httpConfig := &client.TransportConfig{
		Host:         u.Host,
		BasePath:     path,
		Schemes:      []string{u.Scheme},
		NumRetries:   config.Config().GetGDGConfig().GetAppGlobals().RetryCount,
		RetryTimeout: config.Config().GetGDGConfig().GetAppGlobals().GetRetryTimeout(),
		Debug:        s.apiDebug,
	}

	if s.grafanaConf.IsBasicAuth() {
		if s.grafanaConf.OrganizationName != "" {
			orgId, err := api.NewExtendedApi().GetConfiguredOrgId(s.grafanaConf.OrganizationName)
			if err != nil {
				slog.Error("unable to determine org ID, falling back", slog.Any("err", err))
				orgId = 1
			}
			opts = append(opts, func(clientCfg *client.TransportConfig) {
				clientCfg.OrgID = orgId
			})
		} else {
			opts = append(opts, func(clientCfg *client.TransportConfig) {
				clientCfg.OrgID = config.DefaultOrganizationId
			})
		}
	}
	for _, opt := range opts {
		if opt != nil {
			opt(httpConfig)
		}
	}
	if config.Config().IgnoreSSL() {
		ignoreSSL(httpConfig)
	}

	return client.NewHTTPClientWithConfig(strfmt.Default, httpConfig), httpConfig
}

// GetClient Returns a new defaultClient given token precedence over Basic Auth
func (s *DashNGoImpl) GetClient() *client.GrafanaHTTPAPI {
	if s.grafanaConf.APIToken != "" {
		grafanaClient, _ := s.getNewClient(func(clientCfg *client.TransportConfig) {
			clientCfg.APIKey = s.grafanaConf.APIToken
			clientCfg.Debug = s.apiDebug
		})
		return grafanaClient
	} else {
		return s.GetBasicAuthClient()
	}
}

// GetAdminClient Returns the admin defaultClient if one is configured
func (s *DashNGoImpl) GetAdminClient() *client.GrafanaHTTPAPI {
	if !s.grafanaConf.IsGrafanaAdmin() || s.grafanaConf.UserName == "" {
		log.Fatal("Unable to get Grafana Admin SecureData. ")
	}
	return s.GetBasicAuthClient()
}

// GetBasicAuthClient returns a basic auth grafana API Client
func (s *DashNGoImpl) GetBasicAuthClient() *client.GrafanaHTTPAPI {
	grafanaClient, _ := s.getNewClient(func(clientCfg *client.TransportConfig) {
		clientCfg.BasicAuth = url.UserPassword(s.grafanaConf.UserName, s.grafanaConf.Password)
		clientCfg.Debug = s.apiDebug
	})
	return grafanaClient
}

// ignoreSSLErrors when called replaces the default http legacyClient to ignore invalid SSL issues.
// only to be used for testing, highly discouraged in production.
func ignoreSSLErrors() (*http.Client, *http.Transport) {
	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	httpclient := &http.Client{Transport: customTransport}
	return httpclient, customTransport
}
