package test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/esnet/gdg/internal/adapter/grafana/api"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/pkg/test_tooling"
	"github.com/esnet/gdg/pkg/test_tooling/common"
	"github.com/esnet/gdg/pkg/test_tooling/path"
	"github.com/stretchr/testify/assert"
)

func TestServerInfo(t *testing.T) {
	assert.NoError(t, os.Setenv(common.ContextNameEnv, common.TestContextName))
	assert.NoError(t, os.Unsetenv(common.ContextNameEnv))

	assert.NoError(t, path.FixTestDir("test", ".."))
	cfg := config.NewConfig(common.DefaultTestConfig)
	var r *test_tooling.InitContainerResult
	err := Retry(context.Background(), DefaultRetryAttempts, func() error {
		r = test_tooling.InitTest(t, cfg, nil)
		return r.Err
	})
	assert.NotNil(t, r)
	assert.NoError(t, err)
	defer func() {
		cleanupErr := r.CleanUp()
		if cleanupErr != nil {
			slog.Warn("Unable to clean up after test", "test", t.Name())
		}
	}()
	apiClient := r.ApiClient
	assert := assert.New(t)
	res := apiClient.GetServerInfo()
	assert.NotNil(res)
	assert.NotEmpty(res[api.SrvInfoDBKey])
	assert.NotEmpty(res[api.SrvInfoVersionKey])
	assert.NotEmpty(res[api.SrvInfoCommitKey])
	if apiClient.IsEnterprise() {
		assert.NotEmpty(res[api.SrvInfoEnterpriseCommitKey])
	} else {
		assert.Empty(res[api.SrvInfoEnterpriseCommitKey])
	}
}
