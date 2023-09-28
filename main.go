package main

import (
	_ "embed"
	"github.com/esnet/gdg/cmd"
	"github.com/esnet/gdg/cmd/support"
	applogger "github.com/esnet/gdg/internal/log"
	api "github.com/esnet/gdg/internal/service"
	log "github.com/sirupsen/logrus"
	"os"
)

//go:embed config/importer-example.yml
var defaultConfig string

var (
	getGrafanaSvc = func() api.GrafanaService {
		return api.NewApiService()
	}
)

func main() {
	applogger.InitializeAppLogger()

	setGrafanaSvc := func() support.RootOption {
		return func(response *support.RootCommand) {
			response.GrafanaSvc = getGrafanaSvc
		}
	}

	err := cmd.Execute(defaultConfig, os.Args[1:], setGrafanaSvc())
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
}
