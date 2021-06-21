package main

import (
	"github.com/netsage-project/grafana-dashboard-manager/cmd"
	applogger "github.com/netsage-project/grafana-dashboard-manager/log"
)

func main() {

	cmd.Execute()

}

func init() {
	applogger.InitializeAppLogger()

}
