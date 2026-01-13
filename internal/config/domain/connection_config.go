package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/esnet/gdg/pkg/plugins/secure/contract"
	"gopkg.in/yaml.v3"
)

const (
	DefaultOrganizationName = "Main Org."
	DefaultOrganizationId   = 1
)

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

// ConnectionFilters model wraps connection filters for grafana
type ConnectionFilters struct {
	NameExclusions  string   `yaml:"name_exclusions" mapstructure:"name_exclusions"`
	ConnectionTypes []string `yaml:"valid_types" mapstructure:"valid_types"`
}

// GrafanaConnection Default connection credentials
type GrafanaConnection map[string]string

func (r *RegexMatchesList) GetConnectionAuth(path string, encoder contract.CipherEncoder) (*GrafanaConnection, error) {
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
	return g["user"]
}

func (g GrafanaConnection) Password() string {
	return g["basicAuthPassword"]
}
