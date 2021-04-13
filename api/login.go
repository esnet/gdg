package api

import (
	"fmt"

	"github.com/esnet/grafana-dashboard-manager/config"
	"github.com/grafana-tools/sdk"
)

//Login: Logs into grafana returning a client instance using Token or Basic Auth
func Login(grafanaConf *config.GrafanaConfig) *sdk.Client {
	if grafanaConf.APIToken != "" {
		return tokenLogin(grafanaConf.URL, grafanaConf.APIToken)
	} else if grafanaConf.UserName != "" && grafanaConf.Password != "" {
		return authLogin(grafanaConf.URL, grafanaConf.UserName, grafanaConf.Password)
	}

	panic("Invalid auth configuration.  Either Token or password based credentials required")

}

//tokenLogin: given a URL and token return the client
func tokenLogin(url, token string) *sdk.Client {
	return sdk.NewClient(url, token, sdk.DefaultHTTPClient)
}

//AuthLogin: Login using a username/password
func authLogin(url, username, password string) *sdk.Client {
	basicAuth := fmt.Sprintf("%s:%s", username, password)
	return sdk.NewClient(url, basicAuth, sdk.DefaultHTTPClient)
}
