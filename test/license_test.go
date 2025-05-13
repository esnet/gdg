package test

import (
	"log/slog"
	"testing"

	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service"

	"github.com/esnet/gdg/pkg/test_tooling"
	"github.com/esnet/gdg/pkg/test_tooling/containers"
	"github.com/stretchr/testify/assert"
)

func TestLicenseEnterpriseCheck(t *testing.T) {
	config.InitGdgConfig("testing")
	apiClient, _, cleanup := test_tooling.InitTest(t, service.DefaultConfigProvider, nil)
	defer cleanup()
	assert.False(t, apiClient.IsEnterprise())
	props := containers.DefaultGrafanaEnv()
	err := containers.SetupGrafanaLicense(&props)
	if err != nil {
		slog.Error("no valid grafana license found, skipping enterprise tests")
		t.Skip()
	}

	config.InitGdgConfig("testing")
	enterpriseClient, _, enterpriseCleanup := test_tooling.InitTest(t, service.DefaultConfigProvider, props)
	defer enterpriseCleanup()
	assert.True(t, enterpriseClient.IsEnterprise())
}
