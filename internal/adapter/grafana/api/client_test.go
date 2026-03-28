package api

import (
	"context"
	"os"
	"testing"

	"github.com/esnet/gdg/internal/adapter/grafana/extended"
	"github.com/esnet/gdg/internal/adapter/grafana/resources"
	"github.com/esnet/gdg/internal/adapter/plugins/secure/noop"
	"github.com/esnet/gdg/internal/adapter/storage"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/pkg/test_tooling/common"
	"github.com/esnet/gdg/pkg/test_tooling/path"
	"github.com/stretchr/testify/assert"
)

func fixEnvironment(t *testing.T) {
	assert.NoError(t, path.FixTestDir("api", "../../../.."))
	err := os.Setenv(common.ContextNameEnv, "qa")
	assert.Nil(t, err)
	config.NewConfig(common.DefaultTestConfig)
}
func TestRelativePathLogin(t *testing.T) {
	envKey := "GDG_CONTEXTS__QA__URL"
	assert.NoError(t, os.Setenv(envKey, "http://localhost:3000/grafana/"))
	fixEnvironment(t)
	resourcesHelpers := resources.NewHelpers()
	defer os.Unsetenv(common.ContextNameEnv)
	cfg := config.NewConfig(common.DefaultTestConfig)
	defer func() {
		assert.NoError(t, os.Unsetenv(envKey))
	}()

	localEngine := storage.NewLocalStorage(context.Background())
	svc := NewDashNGo(cfg, noop.NoOpEncoder{}, localEngine, extended.NewExtendedApi(cfg), resourcesHelpers)
	_, clientCfg := svc.(*DashNGoImpl).getNewClient()
	assert.Equal(t, clientCfg.Host, "localhost:3000")
	assert.Equal(t, clientCfg.BasePath, "/grafana/api")
}
