package test

import (
	"fmt"
	"log/slog"
	"os"
	"slices"
	"testing"

	"github.com/esnet/gdg/internal/adapter/grafana/api"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/config/config_domain"
	"github.com/esnet/gdg/internal/domain"
	"github.com/esnet/gdg/internal/ports/outbound"
	"github.com/esnet/gdg/pkg/ptr"
	"github.com/esnet/gdg/pkg/test_tooling"
	"github.com/esnet/gdg/pkg/test_tooling/common"
	"github.com/esnet/gdg/pkg/test_tooling/path"
	"github.com/matryer/is"
	"github.com/samber/lo"
	"github.com/tidwall/sjson"
)

func TestAlertingRulesCrud(t *testing.T) {
	is, cfg := setupRuleTest(t)
	r := test_tooling.InitTest(t, cfg, nil)
	is.True(r != nil)
	defer func() {
		cleanupErr := r.CleanUp()
		if cleanupErr != nil {
			slog.Warn("Unable to clean up after test", "test", t.Name())
		}
	}()
	apiClient := r.ApiClient
	setupAlertingEnvironment(is, apiClient, true)

	alertCfg := domain.AlertRuleFilterParams{IgnoreWatchedFolders: false}
	alertFilters := api.NewAlertRuleFilter(cfg, apiClient, alertCfg)
	rulesList, err := apiClient.ListAlertRules(alertFilters)
	is.NoErr(err)
	is.Equal(len(rulesList), 0)
	rules, err := apiClient.UploadAlertRules(alertFilters)
	is.Equal(len(rules), 2)
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
	is, cfg := setupRuleTest(t)
	r := test_tooling.InitTest(t, cfg, nil)
	is.True(r != nil)
	defer func() {
		cleanupErr := r.CleanUp()
		if cleanupErr != nil {
			slog.Warn("Unable to clean up after test", "test", t.Name())
		}
	}()
	apiClient := r.ApiClient
	setupAlertingEnvironment(is, apiClient, true)
	// Upload All rules
	alertFilters := api.NewAlertRuleFilter(cfg, apiClient, domain.AlertRuleFilterParams{IgnoreWatchedFolders: true})
	_, err := apiClient.UploadAlertRules(alertFilters)
	is.NoErr(err)

	type testConfig struct {
		enabled  bool
		name     string
		alertCfg domain.AlertRuleFilterParams
		expected int
		validate func(alertCfg domain.AlertRuleFilterParams, list []*domain.AlertRuleWithNestedFolder)
	}

	t.Run("Running Basic Tests", func(t *testing.T) {
		// Now let's fetch that data
		testCases := []testConfig{
			{
				name:     "FetchAll",
				enabled:  true,
				expected: 4,
				alertCfg: domain.AlertRuleFilterParams{
					IgnoreWatchedFolders: true,
				},
			},
			{
				name:     "Filter by Folder, Ignore watch folders",
				enabled:  true,
				expected: 2,
				alertCfg: domain.AlertRuleFilterParams{
					IgnoreWatchedFolders: true,
					Folder:               "Ignored",
				},
				validate: func(alertCfg domain.AlertRuleFilterParams, rulesList []*domain.AlertRuleWithNestedFolder) {
					matchingList := lo.Uniq(
						lo.Map(rulesList, func(item *domain.AlertRuleWithNestedFolder, index int) string {
							return item.NestedPath
						}),
					)
					is.Equal(len(matchingList), 1)
					is.Equal(matchingList[0], "Ignored")
				},
			},
			{
				name:     "Filter by Label, Ignore watch folders",
				enabled:  true,
				expected: 3,
				alertCfg: domain.AlertRuleFilterParams{
					IgnoreWatchedFolders: true,
					Label:                []string{"environment=alpha"},
				},
				validate: func(alertCfg domain.AlertRuleFilterParams, rulesList []*domain.AlertRuleWithNestedFolder) {
					matchingList := lo.Uniq(
						lo.Map(rulesList, func(item *domain.AlertRuleWithNestedFolder, index int) string {
							for key, val := range item.Labels {
								matchingKey := fmt.Sprintf("%s=%s", key, val)
								if slices.Contains(alertCfg.Label, matchingKey) {
									return matchingKey
								}
							}
							return item.NestedPath
						}),
					)
					is.Equal(len(matchingList), 1)
					is.Equal(matchingList[0], "environment=alpha")
				},
			},
			{
				name:     "Filter by Label and folder, Ignore watch folders",
				enabled:  true,
				expected: 1,
				alertCfg: domain.AlertRuleFilterParams{
					IgnoreWatchedFolders: true,
					Folder:               "Ignored",
					Label:                []string{"environment=alpha"},
				},
				validate: func(alertCfg domain.AlertRuleFilterParams, rulesList []*domain.AlertRuleWithNestedFolder) {
					is.Equal(rulesList[0].NestedPath, "Ignored")
					is.Equal(rulesList[0].Labels["environment"], "alpha")
				},
			},
			{
				name:     "using watch folders",
				enabled:  true,
				expected: 2,
				alertCfg: domain.AlertRuleFilterParams{
					IgnoreWatchedFolders: false,
				},
				validate: func(alertCfg domain.AlertRuleFilterParams, rulesList []*domain.AlertRuleWithNestedFolder) {
					matchingList := lo.Uniq(
						lo.Map(rulesList, func(item *domain.AlertRuleWithNestedFolder, index int) string {
							return item.NestedPath
						}),
					)
					is.Equal(len(matchingList), 2)
					is.True(slices.Contains(matchingList, "linux%2Fgnu/Others/n%2B_%3D23r"))
					is.True(slices.Contains(matchingList, "linux%2Fgnu/Others"))
				},
			},
			{
				name:     "label filter, using watch folders",
				enabled:  true,
				expected: 1,
				alertCfg: domain.AlertRuleFilterParams{
					IgnoreWatchedFolders: false,
					Label:                []string{"deployed=true"},
				},
				validate: func(alertCfg domain.AlertRuleFilterParams, rulesList []*domain.AlertRuleWithNestedFolder) {
					is.Equal(rulesList[0].NestedPath, "linux%2Fgnu/Others")
					is.Equal(rulesList[0].Labels["deployed"], "true")
				},
			},
			{
				name:     "folder filter, using watch folders",
				enabled:  true,
				expected: 2,
				alertCfg: domain.AlertRuleFilterParams{
					IgnoreWatchedFolders: false,
					Folder:               "linux%2Fgnu/*",
				},
			},
			{
				name:     "folder filter and labels, using watch folders",
				enabled:  true,
				expected: 2,
				alertCfg: domain.AlertRuleFilterParams{
					IgnoreWatchedFolders: false,
					Folder:               "linux%2Fgnu/Others*",
					Label:                []string{"deployed=true", "environment=alpha"},
				},
			},
		}
		// Validate testing behavior
		for _, tc := range testCases {
			if !tc.enabled {
				t.Log("Skipping test", tc.name)
				continue
			}
			t.Logf("Running test case: %s", tc.name)
			alertFilters = api.NewAlertRuleFilter(cfg, apiClient, tc.alertCfg)
			rulesList, err := apiClient.ListAlertRules(alertFilters)
			is.NoErr(err)
			is.Equal(len(rulesList), tc.expected)
			if tc.validate != nil {
				tc.validate(tc.alertCfg, rulesList)
			}

		}
	})

	t.Run("name based tests", func(t *testing.T) {
		nameFilterTests := []testConfig{
			{
				name:     "name based filter, ignoring watch folders",
				enabled:  true,
				expected: 1,
				alertCfg: domain.AlertRuleFilterParams{
					IgnoreWatchedFolders: true,
					UID:                  "aeozpk1wn93b4b",
				},
			},
			{
				name:     "name based filter, using watch folders",
				enabled:  true,
				expected: 0,
				alertCfg: domain.AlertRuleFilterParams{
					IgnoreWatchedFolders: false,
					UID:                  "aeozpk1wn93b4b",
				},
			},
			{
				name:     "name based filter, using watch folders, matching",
				enabled:  true,
				expected: 1,
				alertCfg: domain.AlertRuleFilterParams{
					IgnoreWatchedFolders: false,
					UID:                  "afejt30qxdk3kb",
				},
			},
			{
				name:     "name based filter, ignoring watch folders",
				enabled:  true,
				expected: 1,
				alertCfg: domain.AlertRuleFilterParams{
					IgnoreWatchedFolders: true,
					UID:                  "afejt30qxdk3kb",
				},
			},
		}
		for _, tc := range nameFilterTests {
			if !tc.enabled {
				t.Log("Skipping test", tc.name)
				continue
			}
			t.Logf("Running test case: %s", tc.name)
			alertFilters = api.NewAlertRuleFilter(cfg, apiClient, tc.alertCfg)
			rulesList, err := apiClient.ListAlertRules(alertFilters)
			is.NoErr(err)
			is.Equal(len(rulesList), tc.expected)
			if tc.validate != nil {
				tc.validate(tc.alertCfg, rulesList)
			}

		}
	})
}

