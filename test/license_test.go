package test

import (
	"github.com/esnet/gdg/pkg/test_tooling"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLicenseEnterpriseCheck(t *testing.T) {
	apiClient, _, _, cleanup := test_tooling.InitTest(t, nil, false)
	defer cleanup()
	assert.False(t, apiClient.IsEnterprise())
	enterpriseClient, _, _, enterpriseCleanup := test_tooling.InitTest(t, nil, true)
	defer enterpriseCleanup()
	assert.True(t, enterpriseClient.IsEnterprise())

}
