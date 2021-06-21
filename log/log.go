package log

import (
	"os"
	"regexp"

	log "github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
)

//Logrus Stdout/StdErr split
type LogWriter struct{}

var (
	errorLogLevels []string = []string{"error", "fatal", "warning", "panic"}
	levelRegex     *regexp.Regexp
)

//detectLogLevel extracts log level from log
func (w *LogWriter) detectLogLevel(p []byte) (level string) {
	matches := levelRegex.FindStringSubmatch(string(p))
	if len(matches) > 1 {
		level = matches[1]
	}
	return
}

//Write Sends logLevels matching errorLogLevels to stderr
func (w *LogWriter) Write(p []byte) (n int, err error) {
	level := w.detectLogLevel(p)
	if funk.Contains(errorLogLevels, level) {
		return os.Stderr.Write(p)
	}
	return os.Stdout.Write(p)
}

//InitializeAppLogger initialize logger, invoked from main
func InitializeAppLogger() {
	var err error
	levelRegex, err = regexp.Compile("level=([a-z]+)")
	if err != nil {
		log.WithError(err).Fatal("Cannot setup log level")
	}

	log.SetOutput(&LogWriter{})

}
