package main

import (
	_ "embed"
	"github.com/esnet/gdg/cli"
	"github.com/esnet/gdg/cli/support"
	"log"
	"os"

	api "github.com/esnet/gdg/internal/service"
)

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

	err := cli.Execute("importer-example.yml", os.Args[1:], setGrafanaSvc())
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
}
