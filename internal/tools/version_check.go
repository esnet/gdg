package tools

import (
	"fmt"
	"log/slog"

	"golang.org/x/mod/semver"
)

const (
	VersionKey = "Version"
)

type VersionRange struct {
	MinVersion string
	MaxVersion string
}

// Validate ensures that all version number start with a v, in order to be parsed correctly
func (v VersionRange) Validate() bool {
	if v.MaxVersion != "" && v.MaxVersion[0] != 'v' {
		return false
	}
	if v.MinVersion != "" && v.MinVersion[0] != 'v' {
		return false
	}

	return true
}

// InRange returns true if the current grafana version in within all of the ranges
// specified.  Falls if it is not.
func InRange(ranges []VersionRange, api GetVersion) bool {
	versionCheck := api.GetServerInfo()
	currentVersion := fmt.Sprintf("v%s", versionCheck[VersionKey].(string))
	valid := true
	for _, entry := range ranges {
		if !entry.Validate() {
			slog.Info("range is not valid")
			valid = false
			break
		}

		// if currentVersion < minVersion || currentVersion > maxVersion
		if semver.Compare(currentVersion, entry.MinVersion) == -1 /* greater or equal */ ||
			semver.Compare(currentVersion, entry.MaxVersion) == 1 {
			slog.Info("Range is valid",
				slog.String("version", currentVersion),
				slog.String("min", entry.MinVersion), slog.String("max", entry.MaxVersion))
			valid = false
			break
		}
	}
	return valid
}

type GetVersion interface {
	GetServerInfo() map[string]any
}

func ValidateMinimumVersion(minVersion string, api GetVersion) bool {
	if minVersion[0] != 'v' {
		slog.Error("Version check failed, minVersion must start with a 'v'")
		return false

	}
	versionCheck := api.GetServerInfo()
	currentVersion := fmt.Sprintf("v%s", versionCheck[VersionKey].(string))

	return semver.Compare(currentVersion, minVersion) != -1
}
