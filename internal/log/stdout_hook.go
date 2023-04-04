package log

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
)

// StdErrLoggingHook writes all log messages to os.Stderr
type StdOutLoggingHook struct {
	Formatter log.Formatter
}

func (hook *StdOutLoggingHook) Fire(entry *log.Entry) error {
	message, err := hook.Formatter.Format(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to format log entry: %v", err)
		return err
	}

	_, err = os.Stdout.Write(message)
	return err
}

func (hook *StdOutLoggingHook) Levels() []log.Level {
	return []log.Level{
		log.InfoLevel,
		log.DebugLevel,
		log.TraceLevel,
	}
}
