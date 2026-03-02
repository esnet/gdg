package test

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"testing"

	"github.com/esnet/gdg/internal/adapter/grafana/api"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/domain"
	"github.com/esnet/gdg/internal/ports"
	"github.com/esnet/gdg/pkg/ptr"
	"github.com/esnet/gdg/pkg/test_tooling"
	"github.com/esnet/gdg/pkg/test_tooling/common"
	"github.com/esnet/gdg/pkg/test_tooling/path"
	"github.com/matryer/is"
	"github.com/samber/lo"
)

func TestAlertingRulesCrud(t *testing.T) {
	is := is.New(t)
	is.NoErr(os.Setenv(common.ContextNameEnv, common.TestContextName))
	is.NoErr(os.Unsetenv(common.ContextNameEnv))

	is.NoErr(path.FixTestDir("test", ".."))
	cfg := config.NewConfig(common.DefaultTestConfig)
	var r *test_tooling.InitContainerResult
	err := Retry(context.Background(), DefaultRetryAttempts, func() error {
		r = test_tooling.InitTest(t, cfg, nil)
		return r.Err
	})
	is.True(r != nil)
	is.NoErr(err)
	defer func() {
		cleanupErr := r.CleanUp()
		if cleanupErr != nil {
			slog.Warn("Unable to clean up after test", "test", t.Name())
		}
	}()
	apiClient := r.ApiClient
	setupAlertingEnvironment(t, apiClient)

	f := domain.AlertRuleFilterParams{IgnoreWatchedFolders: false}
	alertFilters := api.NewAlertRuleFilter(cfg, apiClient, f)
	rulesList, err := apiClient.ListAlertRules(alertFilters)
	is.NoErr(err)
	is.Equal(len(rulesList), 0)
	err = apiClient.UploadAlertRules(alertFilters)
	is.NoErr(err)
	rulesList, err = apiClient.ListAlertRules(alertFilters)
	is.NoErr(err)
	is.Equal(len(rulesList), 2)
	validateMooEntity(is, rulesList, "ceozp0ovszy80c")
	data, err := apiClient.DownloadAlertRules(alertFilters)
	is.NoErr(err)
	is.Equal(len(data), 2)
	uploadedTemplates, err := apiClient.ClearAlertRules(alertFilters)
	is.NoErr(err)
	is.Equal(len(uploadedTemplates), 2)
	rulesList, err = apiClient.ListAlertRules(alertFilters)
	is.NoErr(err)
	is.Equal(len(rulesList), 0)
}

func TestAlertingRulesFilterTest(t *testing.T) {
	is := is.New(t)
	is.NoErr(os.Setenv(common.ContextNameEnv, common.TestContextName))
	is.NoErr(os.Unsetenv(common.ContextNameEnv))

	is.NoErr(path.FixTestDir("test", ".."))
	cfg := config.NewConfig(common.DefaultTestConfig)
	var r *test_tooling.InitContainerResult
	err := Retry(context.Background(), DefaultRetryAttempts, func() error {
		r = test_tooling.InitTest(t, cfg, nil)
		return r.Err
	})
	is.True(r != nil)
	is.NoErr(err)
	defer func() {
		cleanupErr := r.CleanUp()
		if cleanupErr != nil {
			slog.Warn("Unable to clean up after test", "test", t.Name())
		}
	}()
	apiClient := r.ApiClient
	setupAlertingEnvironment(t, apiClient)

	f := domain.AlertRuleFilterParams{
		IgnoreWatchedFolders: true,
	}
	alertFilters := api.NewAlertRuleFilter(cfg, apiClient, f)
	// Upload everything
	err = apiClient.UploadAlertRules(alertFilters)
	is.NoErr(err)
	// Ignore Watched filters
	rulesList, err := apiClient.ListAlertRules(alertFilters)
	is.NoErr(err)
	is.Equal(len(rulesList), 4)
	// Filter by folder
	f.Folder = "Ignored"
	alertFilters = api.NewAlertRuleFilter(cfg, apiClient, f)
	rulesList, err = apiClient.ListAlertRules(alertFilters)
	is.Equal(len(rulesList), 2)
	is.NoErr(err)
	matchingList := lo.Uniq(
		lo.Map(rulesList, func(item *domain.AlertRuleWithNestedFolder, index int) string {
			return item.NestedPath
		}),
	)
	is.Equal(len(matchingList), 1)
	is.Equal(matchingList[0], "Ignored")
	//Filter by Tags
	//
	f = domain.AlertRuleFilterParams{
		IgnoreWatchedFolders: true,
		Label:                []string{"environment=alpha"},
	}
	alertFilters = api.NewAlertRuleFilter(cfg, apiClient, f)
	rulesList, err = apiClient.ListAlertRules(alertFilters)
	is.NoErr(err)
	is.Equal(len(rulesList), 3)
	matchingList = lo.Uniq(
		lo.Map(rulesList, func(item *domain.AlertRuleWithNestedFolder, index int) string {
			for key, val := range item.Labels {
				matchingKey := fmt.Sprintf("%s=%s", key, val)
				if slices.Contains(f.Label, matchingKey) {
					return matchingKey
				}
			}
			return item.NestedPath
		}),
	)
	is.Equal(len(matchingList), 1)
	is.Equal(matchingList[0], "environment=alpha")
	// Ignore + dual filter
	f = domain.AlertRuleFilterParams{
		IgnoreWatchedFolders: true,
		Folder:               "Ignored",
		Label:                []string{"environment=alpha"},
	}
	alertFilters = api.NewAlertRuleFilter(cfg, apiClient, f)
	rulesList, err = apiClient.ListAlertRules(alertFilters)
	is.NoErr(err)
	is.Equal(len(rulesList), 1)
	is.Equal(rulesList[0].NestedPath, "Ignored")
	is.Equal(rulesList[0].Labels["environment"], "alpha")
	// Now same filters but with IgnoreWatchedFolders being false
	f = domain.AlertRuleFilterParams{
		IgnoreWatchedFolders: false,
	}
	alertFilters = api.NewAlertRuleFilter(cfg, apiClient, f)
	rulesList, err = apiClient.ListAlertRules(alertFilters)
	is.NoErr(err)
	is.Equal(len(rulesList), 2)
	matchingList = lo.Uniq(
		lo.Map(rulesList, func(item *domain.AlertRuleWithNestedFolder, index int) string {
			return item.NestedPath
		}),
	)
	is.Equal(len(matchingList), 2)
	is.True(slices.Contains(matchingList, "linux%2Fgnu/Others/n%2B_%3D23r"))
	is.True(slices.Contains(matchingList, "linux%2Fgnu/Others"))
	f = domain.AlertRuleFilterParams{
		IgnoreWatchedFolders: false,
		Label:                []string{"deployed=true"},
	}
	alertFilters = api.NewAlertRuleFilter(cfg, apiClient, f)
	rulesList, err = apiClient.ListAlertRules(alertFilters)
	is.NoErr(err)
	is.Equal(len(rulesList), 1)
	is.Equal(rulesList[0].NestedPath, "linux%2Fgnu/Others")
	is.Equal(rulesList[0].Labels["deployed"], "true")
	// Folder Filter
	f = domain.AlertRuleFilterParams{
		IgnoreWatchedFolders: false,
		Folder:               "linux%2Fgnu/*",
	}
	alertFilters = api.NewAlertRuleFilter(cfg, apiClient, f)
	rulesList, err = apiClient.ListAlertRules(alertFilters)
	is.NoErr(err)
	is.Equal(len(rulesList), 2)
	// Both Filters using Ignore Watch Additive behavior on labels
	f = domain.AlertRuleFilterParams{
		IgnoreWatchedFolders: false,
		Folder:               "linux%2Fgnu/Others*",
		Label:                []string{"deployed=true", "environment=alpha"},
	}
	alertFilters = api.NewAlertRuleFilter(cfg, apiClient, f)
	rulesList, err = apiClient.ListAlertRules(alertFilters)
	is.NoErr(err)
	is.Equal(len(rulesList), 2)
}

