package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

const (
	ViperGdgConfig          = "gdg"
	ViperTemplateConfig     = "template"
	DefaultOrganizationName = "Main Org."
	DefaultOrganizationId   = 1
)

// GrafanaConnection Default connection credentials
type GrafanaConnection map[string]string

func (r RegexMatchesList) GetConnectionAuth(path string) (*GrafanaConnection, error) {
	if r.SecureData == "" {
		return nil, fmt.Errorf("no valid auth can be found for the given path %s", path)
	}
	secretLocation := filepath.Join(path, r.SecureData)
	result := new(GrafanaConnection)
	raw, err := os.ReadFile(secretLocation)
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
type CredentialRule struct {
	RegexMatchesList
	Auth *GrafanaConnection `mapstructure:"auth" yaml:"auth,omitempty"`
}

// MatchingRule defines a single matching rule for Grafana Connections
type MatchingRule struct {
	Field     string `yaml:"field,omitempty"`
	Regex     string `yaml:"regex,omitempty"`
	Inclusive bool   `yaml:"inclusive,omitempty"`
}

// FilterOverrides model wraps filter overrides for grafana
type FilterOverrides struct {
	IgnoreDashboardFilters bool `yaml:"ignore_dashboard_filters"`
}

// ConnectionFilters model wraps connection filters for grafana
type ConnectionFilters struct {
	NameExclusions  string   `yaml:"name_exclusions"`
	ConnectionTypes []string `yaml:"valid_types"`
	//	pattern         *regexp.Regexp
}

func (g GrafanaConnection) User() string {
	return g["user"]
}

func (g GrafanaConnection) Password() string {
	return g["basicAuthPassword"]
}
