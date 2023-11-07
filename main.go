package main

import (
	_ "embed"
	"github.com/esnet/gdg/cmd"
	"github.com/esnet/gdg/cmd/support"
	"log"
	"os"

	api "github.com/esnet/gdg/internal/service"
)

//go:embed config/importer-example.yml
var defaultConfig string

var (
	getGrafanaSvc = func() api.GrafanaService {
		return api.NewApiService()
	}
)

func main() {
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