func TestAlertingRulesNoFilterCrud(t *testing.T) {
	is := is.New(t)
	is.NoErr(os.Setenv(common.ContextNameEnv, common.TestContextName))
	is.NoErr(os.Unsetenv(common.ContextNameEnv))
	is.NoErr(path.FixTestDir("test", ".."))
	cfg := config.NewConfig(common.DefaultTestConfig)
	var r *test_tooling.InitContainerResult
	err := Retry(context.Background(), DefaultRetryAttempts, func() error {
		r = test_tooling.InitTest(t, cfg, nil)
		return r.Err
	})
	is.True(r != nil)
	is.NoErr(err)
	defer func() {
		cleanupErr := r.CleanUp()
		if cleanupErr != nil {
			slog.Warn("Unable to clean up after test", "test", t.Name())
		}
	}()
	apiClient := r.ApiClient

	setupAlertingEnvironment(t, apiClient)

	f := domain.AlertRuleFilterParams{
		IgnoreWatchedFolders: true,
	}
	alertFilters := api.NewAlertRuleFilter(cfg, apiClient, f)
	rulesList, err := apiClient.ListAlertRules(alertFilters)
	is.NoErr(err)
	is.Equal(len(rulesList), 0)
	err = apiClient.UploadAlertRules(alertFilters)
	is.NoErr(err)
	rulesList, err = apiClient.ListAlertRules(alertFilters)
	is.NoErr(err)
	is.Equal(len(rulesList), 4)
	validateMooEntity(is, rulesList, "ceozp0ovszy80c")
	data, err := apiClient.DownloadAlertRules(alertFilters)
	is.NoErr(err)
	is.Equal(len(data), 4)
	uploadedTemplates, err := apiClient.ClearAlertRules(alertFilters)
	is.NoErr(err)
	is.Equal(len(uploadedTemplates), 4)
	rulesList, err = apiClient.ListAlertRules(alertFilters)
	is.NoErr(err)
	is.Equal(len(rulesList), 0)
}

func setupAlertingEnvironment(t *testing.T, apiClient ports.GrafanaService) {
	is := is.New(t)
	slog.Info("Uploading Connections")
	conn := apiClient.UploadConnections(api.NewConnectionFilter(""))
	is.True(len(conn) > 0)
	//
	slog.Info("Creating Folders")
	folders := apiClient.UploadFolders(nil)
	is.True(len(folders) > 0)
	//
	slog.Info("Uploading Connections")
	_, err := apiClient.UploadContactPoints()
	is.NoErr(err)
}

func validateMooEntity(is *is.I, rulesList []*domain.AlertRuleWithNestedFolder, id string) {
	p := lo.FindOrElse(rulesList, nil, func(item *domain.AlertRuleWithNestedFolder) bool {
		return item.UID == id
	})
	is.True(p != nil)
	is.Equal(len(p.ProvisionedAlertRule.Data), 2)
	is.Equal(ptr.ValueOrDefault(p.ProvisionedAlertRule.Title, ""), "moo")
}
