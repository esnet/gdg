package service

import (
	"crypto/tls"
	"github.com/esnet/gdg/internal/api"
	"github.com/esnet/gdg/internal/config"
	gapi "github.com/esnet/grafana-swagger-api-golang"
	"github.com/go-openapi/runtime/client"
	"net/url"

	gclient "github.com/esnet/grafana-swagger-api-golang/goclient/client"
	"github.com/go-openapi/runtime"
	log "github.com/sirupsen/logrus"
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
	httpClient := &http.Client{}
	if config.Config().IgnoreSSL() {
		httpClient = ignoreSSLErrors()
	}

	runtimeClient := client.NewWithClient(u.Host, "/api", []string{u.Scheme}, httpClient)
	s.client = gclient.New(runtimeClient, nil)

	s.extended = api.NewExtendedApi()

}

func (s *DashNGoImpl) getAdminAuth() runtime.ClientAuthInfoWriter {
	if s.grafanaConf.UserName == "" {
		log.Warnf("Unable to get Admin Auth.  Basic Auth credentials, continuing with token")
		return s.getAuth()
	}
	return &gapi.BasicAuthenticator{
		Username: s.grafanaConf.UserName,
		Password: s.grafanaConf.Password,
	}

}

func (s *DashNGoImpl) getAuth() runtime.ClientAuthInfoWriter {
	if s.grafanaConf.APIToken != "" {
		return &gapi.APIKeyAuthenticator{
			APIKey: s.grafanaConf.APIToken,
		}

	} else {
		return &gapi.BasicAuthenticator{
			Username: s.grafanaConf.UserName,
			Password: s.grafanaConf.Password,
		}
	}
}

// ignoreSSLErrors when called replaces the default http legacyClient to ignore invalid SSL issues.
// only to be used for testing, highly discouraged in production.
func ignoreSSLErrors() *http.Client {
	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	httpclient := &http.Client{Transport: customTransport}
	return httpclient

}
