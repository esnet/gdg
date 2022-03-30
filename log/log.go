package log

import (
	log "github.com/sirupsen/logrus"
	"io/ioutil"
)

//InitializeAppLogger initialize logger, invoked from main
func InitializeAppLogger() {
	log.SetOutput(ioutil.Discard)
	log.AddHook(&StdOutLoggingHook{&log.TextFormatter{ForceColors: true}})
	log.AddHook(&StdErrLoggingHook{&log.TextFormatter{ForceColors: true}})
}
