package config_domain

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"

	"github.com/esnet/gdg/internal/ports/outbound"
	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v3"
)

const (
	DefaultOrganizationName = "Main Org."
	DefaultOrganizationId   = 1
	connectionUser          = "user"
	connectionPassword      = "basicAuthPassword"
)

var ErrMissingField = errors.New("no valid field found")

type CredentialRule struct {
	RegexMatchesList
	Auth *GrafanaConnection `mapstructure:"auth" yaml:"auth,omitempty"`
}

// MatchingRule defines a single matching rule for Grafana Connections
type MatchingRule struct {
	Field     string `yaml:"field,omitempty" mapstructure:"field,omitempty"`
	Regex     string `yaml:"regex,omitempty" mapstructure:"regex,omitempty"`
	Inclusive bool   `yaml:"inclusive,omitempty" mapstructure:"inclusive,omitempty"`
}

// IsValid evaluates if the provided data satisfies the matching rule based on field existence, regex, and inclusion criteria.
func (mr MatchingRule) IsValid(data []byte) (bool, error) {
	fieldParse := gjson.GetBytes(data, mr.Field)
	if !fieldParse.Exists() || mr.Regex == "" {
		return false, ErrMissingField
	}

	matchFunc := func(val string, matcher *regexp.Regexp) (bool, bool) {
		matchFound := false
		match := matcher.Match([]byte(val))
		// If inclusive, then the boolean is flipped
		if match {
			matchFound = true
		}
		if mr.Inclusive {
			match = !match
		}
		if match {
			return match, matchFound
		}
		return false, matchFound
	}

	p, err := regexp.Compile(mr.Regex)
	if err != nil {
		slog.Warn("Invalid regex for filter rule", "field", mr.Field)
		return true, nil
	}
	if fieldParse.IsArray() {
		for _, item := range fieldParse.Array() {
			fieldValue := item.String()
			match, found := matchFunc(fieldValue, p)
			if found {
				return match, nil
			}
		}
		// No element matched the regex. For an inclusive rule this means the
		// item does not satisfy the keep-condition, so it should be excluded.
		if mr.Inclusive {
			return true, nil
		}
		return false, nil
	}

	fieldValue := fieldParse.String()
	match, _ := matchFunc(fieldValue, p)
	return match, nil
}

// ConnectionFilters model wraps connection filters for grafana
type ConnectionFilters struct {
	NameExclusions  string   `yaml:"name_exclusions" mapstructure:"name_exclusions"`
	ConnectionTypes []string `yaml:"valid_types" mapstructure:"valid_types"`
}

// GrafanaConnection Default connection credentials
type GrafanaConnection map[string]string

func (r *RegexMatchesList) GetConnectionAuth(path string, encoder outbound.CipherEncoder) (*GrafanaConnection, error) {
	if r.result != nil {
		return r.result, nil
	}
	if r.SecureData == "" {
		return nil, fmt.Errorf("no valid auth can be found for the given path %s", path)
	}

	secretLocation := filepath.Join(path, r.SecureData)
	result := new(GrafanaConnection)
	raw, err := os.ReadFile(secretLocation) // #nosec G304
	if err != nil {
		msg := "unable to read secrets at location"
		slog.Error(msg, slog.String("file", secretLocation))
		return nil, errors.New(msg)
	}
	ext := filepath.Ext(secretLocation)
	switch ext {
	case ".yml", ".yaml":
		err = yaml.Unmarshal(raw, result)
		if err != nil {
			msg := "unable to read JSON secrets"
			slog.Error(msg, slog.Any("err", err), slog.String("file", secretLocation))
			return nil, errors.New(msg)
		}
	case ".json":
		err = json.Unmarshal(raw, result)
		if err != nil {
			msg := "unable to read JSON secrets"
			slog.Error(msg, slog.Any("err", err), slog.String("file", secretLocation))
			return nil, errors.New(msg)
		}
	default:
		return nil, fmt.Errorf("invalid file extension %s", ext)
	}

	for key, value := range *result {
		if encoder != nil {
			newVal, decodeErr := encoder.DecodeValue(value)
			if decodeErr == nil {
				(*result)[key] = newVal
			} else {
				slog.Debug("error decoding value for key",
					slog.String("key", key),
					slog.String("file", secretLocation),
					slog.Any("err", decodeErr))
			}
		}
	}

	r.result = result
	return r.result, nil
}

// CredentialRule model wraps regex and auth for grafana

func (g GrafanaConnection) User() string {
	return g[connectionUser]
}

func (g GrafanaConnection) Password() string {
	return g[connectionPassword]
}
