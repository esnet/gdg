package api

import (
	"fmt"

	"github.com/grafana-tools/sdk"
)

//TokenLogin: given a URL and token return the client
func TokenLogin(url, token string) *sdk.Client {
	return sdk.NewClient(url, token, sdk.DefaultHTTPClient)
}

//AuthLogin: Login using a username/password
func AuthLogin(url, username, password string) *sdk.Client {
	basicAuth := fmt.Sprintf("%s:%s", username, password)
	return sdk.NewClient(url, basicAuth, sdk.DefaultHTTPClient)
}
