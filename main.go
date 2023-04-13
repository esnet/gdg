package main

import (
	_ "embed"
	applogger "github.com/esnet/gdg/internal/log"
	"sync"

	"github.com/esnet/gdg/cmd"
)

//go:embed config/importer-example.yml
var defaultConfig string

var doOnce sync.Once

func main() {
	cmd.DefaultConfig = defaultConfig
	cmd.Execute()
}

func init() {
	doOnce.Do(func() {
		applogger.InitializeAppLogger()
	})
}
