package tools

import (
	"log/slog"

	"github.com/Masterminds/semver/v3"
)

const (
	VersionKey = "Version"
)

type VersionRange struct {
	MinVersion string
	MaxVersion string
}

// Validate ensures the version strings in the range are parseable by semver.
func (v VersionRange) Validate() bool {
	if v.MinVersion != "" {
		if _, err := semver.NewVersion(v.MinVersion); err != nil {
			return false
		}
	}
	if v.MaxVersion != "" {
		if _, err := semver.NewVersion(v.MaxVersion); err != nil {
			return false
		}
	}
	return true
}

// InRange returns true if the current Grafana version falls within all of the
// supplied ranges (inclusive on both ends). Returns false if any range is
// invalid or the version lies outside any range.
func InRange(ranges []VersionRange, api GetVersion) bool {
	versionCheck := api.GetServerInfo()
	currentVersion := versionCheck[VersionKey].(string)

	current, err := semver.NewVersion(currentVersion)
	if err != nil {
		slog.Error("unable to parse current Grafana version", slog.String("version", currentVersion), slog.Any("err", err))
		return false
	}

	for _, entry := range ranges {
		if !entry.Validate() {
			slog.Info("version range is not valid", slog.String("min", entry.MinVersion), slog.String("max", entry.MaxVersion))
			return false
		}

		if entry.MinVersion != "" {
			min, _ := semver.NewVersion(entry.MinVersion)
			if current.LessThan(min) {
				slog.Warn("version below range minimum",
					slog.String("version", currentVersion),
					slog.String("min", entry.MinVersion))
				return false
			}
		}

		if entry.MaxVersion != "" {
			max, _ := semver.NewVersion(entry.MaxVersion)
			if current.GreaterThan(max) {
				slog.Warn("version above range maximum",
					slog.String("version", currentVersion),
					slog.String("max", entry.MaxVersion))
				return false
			}
		}
	}
	return true
}

// GetVersion is implemented by any service that can report the Grafana server info.
type GetVersion interface {
	GetServerInfo() map[string]any
}

// ValidateMinimumVersion returns true if the current Grafana server version is
// greater than or equal to minVersion. minVersion may be prefixed with "v" or
// given as a bare semver string (e.g. "13.0.0" or "v13.0.0").
func ValidateMinimumVersion(minVersion string, api GetVersion) bool {
	minVersionChecker, err := semver.NewVersion(minVersion)
	if err != nil {
		slog.Error("ValidateMinimumVersion: invalid minVersion", slog.String("minVersion", minVersion), slog.Any("err", err))
		return false
	}

	versionCheck := api.GetServerInfo()
	currentVersion := versionCheck[VersionKey].(string)
	current, err := semver.NewVersion(currentVersion)
	if err != nil {
		slog.Error("ValidateMinimumVersion: unable to parse current Grafana version",
			slog.String("version", currentVersion), slog.Any("err", err))
		return false
	}

	return current.GreaterThanEqual(minVersionChecker)
}
