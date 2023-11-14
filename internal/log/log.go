package log

import (
	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
	"log"
	"log/slog"
	"os"
	"time"
)

// InitializeAppLogger initialize logger, invoked from main
func InitializeAppLogger(stdout *os.File, stderr *os.File, debug bool) {
	errStream := stderr
	outStream := stdout
	level := slog.LevelInfo
	showSource := false
	if debug {
		level = slog.LevelDebug
		showSource = true
	}

	opts := &tint.Options{
		Level:      level,
		TimeFormat: time.DateTime,
		AddSource:  showSource,
		NoColor:    !isatty.IsTerminal(outStream.Fd())}

	//Splits the logging between stdout/stderr as appropriate
	myHandler := NewContextHandler(slog.Default().Handler(), outStream, errStream, opts)
	customSplitStreamLogger := slog.New(myHandler)
	slog.SetDefault(customSplitStreamLogger)
	log.SetOutput(os.Stderr)

}
