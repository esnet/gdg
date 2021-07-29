package api

import (
	"crypto/tls"
	"fmt"

	"net/http"

	"github.com/grafana-tools/sdk"
	"github.com/netsage-project/grafana-dashboard-manager/config"
	log "github.com/sirupsen/logrus"
)

//Login: Logs into grafana returning a client instance using Token or Basic Auth
func (s *DashNGoImpl) Login() *sdk.Client {

	//If ignoreSSL create custom http client
	if config.Config().IgnoreSSL() {
		ignoreSSLErrors()
	}
	if s.grafanaConf.APIToken != "" {
		return s.tokenLogin()
	} else if s.grafanaConf.UserName != "" && s.grafanaConf.Password != "" {
		return s.authLogin()
	}

	panic("Invalid auth configuration.  Either Token or password based credentials required")

}

func (s *DashNGoImpl) AdminLogin() *sdk.Client {
	if s.grafanaConf.UserName != "" && s.grafanaConf.Password != "" {
		s.grafanaConf.AdminEnabled = true
		return s.authLogin()
	} else {
		s.grafanaConf.AdminEnabled = false
		return nil
	}

}

//ignoreSSLErrors when called replaces the default http client to ignore invalid SSL issues.
//only to be used for testing, highly discouraged in production.
func ignoreSSLErrors() {
	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	httpclient := &http.Client{Transport: customTransport}
	sdk.DefaultHTTPClient = httpclient

}

//tokenLogin: given a URL and token return the client
func (s *DashNGoImpl) tokenLogin() *sdk.Client {
	client, err := sdk.NewClient(s.grafanaConf.URL, s.grafanaConf.APIToken, sdk.DefaultHTTPClient)
	if err != nil {
		log.Fatal("failed to get a valid client using token auth")
	}

	return client
}

//AuthLogin: Login using a username/password
func (s *DashNGoImpl) authLogin() *sdk.Client {
	basicAuth := fmt.Sprintf("%s:%s", s.grafanaConf.UserName, s.grafanaConf.Password)
	client, err := sdk.NewClient(s.grafanaConf.URL, basicAuth, sdk.DefaultHTTPClient)
	if err != nil {
		log.Fatal("failed to get a valid client using basic auth")
	}
	return client
}
