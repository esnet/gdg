package api

import (
	"fmt"
	"log"

	"github.com/grafana-tools/sdk"
)

//Login: Logs into grafana returning a client instance using Token or Basic Auth
func (s *DashNGoImpl) Login() *sdk.Client {
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
