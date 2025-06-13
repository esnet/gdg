package test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/esnet/gdg/pkg/test_tooling/common"

	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service"

	"github.com/esnet/gdg/pkg/test_tooling"
	"github.com/esnet/gdg/pkg/test_tooling/containers"
	"github.com/stretchr/testify/assert"
)

func TestLicenseEnterpriseCheck(t *testing.T) {
	tcCases := []struct {
		name       string
		disabled   bool
		enterprise bool
	}{
		{
			name:       "OS Licence test",
			enterprise: false,
		},
		{
			name:       "Enterprise Licence test",
			enterprise: true,
		},
	}

	for _, tc := range tcCases {
		if tc.disabled {
			t.Log("Skipping test", tc.name)
		}
		t.Log("Running test", tc.name)
		config.InitGdgConfig(common.DefaultTestConfig)
		var r *test_tooling.InitContainerResult
		var props map[string]string
		if tc.enterprise {
			props = containers.DefaultGrafanaEnv()
			err := containers.SetupGrafanaLicense(&props)
			assert.NoError(t, err)
		}
		err := Retry(context.Background(), DefaultRetryAttempts, func() error {
			r = test_tooling.InitTest(t, service.DefaultConfigProvider, props)
			return r.Err
		})
		assert.NotNil(t, r)
		assert.NoError(t, err)
		assert.Equal(t, r.ApiClient.IsEnterprise(), tc.enterprise)
		func() {
			err := r.CleanUp()
			if err != nil {
				slog.Warn("Unable to clean up after test", "test", t.Name())
			}
		}()
	}
}