// TestSingleUploadFolderMatch Test the patching behavior by omitting the folder creation
func TestSingleUploadMissingFolderMatch(t *testing.T) {
	is, cfg := setupRuleTest(t)
	alertCfg := domain.AlertRuleFilterParams{
		IgnoreWatchedFolders: true,
		UID:                  "afejt30qxdk3kb",
	}
	r := test_tooling.InitTest(t, cfg, nil)
	is.True(r != nil)
	apiClient := r.ApiClient
	setupAlertingEnvironment(is, apiClient, false)
	alertFilters := api.NewAlertRuleFilter(cfg, apiClient, alertCfg)
	rules, err := r.ApiClient.UploadAlertRules(alertFilters)
	is.NoErr(err)
	is.Equal(len(rules), 1)
	is.Equal(rules[0].NestedPath, "linux%2Fgnu/Others")
	fldList := apiClient.ListFolders(nil)
	is.True(fldList != nil)
	is.Equal(len(fldList), 2)
	entry, found := lo.Find(fldList, func(item *domain.NestedHit) bool {
		return item.NestedPath == "linux%2Fgnu/Others"
	})
	is.True(found)
	is.Equal(ptr.ValueOrDefault(rules[0].FolderUID, ""), entry.UID)
}

// TestSingleUploadFolderMatch Test the patching behavior by omitting the folder creation
func TestSingleUploadFolderUIDMismatch(t *testing.T) {
	is, cfg := setupRuleTest(t)
	r := test_tooling.InitTest(t, cfg, nil)
	is.True(r != nil)
	apiClient := r.ApiClient
	setupAlertingEnvironment(is, apiClient, true)

	// Read current data
	data, err := os.ReadFile("test/data/org_main-org/alerting-rules/linux%2Fgnu/Others/woof.json")
	is.NoErr(err)
	data, err = sjson.SetBytes(data, "uid", "testing")
	is.NoErr(err)
	data, err = sjson.SetBytes(data, "folderUID", "dummyValue")
	is.NoErr(err)
	const name = "test/data/org_main-org/alerting-rules/linux%2Fgnu/Others/woof_invalid.json"
	err = os.WriteFile(name, data, 0o644)
	is.NoErr(err)
	defer func() {
		os.Remove(name)
	}()

	alertCfg := domain.AlertRuleFilterParams{
		IgnoreWatchedFolders: true,
		UID:                  "testing",
	}
	alertFilters := api.NewAlertRuleFilter(cfg, apiClient, alertCfg)
	rules, err := r.ApiClient.UploadAlertRules(alertFilters)
	is.NoErr(err)
	is.Equal(len(rules), 1)
	is.Equal(rules[0].NestedPath, "linux%2Fgnu/Others")
	rules, err = r.ApiClient.UploadAlertRules(alertFilters)
	ruleEntry, found := lo.Find(rules, func(item *domain.AlertRuleWithNestedFolder) bool {
		return item.UID == "testing"
	})
	is.True(found)
	is.NoErr(err)
	fldList := apiClient.ListFolders(nil)
	is.True(fldList != nil)
	is.Equal(len(fldList), 4)
	folderEntry, found := lo.Find(fldList, func(item *domain.NestedHit) bool {
		return item.NestedPath == "linux%2Fgnu/Others"
	})
	is.True(found)
	is.Equal(folderEntry.UID, ptr.ValueOrDefault(ruleEntry.FolderUID, ""))
	is.Equal(folderEntry.NestedPath, ruleEntry.NestedPath)
}

