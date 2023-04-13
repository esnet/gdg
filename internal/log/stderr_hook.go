package log

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
)

// StdErrLoggingHook writes all log messages to os.Stderr
type StdErrLoggingHook struct {
	Formatter log.Formatter
}

func (hook *StdErrLoggingHook) Fire(entry *log.Entry) error {
	message, err := hook.Formatter.Format(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to format log entry: %v", err)
		return err
	}

	_, err = os.Stderr.Write(message)
	return err
}

func (hook *StdErrLoggingHook) Levels() []log.Level {
	return []log.Level{
		log.PanicLevel,
		log.FatalLevel,
		log.ErrorLevel,
		log.WarnLevel,
	}
}
