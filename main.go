package main

import (
	"github.com/netsage-project/gdg/cmd"
	applogger "github.com/netsage-project/gdg/log"
)

func main() {

	cmd.Execute()

}

func init() {
	applogger.InitializeAppLogger()

}