func TestAlertingRulesNoFilterCrud(t *testing.T) {
	is, cfg := setupRuleTest(t)
	r := test_tooling.InitTest(t, cfg, nil)
	is.True(r != nil)
	defer func() {
		cleanupErr := r.CleanUp()
		if cleanupErr != nil {
			slog.Warn("Unable to clean up after test", "test", t.Name())
		}
	}()
	apiClient := r.ApiClient

	setupAlertingEnvironment(is, apiClient, true)

	f := domain.AlertRuleFilterParams{
		IgnoreWatchedFolders: true,
	}
	alertFilters := api.NewAlertRuleFilter(cfg, apiClient, f)
	rulesList, err := apiClient.ListAlertRules(alertFilters)
	is.NoErr(err)
	is.Equal(len(rulesList), 0)
	_, err = apiClient.UploadAlertRules(alertFilters)
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

// setupAlertingEnvironment prepares the prerequisite resources needed for alerting tests by uploading
// connections, creating folders, and uploading contact points via the provided GrafanaService API client.
func setupAlertingEnvironment(is *is.I, apiClient outbound.GrafanaService, includeFolder bool) {
	slog.Info("Uploading Connections")
	conn := apiClient.UploadConnections(api.NewConnectionFilter(""))
	is.True(len(conn) > 0)
	//
	if includeFolder {
		slog.Info("Creating Folders")
		folders := apiClient.UploadFolders(nil)
		is.True(len(folders) > 0)
	} else {
		slog.Info("skipping folders creation")
	}

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

// setupRuleTest initializes the test environment for rule-related tests by configuring the test context,
// fixing the working directory, and loading the default test configuration. It returns an assertion
// helper and the loaded application configuration.
func setupRuleTest(t *testing.T) (*is.I, *config_domain.GDGAppConfiguration) {
	is := is.New(t)
	is.NoErr(os.Setenv(common.ContextNameEnv, common.TestContextName))
	is.NoErr(os.Unsetenv(common.ContextNameEnv))
	is.NoErr(path.FixTestDir("test", ".."))
	cfg := config.NewConfig(common.DefaultTestConfig)
	return is, cfg
}
