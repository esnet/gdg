package config_tooling

import (
	"encoding/json/v2"
	"errors"
	"log/slog"

	"github.com/esnet/gdg/internal/config/config_domain"
)

// IsExcluded determines if an item matches any exclusion rule from the provided list of matching rules.
func IsExcluded[T any](item T, rules []config_domain.MatchingRule) bool {
	data, err := json.Marshal(item)
	if err != nil {
		slog.Warn("Unable to serialize object, cannot validate")
		return true
	}

	// Since filters are always converted, only check we need should be this one.
	for _, field := range rules {
		match, fieldParseErr := field.IsValid(data)

		if errors.Is(fieldParseErr, config_domain.ErrMissingField) {
			continue
		}
		if fieldParseErr == nil && match {
			return match
		}
	}

	return false
}
