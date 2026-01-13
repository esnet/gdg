package test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	customModels "github.com/esnet/gdg/internal/service/domain"
	"github.com/esnet/gdg/internal/service/filters"
	"github.com/esnet/gdg/internal/tools/ptr"
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
	cfg := config.InitGdgConfig(common.DefaultTestConfig)
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

	alertFilters := service.NewAlertRuleFilter(cfg, apiClient)
	rulesList, err := apiClient.ListAlertRules(alertFilters)
	assert.NoError(t, err)
	assert.Equal(t, len(rulesList), 0, "Validate initial rules list is empty")
	err = apiClient.UploadAlertRules(alertFilters)
	assert.NoError(t, err)
	rulesList, err = apiClient.ListAlertRules(alertFilters)
	assert.NoError(t, err)
	assert.Equal(t, len(rulesList), 1)
	p := lo.FindOrElse(rulesList, nil, func(item *customModels.AlertRuleWithNestedFolder) bool {
		return item.UID == "ceozp0ovszy80c"
	})
	assert.NotNil(t, p)
	assert.Equal(t, len(p.ProvisionedAlertRule.Data), 2)
	assert.Equal(t, ptr.ValueOrDefault(p.ProvisionedAlertRule.Title, ""), "moo")
	data, err := apiClient.DownloadAlertRules(alertFilters)
	assert.NoError(t, err)
	assert.Equal(t, len(data), 1)
	uploadedTemplates, err := apiClient.ClearAlertRules(alertFilters)
	assert.NoError(t, err)
	assert.Equal(t, len(uploadedTemplates), 1)
	rulesList, err = apiClient.ListAlertRules(alertFilters)
	assert.NoError(t, err)
	assert.Equal(t, len(rulesList), 0)
}

func TestAlertingRulesNoFilterCrud(t *testing.T) {
	assert.NoError(t, os.Setenv("GDG_CONTEXT_NAME", common.TestContextName))

	assert.NoError(t, path.FixTestDir("test", ".."))
	cfg := config.InitGdgConfig(common.DefaultTestConfig)
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
	slog.Info("Uploading Connections")
	conn := apiClient.UploadConnections(service.NewConnectionFilter(""))
	assert.True(t, len(conn) > 0)
	//
	slog.Info("Creating Folders")
	folders := apiClient.UploadFolders(nil)
	assert.True(t, len(folders) > 0)
	//
	slog.Info("Uploading Contact Points")
	_, err = apiClient.UploadContactPoints()
	assert.NoError(t, err)

	var alertFilters filters.V2Filter = nil
	rulesList, err := apiClient.ListAlertRules(alertFilters)
	assert.NoError(t, err)
	assert.Equal(t, len(rulesList), 0, "Validate initial rules list is empty")
	err = apiClient.UploadAlertRules(alertFilters)
	assert.NoError(t, err)
	rulesList, err = apiClient.ListAlertRules(alertFilters)
	assert.NoError(t, err)
	assert.Equal(t, len(rulesList), 2)
	p := lo.FindOrElse(rulesList, nil, func(item *customModels.AlertRuleWithNestedFolder) bool {
		return item.UID == "ceozp0ovszy80c"
	})
	assert.NotNil(t, p)
	assert.Equal(t, len(p.ProvisionedAlertRule.Data), 2)
	assert.Equal(t, ptr.ValueOrDefault(p.ProvisionedAlertRule.Title, ""), "moo")
	data, err := apiClient.DownloadAlertRules(alertFilters)
	assert.NoError(t, err)
	assert.Equal(t, len(data), 2)
	uploadedTemplates, err := apiClient.ClearAlertRules(alertFilters)
	assert.NoError(t, err)
	assert.Equal(t, len(uploadedTemplates), 2)
	rulesList, err = apiClient.ListAlertRules(alertFilters)
	assert.NoError(t, err)
	assert.Equal(t, len(rulesList), 0)
}
