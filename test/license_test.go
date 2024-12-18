package test

import (
	"log/slog"
	"testing"

	"github.com/esnet/gdg/pkg/test_tooling"
	"github.com/esnet/gdg/pkg/test_tooling/containers"
	"github.com/stretchr/testify/assert"
)

func TestLicenseEnterpriseCheck(t *testing.T) {
	apiClient, _, _, cleanup := test_tooling.InitTestLegacy(t, nil, nil)
	defer cleanup()
	assert.False(t, apiClient.IsEnterprise())
	props := containers.DefaultGrafanaEnv()
	err := containers.SetupGrafanaLicense(&props)
	if err != nil {
		slog.Error("no valid grafana license found, skipping enterprise tests")
		t.Skip()
	}
	enterpriseClient, _, _, enterpriseCleanup := test_tooling.InitTestLegacy(t, nil, props)
	defer enterpriseCleanup()
	assert.True(t, enterpriseClient.IsEnterprise())
}
