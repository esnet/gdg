package api

import (
	"crypto/tls"
	"fmt"
	"net/url"

	"github.com/esnet/gdg/config"
	"github.com/grafana-tools/sdk"
	gclient "github.com/grafana/grafana-api-golang-client"
	log "github.com/sirupsen/logrus"
	"net/http"
)

// LegacyLogin Logs into grafana returning a legacyClient instance using Token or Basic Auth
func (s *DashNGoImpl) Login() {
	s.legacyLogin()
	s.newLogin()

}

func (s *DashNGoImpl) newLogin() {
	var cfg gclient.Config

	httpClient := http.DefaultClient
	if config.Config().IgnoreSSL() {
		httpClient = ignoreSSLErrors()
	}
	cfg = gclient.Config{
		APIKey:    s.grafanaConf.APIToken,
		BasicAuth: url.UserPassword(s.grafanaConf.UserName, s.grafanaConf.Password),
		Client:    httpClient,
	}

	client, err := gclient.New(s.grafanaConf.URL, cfg)
	if err != nil {
		log.Fatal("unable to get Grafana API Client")
	}
	s.client = client
}

func (s *DashNGoImpl) legacyLogin() {
	//If ignoreSSL create custom http legacyClient
	if config.Config().IgnoreSSL() {
		customClient := ignoreSSLErrors()
		sdk.DefaultHTTPClient = customClient
	}
	if s.grafanaConf.APIToken != "" {
		client := s.tokenLogin()
		s.legacyClient = client
	} else if s.grafanaConf.UserName != "" && s.grafanaConf.Password != "" {
		s.legacyClient = s.authLogin()
	} else {
		panic("Invalid auth configuration.  Either Token or password based credentials required")
	}

}

func (s *DashNGoImpl) AdminLogin() {
	s.legacyAdminLogin()
	s.newAdminLogin()

}

func (s *DashNGoImpl) newAdminLogin() {
	if s.grafanaConf.UserName != "" && s.grafanaConf.Password != "" {
		s.grafanaConf.AdminEnabled = true
		s.adminClient = s.client
	} else {
		s.grafanaConf.AdminEnabled = false
		s.legacyAdminClient = nil
	}
}

func (s *DashNGoImpl) legacyAdminLogin() {
	if s.grafanaConf.UserName != "" && s.grafanaConf.Password != "" {
		s.grafanaConf.AdminEnabled = true
		s.legacyAdminClient = s.authLogin()
	} else {
		s.grafanaConf.AdminEnabled = false
		s.legacyAdminClient = nil
	}
}

//ignoreSSLErrors when called replaces the default http legacyClient to ignore invalid SSL issues.
//only to be used for testing, highly discouraged in production.
func ignoreSSLErrors() *http.Client {
	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	httpclient := &http.Client{Transport: customTransport}
	return httpclient

}

//tokenLogin: given a URL and token return the legacyClient
func (s *DashNGoImpl) tokenLogin() *sdk.Client {
	client, err := sdk.NewClient(s.grafanaConf.URL, s.grafanaConf.APIToken, sdk.DefaultHTTPClient)
	if err != nil {
		log.Fatal("failed to get a valid legacyClient using token auth")
	}

	return client
}

//AuthLogin: Login using a username/password
func (s *DashNGoImpl) authLogin() *sdk.Client {
	basicAuth := fmt.Sprintf("%s:%s", s.grafanaConf.UserName, s.grafanaConf.Password)
	client, err := sdk.NewClient(s.grafanaConf.URL, basicAuth, sdk.DefaultHTTPClient)
	if err != nil {
		log.Fatal("failed to get a valid legacyClient using basic auth")
	}
	return client
}
