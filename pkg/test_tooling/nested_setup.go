package test_tooling

import (
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
)

// InitOrganizations will upload all known organizations and return the grafana container object
func InitOrganizations(t *testing.T) (testcontainers.Container, func() error) {
	if config.Config() == nil {
		config.InitGdgConfig("testing")
	}
	props := containers.DefaultGrafanaEnv()
	r := InitTest(t, service.DefaultConfigProvider, props)
	newOrgs := r.ApiClient.UploadOrganizations(service.NewOrganizationFilter())
	assert.Equal(t, 4, len(newOrgs))
	return r.Container, r.CleanUp
}
