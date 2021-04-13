package version

import (
	"fmt"
	"runtime"
)

// Version returns the main version number that is being run at the moment.
const Version = "0.1.1"

// GoVersion returns the version of the go runtime used to compile the binary
var GoVersion = runtime.Version()

// OsArch returns the os and arch used to build the binary
var OsArch = fmt.Sprintf("%s %s", runtime.GOOS, runtime.GOARCH)
