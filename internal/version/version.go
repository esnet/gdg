package version

import (
	"fmt"
	"log/slog"
	"runtime"
)

// GitCommit returns the git commit that was compiled. This will be filled in by the compiler.
var GitCommit string

// Version returns current version.  Set to release CICD
var Version = "DEVEL"

// BuildDate returns the date the binary was built
var BuildDate = ""

// GoVersion returns the version of the go runtime used to compile the binary
var GoVersion = runtime.Version()

// OsArch returns the os and arch used to build the binary
var OsArch = fmt.Sprintf("%s %s", runtime.GOOS, runtime.GOARCH)

func PrintVersionInfo() {
	slog.Info(fmt.Sprintf("Build Date: %s", BuildDate))
	slog.Info(fmt.Sprintf("Git Commit: %s", GitCommit))
	slog.Info(fmt.Sprintf("Version: %s", Version))
	slog.Info(fmt.Sprintf("Go Version: %s", GoVersion))
	slog.Info(fmt.Sprintf("OS / Arch: %s", OsArch))
}
