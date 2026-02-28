package cli_helper

import (
	"os"
	"strings"
)

// ParseConfigContextParams pre-parses os.Args to extract --config and --context flag values before Cobra processes
// them. This allows early loading of configuration and context so that services can be fully wired before CLI execution.
func ParseConfigContextParams() (configPath string, contextOverride string) {
	for i, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "--config=") {
			configPath = strings.TrimPrefix(arg, "--config=")
		} else if arg == "--config" || arg == "-c" {
			if i+1 < len(os.Args[1:]) {
				configPath = os.Args[i+2]
			}
		}
		if strings.HasPrefix(arg, "--context=") {
			contextOverride = strings.TrimPrefix(arg, "--context=")
		} else if arg == "--context" {
			if i+1 < len(os.Args[1:]) {
				contextOverride = os.Args[i+2]
			}
		}
	}
	return configPath, contextOverride
}
