package main

import (
	_ "embed"
	"github.com/netsage-project/gdg/cmd"
	applogger "github.com/netsage-project/gdg/log"
)

//go:embed conf/importer-example.yml
var defaultConfig string

func main() {
	cmd.DefaultConfig = defaultConfig
	cmd.Execute()

}

func init() {
	applogger.InitializeAppLogger()

}
