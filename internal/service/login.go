package service

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"github.com/esnet/gdg/internal/api"
	"github.com/esnet/gdg/internal/config"
	"github.com/go-openapi/strfmt"
	"net/url"

	"github.com/go-openapi/runtime"
	"github.com/grafana/grafana-openapi-client-go/client"
	"log"
	"net/http"
)

// AuthenticationApi Contract definition
type AuthenticationApi interface {
	TokenApi
	ServiceAccountApi
	Login()
}

// Login Logs into grafana returning a legacyClient instance using Token or Basic Auth
func (s *DashNGoImpl) Login() {
	var err error
	u, err := url.Parse(s.grafanaConf.URL)
	if err != nil {
		log.Fatal("invalid Grafana URL")
	}
	var clientTransport *http.Transport
	httpTransportCfg := client.TransportConfig{
		Host:     u.Host,
		BasePath: "/api",
		Schemes:  []string{u.Scheme},
		//NumRetries: 3,
	}

	if config.Config().IgnoreSSL() {
		_, clientTransport = ignoreSSLErrors()
		httpTransportCfg.TLSConfig = clientTransport.TLSClientConfig
	}
	if s.grafanaConf.UserName != "" && s.grafanaConf.Password != "" {
		httpTransportCfg.BasicAuth = url.UserPassword(s.grafanaConf.UserName, s.grafanaConf.Password)
	}
	if s.grafanaConf.APIToken != "" {
		httpTransportCfg.APIKey = s.grafanaConf.APIToken
	}
	if s.grafanaConf.OrganizationId != 0 {
		httpTransportCfg.OrgID = s.grafanaConf.OrganizationId
	}
	s.client = client.NewHTTPClientWithConfig(strfmt.Default, &httpTransportCfg)

	userInfo, err := s.GetUserInfo()
	//Sets state based on user permissions
	if err == nil {
		s.grafanaConf.SetAdmin(userInfo.IsGrafanaAdmin)
	}

	s.extended = api.NewExtendedApi()

}

// getGrafanaAdminAuth returns a runtime.ClientAuthInfoWriter that represents a Grafana Admin
func (s *DashNGoImpl) getGrafanaAdminAuth() runtime.ClientAuthInfoWriter {
	if !s.grafanaConf.IsAdminEnabled() || s.grafanaConf.UserName == "" {
		log.Fatal("Unable to get Grafana Admin Auth. ")
	}

	return s.getBasicAuth()
}

// getBasicAuth returns a valid user/password auth
func (s *DashNGoImpl) getBasicAuth() runtime.ClientAuthInfoWriter {
	basicAuth := runtime.ClientAuthInfoWriterFunc(func(req runtime.ClientRequest, registry strfmt.Registry) error {
		creds := fmt.Sprintf("%s:%s", s.grafanaConf.UserName, s.grafanaConf.Password)
		return req.SetHeaderParam("Authorization", fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(creds))))
	})
	return basicAuth
}

// getAuth returns token if present or basic auth
func (s *DashNGoImpl) getAuth() runtime.ClientAuthInfoWriter {
	if s.grafanaConf.APIToken != "" {
		return runtime.ClientAuthInfoWriterFunc(func(req runtime.ClientRequest, registry strfmt.Registry) error {
			return req.SetHeaderParam("Authorization", fmt.Sprintf("Bearer %s", s.grafanaConf.APIToken))
		})
	} else {
		return s.getBasicAuth()
	}
}

// ignoreSSLErrors when called replaces the default http legacyClient to ignore invalid SSL issues.
// only to be used for testing, highly discouraged in production.
func ignoreSSLErrors() (*http.Client, *http.Transport) {
	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	httpclient := &http.Client{Transport: customTransport}
	return httpclient, customTransport

}
