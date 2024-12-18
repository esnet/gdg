package test_tooling

import (
	"log/slog"
	"testing"

	"github.com/esnet/gdg/internal/config"

	"github.com/esnet/gdg/internal/service"
	"github.com/esnet/gdg/pkg/test_tooling/containers"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
)

const (
	OrgNameOverride      = "GDG_CONTEXTS__TESTING__ORGANIZATION_NAME"
	EnableNestedBehavior = "GDG_CONTEXTS__TESTING__DASHBOARD_SETTINGS__NESTED_FOLDERS"
	grafanaNestedToggle  = "GF_FEATURE_TOGGLES_ENABLE"
	IgnoreDashFilters    = "GDG_CONTEXTS__TESTING__DASHBOARD_SETTINGS__IGNORE_FILTERS"
)

// setupNestedProps adds the nestedFolder feature to the given Env properties map
func setupNestedProps(t *testing.T, enterprise bool) map[string]string {
	props := containers.DefaultGrafanaEnv()
	props[grafanaNestedToggle] = "nestedFolders"
	if enterprise {
		err := containers.SetupGrafanaLicense(&props)
		if err != nil {
			slog.Error("no valid grafana license found, skipping enterprise tests")
			t.Skip()
		}
	}

	return props
}

// InitOrganizations will upload all known organizations and return the grafana container object
func InitOrganizations(t *testing.T) (testcontainers.Container, func() error) {
	if config.Config() == nil {
		config.InitGdgConfig("testing")
	}
	props := setupNestedProps(t, false)
	apiClient, containerObj, cleanup := InitTest(t, service.DefaultConfigProvider, props)
	newOrgs := apiClient.UploadOrganizations(service.NewOrganizationFilter())
	assert.Equal(t, 4, len(newOrgs))
	return containerObj, cleanup
}
