package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
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

func (r RegexMatchesList) GetConnectionAuth(path string) (*GrafanaConnection, error) {
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
	err = json.Unmarshal(raw, result)
	if err != nil {
		msg := "unable to read JSON secrets"
		slog.Error(msg, slog.Any("err", err), slog.String("file", secretLocation))
		return nil, errors.New(msg)
	}

	return result, nil
}

// CredentialRule model wraps regex and auth for grafana

func (g GrafanaConnection) User() string {
	return g["user"]
}

func (g GrafanaConnection) Password() string {
	return g["basicAuthPassword"]
}
