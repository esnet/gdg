package backup_test

import (
	"fmt"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/esnet/gdg/cli"
	customModels "github.com/esnet/gdg/internal/domain"
	"github.com/esnet/gdg/internal/ports/outbound/mocks"
	"github.com/esnet/gdg/pkg/test_tooling"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListAlertRules(t *testing.T) {
	listCmd := []string{"backup", "alerting", "rules", "list"}
	testCases := []struct {
		skip       bool
		name       string
		validateFn func(t *testing.T, output string)
		setupMocks func(testSvc *mocks.GrafanaService)
		expectErr  bool
	}{
		{
			name: "NoDataTest",
			validateFn: func(t *testing.T, output string) {
				assert.True(t, strings.Contains(output, "No alert rules found"))
			},
			setupMocks: func(testSvc *mocks.GrafanaService) {
				testSvc.EXPECT().Login().Return()
				testSvc.EXPECT().InitOrganizations().Return()
				testSvc.EXPECT().ListAlertRules(mock.Anything).Return(nil, nil)
			},
		},
		{
			name: "ListingTest",
			validateFn: func(t *testing.T, output string) {
				assert.True(t, strings.Contains(output, "NAME"), "header NAME not found")
				assert.True(t, strings.Contains(output, "UID"), "header UID not found")
				assert.True(t, strings.Contains(output, "LABELS"), "header LABELS not found")
				assert.True(t, strings.Contains(output, "─────"), "table structure not found")
				assert.True(t, strings.Contains(output, "my-alert-rule"))
				assert.True(t, strings.Contains(output, "my-alert-uid"))
				assert.True(t, strings.Contains(output, "my-folder"))
				assert.True(t, strings.Contains(output, "my-group"))
				assert.True(t, strings.Contains(output, `{"env":"staging"}`))
			},
			setupMocks: func(testSvc *mocks.GrafanaService) {
				testSvc.EXPECT().Login().Return()
				testSvc.EXPECT().InitOrganizations().Return()
				rules := []*customModels.AlertRuleWithNestedFolder{
					{
						ProvisionedAlertRule: &models.ProvisionedAlertRule{
							Title:     new("my-alert-rule"),
							UID:       "my-alert-uid",
							RuleGroup: new("my-group"),
							Labels:    map[string]string{"env": "staging"},
						},
						NestedPath: "my-folder",
					},
				}
				testSvc.EXPECT().ListAlertRules(mock.Anything).Return(rules, nil)
			},
		},
		{
			name:      "ErrorTest",
			expectErr: false,
			validateFn: func(t *testing.T, output string) {
				assert.True(t, strings.Contains(output, `ERR unable to retrieve Orgs rule alerts err="failed to list alert rules"`))
			},
			setupMocks: func(testSvc *mocks.GrafanaService) {
				testSvc.EXPECT().Login().Return()
				testSvc.EXPECT().InitOrganizations().Return()
				testSvc.EXPECT().ListAlertRules(mock.Anything).Return(nil, fmt.Errorf("failed to list alert rules"))
			},
		},
	}

	for _, tc := range testCases {
		if tc.skip {
			slog.Debug("Skipping test", slog.Any("testName", tc.name))
			continue
		}
		t.Log("Running test", tc.name)
		testSvc := new(mocks.GrafanaService)
		if tc.setupMocks != nil {
			tc.setupMocks(testSvc)
		}
		optionMockSvc := GetOptionMockSvc(testSvc)
		r, w, cleanup := test_tooling.InterceptStdout()
		defer cleanup()

		rootSvc := cli.NewRootService()
		err := cli.Execute(rootSvc, listCmd, optionMockSvc())
		if tc.expectErr {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
		defer cleanup()
		assert.NoError(t, w.Close())

		out, _ := io.ReadAll(r)
		outStr := string(out)
		assert.NotNil(t, tc.validateFn)
		tc.validateFn(t, outStr)
	}
}

func TestDownloadAlertRules(t *testing.T) {
	listCmd := []string{"backup", "alerting", "rules", "download"}
	testCases := []struct {
		skip       bool
		name       string
		validateFn func(t *testing.T, output string)
		setupMocks func(testSvc *mocks.GrafanaService)
		expectErr  bool
	}{
		{
			name: "ErrNoDataTest",
			validateFn: func(t *testing.T, output string) {
				assert.True(t, strings.Contains(output, `ERR unable to retrieve Org's rule alerts err="failed to download alert rules"`))
			},
			setupMocks: func(testSvc *mocks.GrafanaService) {
				testSvc.EXPECT().Login().Return()
				testSvc.EXPECT().InitOrganizations().Return()
				testSvc.EXPECT().DownloadAlertRules(mock.Anything).Return(nil, fmt.Errorf("failed to download alert rules"))
			},
		},
		{
			name: "SuccessDownload",
			validateFn: func(t *testing.T, output string) {
				assert.True(t, strings.Contains(output, "alert-rule"))
				assert.True(t, strings.Contains(output, "─────"), "table structure not found")
				assert.True(t, strings.Contains(output, "rules/my-alert-rule.json"))
				assert.True(t, strings.Contains(output, "rules/my-other-rule.json"))
			},
			setupMocks: func(testSvc *mocks.GrafanaService) {
				testSvc.EXPECT().Login().Return()
				testSvc.EXPECT().InitOrganizations().Return()
				testSvc.EXPECT().DownloadAlertRules(mock.Anything).Return([]string{
					"rules/my-alert-rule.json",
					"rules/my-other-rule.json",
				}, nil)
			},
		},
	}

	for _, tc := range testCases {
		if tc.skip {
			slog.Debug("Skipping test", slog.Any("testName", tc.name))
			continue
		}
		t.Log("Running test", tc.name)
		testSvc := new(mocks.GrafanaService)
		if tc.setupMocks != nil {
			tc.setupMocks(testSvc)
		}
		optionMockSvc := GetOptionMockSvc(testSvc)
		r, w, cleanup := test_tooling.InterceptStdout()
		defer cleanup()

		rootSvc := cli.NewRootService()
		err := cli.Execute(rootSvc, listCmd, optionMockSvc())
		if tc.expectErr {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
		defer cleanup()
		assert.NoError(t, w.Close())

		out, _ := io.ReadAll(r)
		outStr := string(out)
		assert.NotNil(t, tc.validateFn)
		tc.validateFn(t, outStr)
	}
}

func TestUploadAlertRules(t *testing.T) {
	listCmd := []string{"backup", "alerting", "rules", "upload"}
	testCases := []struct {
		skip       bool
		name       string
		validateFn func(t *testing.T, output string)
		setupMocks func(testSvc *mocks.GrafanaService)
		expectErr  bool
	}{
		{
			name: "ErrUploadTest",
			validateFn: func(t *testing.T, output string) {
				assert.True(t, strings.Contains(output, `ERR unable to upload Org's rule alerts err="upload failed"`))
			},
			setupMocks: func(testSvc *mocks.GrafanaService) {
				testSvc.EXPECT().Login().Return()
				testSvc.EXPECT().InitOrganizations().Return()
				testSvc.EXPECT().UploadAlertRules(mock.Anything).Return(nil, fmt.Errorf("upload failed"))
			},
		},
		{
			name: "SuccessUpload",
			validateFn: func(t *testing.T, output string) {
				assert.True(t, strings.Contains(output, "Rules have been successfully uploaded to Grafana"))
				assert.True(t, strings.Contains(output, "NAME"), "header NAME not found")
				assert.True(t, strings.Contains(output, "UID"), "header UID not found")
				assert.True(t, strings.Contains(output, "LABELS"), "header LABELS not found")
				assert.True(t, strings.Contains(output, "─────"), "table structure not found")
				assert.True(t, strings.Contains(output, "my-alert-rule"))
				assert.True(t, strings.Contains(output, "my-alert-uid"))
				assert.True(t, strings.Contains(output, "my-folder"))
				assert.True(t, strings.Contains(output, "my-group"))
				assert.True(t, strings.Contains(output, `{"env":"staging"}`))
			},
			setupMocks: func(testSvc *mocks.GrafanaService) {
				testSvc.EXPECT().Login().Return()
				testSvc.EXPECT().InitOrganizations().Return()
				rules := []*customModels.AlertRuleWithNestedFolder{
					{
						ProvisionedAlertRule: &models.ProvisionedAlertRule{
							Title:     new("my-alert-rule"),
							UID:       "my-alert-uid",
							RuleGroup: new("my-group"),
							Labels:    map[string]string{"env": "staging"},
						},
						NestedPath: "my-folder",
					},
				}
				testSvc.EXPECT().UploadAlertRules(mock.Anything).Return(rules, nil)
			},
		},
	}

	for _, tc := range testCases {
		if tc.skip {
			slog.Debug("Skipping test", slog.Any("testName", tc.name))
			continue
		}
		t.Log("Running test", tc.name)
		testSvc := new(mocks.GrafanaService)
		if tc.setupMocks != nil {
			tc.setupMocks(testSvc)
		}
		optionMockSvc := GetOptionMockSvc(testSvc)
		r, w, cleanup := test_tooling.InterceptStdout()
		defer cleanup()

		rootSvc := cli.NewRootService()
		err := cli.Execute(rootSvc, listCmd, optionMockSvc())
		if tc.expectErr {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
		defer cleanup()
		assert.NoError(t, w.Close())

		out, _ := io.ReadAll(r)
		outStr := string(out)
		assert.NotNil(t, tc.validateFn)
		tc.validateFn(t, outStr)
	}
}

func TestClearAlertRules(t *testing.T) {
	clearCmd := []string{"backup", "alerting", "rules", "clear"}
	testCases := []struct {
		skip       bool
		name       string
		validateFn func(t *testing.T, output string)
		setupMocks func(testSvc *mocks.GrafanaService)
		expectErr  bool
	}{
		{
			name: "ErrClearTest",
			validateFn: func(t *testing.T, output string) {
				assert.True(t, strings.Contains(output, `ERR unable to deleting Org's rule alerts err="clear failed"`))
			},
			setupMocks: func(testSvc *mocks.GrafanaService) {
				testSvc.EXPECT().Login().Return()
				testSvc.EXPECT().InitOrganizations().Return()
				testSvc.EXPECT().ClearAlertRules(mock.Anything).Return(nil, fmt.Errorf("clear failed"))
			},
		},
		{
			name: "NoDataTest",
			validateFn: func(t *testing.T, output string) {
				assert.True(t, strings.Contains(output, "No Alerting rules were found"))
			},
			setupMocks: func(testSvc *mocks.GrafanaService) {
				testSvc.EXPECT().Login().Return()
				testSvc.EXPECT().InitOrganizations().Return()
				testSvc.EXPECT().ClearAlertRules(mock.Anything).Return(nil, nil)
			},
		},
		{
			name: "SuccessClearTest",
			validateFn: func(t *testing.T, output string) {
				assert.True(t, strings.Contains(output, "─────"), "table structure not found")
				assert.True(t, strings.Contains(output, "my-alert-rule"))
				assert.True(t, strings.Contains(output, "my-other-rule"))
			},
			setupMocks: func(testSvc *mocks.GrafanaService) {
				testSvc.EXPECT().Login().Return()
				testSvc.EXPECT().InitOrganizations().Return()
				testSvc.EXPECT().ClearAlertRules(mock.Anything).Return([]string{
					"my-alert-rule",
					"my-other-rule",
				}, nil)
			},
		},
	}

	for _, tc := range testCases {
		if tc.skip {
			slog.Debug("Skipping test", slog.Any("testName", tc.name))
			continue
		}
		t.Log("Running test", tc.name)
		testSvc := new(mocks.GrafanaService)
		if tc.setupMocks != nil {
			tc.setupMocks(testSvc)
		}
		optionMockSvc := GetOptionMockSvc(testSvc)
		r, w, cleanup := test_tooling.InterceptStdout()
		defer cleanup()

		rootSvc := cli.NewRootService()
		err := cli.Execute(rootSvc, clearCmd, optionMockSvc())
		if tc.expectErr {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
		defer cleanup()
		assert.NoError(t, w.Close())

		out, _ := io.ReadAll(r)
		outStr := string(out)
		assert.NotNil(t, tc.validateFn)
		tc.validateFn(t, outStr)
	}
}
