package main

import (
	"log"
	"os"

	"github.com/esnet/gdg/cli"
	"github.com/esnet/gdg/cli/support"

	api "github.com/esnet/gdg/internal/service"
)

var getGrafanaSvc = func() api.GrafanaService {
	return api.NewDashNGoImpl()
}

func main() {
	setGrafanaSvc := func() support.RootOption {
		return func(response *support.RootCommand) {
			response.GrafanaSvc = getGrafanaSvc
		}
	}

	err := cli.Execute(os.Args[1:], setGrafanaSvc())
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
}
