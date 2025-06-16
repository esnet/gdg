package test

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/esnet/gdg/internal/tools/ptr"
	"github.com/safaci2000/grafana-openapi-client-go/models"
	"github.com/samber/lo"

	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service"
	"github.com/esnet/gdg/pkg/test_tooling"
	"github.com/esnet/gdg/pkg/test_tooling/common"
	"github.com/esnet/gdg/pkg/test_tooling/path"
	"github.com/stretchr/testify/assert"
)

func TestAlertingRulesCrud(t *testing.T) {
	assert.NoError(t, os.Setenv("GDG_CONTEXT_NAME", common.TestContextName))

	assert.NoError(t, path.FixTestDir("test", ".."))
	config.InitGdgConfig(common.DefaultTestConfig)
	var r *test_tooling.InitContainerResult
	err := Retry(context.Background(), DefaultRetryAttempts, func() error {
		r = test_tooling.InitTest(t, service.DefaultConfigProvider, nil)
		return r.Err
	})
	assert.NotNil(t, r)
	assert.NoError(t, err)
	defer func() {
		err := r.CleanUp()
		if err != nil {
			slog.Warn("Unable to clean up after test", "test", t.Name())
		}
	}()
	apiClient := r.ApiClient
	slog.Info("Uploading Connections")
	conn := apiClient.UploadConnections(service.NewConnectionFilter(""))
	assert.True(t, len(conn) > 0)
	//
	slog.Info("Creating Folders")
	folders := apiClient.UploadFolders(nil)
	assert.True(t, len(folders) > 0)
	//
	slog.Info("Uploading Connections")
	_, err = apiClient.UploadContactPoints()
	assert.NoError(t, err)

	templates, err := apiClient.ListAlertRules()
	assert.NoError(t, err)
	assert.Equal(t, len(templates), 0, "Validate initial rules list is empty")
	err = apiClient.UploadAlertRules()
	assert.NoError(t, err)
	templates, err = apiClient.ListAlertRules()
	assert.NoError(t, err)
	assert.Equal(t, len(templates), 2)
	p := lo.FindOrElse(templates, nil, func(item *models.ProvisionedAlertRule) bool {
		return item.UID == "aeozpk1wn93b4b"
	})
	assert.NotNil(t, p)
	assert.Equal(t, len(p.Data), 2)
	assert.Equal(t, ptr.ValOf(p.Title), "boom")
	data, err := apiClient.DownloadAlertRules()
	assert.NoError(t, err)
	assert.Equal(t, "test/data/org_main-org/alerting/rules.json", data)
	rawData, err := os.ReadFile(data)
	assert.NoError(t, err)
	assert.True(t, bytes.Contains(rawData, []byte("prometheus")))
	assert.True(t, bytes.Contains(rawData, []byte("go_gc_duration_seconds")))
	uploadedTemplates, err := apiClient.ClearAlertRules()
	assert.NoError(t, err)
	assert.Equal(t, len(uploadedTemplates), 2)
	templates, err = apiClient.ListAlertRules()
	assert.NoError(t, err)
	assert.Equal(t, len(templates), 0)
}
