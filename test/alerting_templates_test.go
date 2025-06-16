package test

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"slices"
	"testing"

	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service"
	"github.com/esnet/gdg/pkg/test_tooling"
	"github.com/esnet/gdg/pkg/test_tooling/common"
	"github.com/esnet/gdg/pkg/test_tooling/path"
	"github.com/stretchr/testify/assert"
)

func TestTemplatesCrud(t *testing.T) {
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
	templates, err := apiClient.ListAlertTemplates()
	assert.NoError(t, err)
	assert.Equal(t, len(templates), 0, "Validate initial templates list is empty")
	uploadedTemplates, err := apiClient.UploadAlertTemplates()
	assert.NoError(t, err)
	assert.Equal(t, len(uploadedTemplates), 2)
	assert.True(t, slices.Contains(uploadedTemplates, "test_tpl1"))
	// Update TPL when they already exist
	uploadedTemplates, err = apiClient.UploadAlertTemplates()
	assert.NoError(t, err)
	assert.Equal(t, len(uploadedTemplates), 2)
	assert.True(t, slices.Contains(uploadedTemplates, "tpl2_test"))
	templates, err = apiClient.ListAlertTemplates()
	assert.NoError(t, err)
	assert.Equal(t, len(templates), 2)
	data, err := apiClient.DownloadAlertTemplates()
	assert.NoError(t, err)
	assert.Equal(t, "test/data/org_main-org/alerting/templates.json", data)
	rawData, err := os.ReadFile(data)
	assert.NoError(t, err)
	assert.True(t, bytes.Contains(rawData, []byte("test_tpl1")))
	assert.True(t, bytes.Contains(rawData, []byte("tpl2_test")))
	uploadedTemplates, err = apiClient.ClearAlertTemplates()
	assert.NoError(t, err)
	assert.Equal(t, len(uploadedTemplates), 2)
	templates, err = apiClient.ListAlertTemplates()
	assert.NoError(t, err)
	assert.Equal(t, len(templates), 0)
}
