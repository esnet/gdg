package log

import (
	log "github.com/sirupsen/logrus"
	"io"
)

// InitializeAppLogger initialize logger, invoked from main
func InitializeAppLogger() {
	log.SetOutput(io.Discard)
	log.AddHook(&StdOutLoggingHook{&log.TextFormatter{ForceColors: true}})
	log.AddHook(&StdErrLoggingHook{&log.TextFormatter{ForceColors: true}})

}
