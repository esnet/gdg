package test

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"slices"
	"testing"

	"github.com/esnet/gdg/internal/adapter/grafana/api"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/pkg/test_tooling"
	"github.com/esnet/gdg/pkg/test_tooling/common"
	"github.com/esnet/gdg/pkg/test_tooling/path"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/stretchr/testify/assert"
)

func TestTemplatesCrud(t *testing.T) {
	assert.NoError(t, os.Setenv(common.ContextNameEnv, common.TestContextName))
	defer os.Unsetenv(common.ContextNameEnv)

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
		err := r.CleanUp()
		if err != nil {
			slog.Warn("Unable to clean up after test", "test", t.Name())
		}
	}()
	apiClient := r.ApiClient
	templates, err := apiClient.ListAlertTemplates(api.NewAlertTemplatesFilter(""))
	assert.NoError(t, err)
	assert.Equal(t, len(templates), 0, "Validate initial templates list is empty")
	uploadedTemplates, err := apiClient.UploadAlertTemplates(api.NewAlertTemplatesFilter(""))
	assert.NoError(t, err)
	assert.Equal(t, len(uploadedTemplates), 2)
	assert.True(t, slices.Contains(uploadedTemplates, "test_tpl1"))
	// Update TPL when they already exist
	uploadedTemplates, err = apiClient.UploadAlertTemplates(api.NewAlertTemplatesFilter(""))
	assert.NoError(t, err)
	assert.Equal(t, len(uploadedTemplates), 2)
	assert.True(t, slices.Contains(uploadedTemplates, "tpl2_test"))
	templates, err = apiClient.ListAlertTemplates(api.NewAlertTemplatesFilter(""))
	assert.NoError(t, err)
	assert.Equal(t, len(templates), 2)
	data, err := apiClient.DownloadAlertTemplates(api.NewAlertTemplatesFilter(""))
	assert.NoError(t, err)
	assert.Equal(t, "test/data/org_main-org/alerting/templates.json", data)
	rawData, err := os.ReadFile(data)
	assert.NoError(t, err)
	assert.True(t, bytes.Contains(rawData, []byte("test_tpl1")))
	assert.True(t, bytes.Contains(rawData, []byte("tpl2_test")))
	uploadedTemplates, err = apiClient.ClearAlertTemplates(api.NewAlertTemplatesFilter(""))
	assert.NoError(t, err)
	assert.Equal(t, len(uploadedTemplates), 2)
	templates, err = apiClient.ListAlertTemplates(api.NewAlertTemplatesFilter(""))
	assert.NoError(t, err)
	assert.Equal(t, len(templates), 0)
}

// TestTemplatesFilterTest validates that the regex based "filter" pattern (mirroring
// api.NewAlertTemplatesFilter) is correctly applied across List, Clear, Upload, and
// Download operations for alert templates.
func TestTemplatesFilterTest(t *testing.T) {
	assert.NoError(t, os.Setenv(common.ContextNameEnv, common.TestContextName))
	defer os.Unsetenv(common.ContextNameEnv)

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
		err := r.CleanUp()
		if err != nil {
			slog.Warn("Unable to clean up after test", "test", t.Name())
		}
	}()
	apiClient := r.ApiClient

	// Seed both fixture templates (test_tpl1, tpl2_test) unfiltered so subsequent
	// sub-tests have something meaningful to filter against.
	seeded, err := apiClient.UploadAlertTemplates(api.NewAlertTemplatesFilter(""))
	assert.NoError(t, err)
	assert.Equal(t, 2, len(seeded))

	type testConfig struct {
		name     string
		regex    string
		expected int
		validate func(t *testing.T, list []*models.NotificationTemplate)
	}

	t.Run("List filtering", func(t *testing.T) {
		testCases := []testConfig{
			{
				name:     "empty filter matches everything",
				regex:    "",
				expected: 2,
			},
			{
				name:     "exact name match",
				regex:    "^test_tpl1$",
				expected: 1,
				validate: func(t *testing.T, list []*models.NotificationTemplate) {
					assert.Equal(t, "test_tpl1", list[0].Name)
				},
			},
			{
				name:     "suffix match",
				regex:    "_test$",
				expected: 1,
				validate: func(t *testing.T, list []*models.NotificationTemplate) {
					assert.Equal(t, "tpl2_test", list[0].Name)
				},
			},
			{
				name:     "substring match on both templates",
				regex:    "tpl",
				expected: 2,
			},
			{
				name:     "no match",
				regex:    "^doesnotexist$",
				expected: 0,
			},
			{
				name:     "invalid regex matches nothing",
				regex:    "(unclosed",
				expected: 0,
			},
		}
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				filter := api.NewAlertTemplatesFilter(tc.regex)
				list, listErr := apiClient.ListAlertTemplates(filter)
				assert.NoError(t, listErr)
				assert.Equal(t, tc.expected, len(list))
				if tc.validate != nil {
					tc.validate(t, list)
				}
			})
		}
	})

	t.Run("Clear respects filter", func(t *testing.T) {
		cleared, clearErr := apiClient.ClearAlertTemplates(api.NewAlertTemplatesFilter("^test_tpl1$"))
		assert.NoError(t, clearErr)
		assert.Equal(t, 1, len(cleared))
		assert.Equal(t, "test_tpl1", cleared[0])

		remaining, listErr := apiClient.ListAlertTemplates(api.NewAlertTemplatesFilter(""))
		assert.NoError(t, listErr)
		assert.Equal(t, 1, len(remaining))
		assert.Equal(t, "tpl2_test", remaining[0].Name)

		// Restore the cleared template so later sub-tests see both templates again.
		restored, uploadErr := apiClient.UploadAlertTemplates(api.NewAlertTemplatesFilter("^test_tpl1$"))
		assert.NoError(t, uploadErr)
		assert.Equal(t, 1, len(restored))
		assert.Equal(t, "test_tpl1", restored[0])
	})

	t.Run("Upload respects filter", func(t *testing.T) {
		// Clear everything, then upload with a filter that only matches one template.
		_, clearErr := apiClient.ClearAlertTemplates(api.NewAlertTemplatesFilter(""))
		assert.NoError(t, clearErr)

		uploaded, uploadErr := apiClient.UploadAlertTemplates(api.NewAlertTemplatesFilter("_test$"))
		assert.NoError(t, uploadErr)
		assert.Equal(t, 1, len(uploaded))
		assert.Equal(t, "tpl2_test", uploaded[0])

		afterUpload, listErr := apiClient.ListAlertTemplates(api.NewAlertTemplatesFilter(""))
		assert.NoError(t, listErr)
		assert.Equal(t, 1, len(afterUpload))
		assert.Equal(t, "tpl2_test", afterUpload[0].Name)

		// Restore full fixture set for the remaining sub-tests.
		restored, uploadErr := apiClient.UploadAlertTemplates(api.NewAlertTemplatesFilter(""))
		assert.NoError(t, uploadErr)
		assert.Equal(t, 2, len(restored))
	})

	t.Run("Download respects filter", func(t *testing.T) {
		data, downloadErr := apiClient.DownloadAlertTemplates(api.NewAlertTemplatesFilter("_test$"))
		assert.NoError(t, downloadErr)
		rawData, readErr := os.ReadFile(data)
		assert.NoError(t, readErr)
		assert.True(t, bytes.Contains(rawData, []byte("tpl2_test")))
		assert.False(t, bytes.Contains(rawData, []byte("test_tpl1")))
	})
}
