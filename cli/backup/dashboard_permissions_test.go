package backup_test

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/grafana/grafana-openapi-client-go/models"

	"github.com/esnet/gdg/cli"
	"github.com/esnet/gdg/internal/domain"
	"github.com/esnet/gdg/internal/ports/outbound/mocks"
	"github.com/esnet/gdg/pkg/test_tooling"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ── list ─────────────────────────────────────────────────────────────────────

func TestListDashboardPermissions(t *testing.T) {
	listCmd := []string{"backup", "dashboards", "permission", "list"}
	testCases := []struct {
		name       string
		validateFn func(t *testing.T, output string)
		setupMocks func(testSvc *mocks.GrafanaService)
	}{
		{
			name: "NoDataTest",
			validateFn: func(t *testing.T, output string) {
				assert.True(t, strings.Contains(output, "No Dashboards found"))
			},
			setupMocks: func(testSvc *mocks.GrafanaService) {
				testSvc.EXPECT().Login().Return()
				testSvc.EXPECT().InitOrganizations().Return()
				testSvc.EXPECT().ListDashboardPermissions(mock.Anything).Return(nil, nil)
			},
		},
		{
			name: "ListingTest",
			validateFn: func(t *testing.T, output string) {
				assert.True(t, strings.Contains(output, "dashboard-uid"))
				assert.True(t, strings.Contains(output, "My Dashboard"))
				assert.True(t, strings.Contains(output, "user:alice"))
				assert.True(t, strings.Contains(output, "Editor"))
			},
			setupMocks: func(testSvc *mocks.GrafanaService) {
				testSvc.EXPECT().Login().Return()
				testSvc.EXPECT().InitOrganizations().Return()
				resp := []domain.DashboardAndPermissions{
					{
						Dashboard: &domain.NestedHit{
							Hit: &models.Hit{
								ID:    1,
								UID:   "dashboard-uid",
								Title: "My Dashboard",
								Slug:  "my-dashboard",
							},
							NestedPath: "General",
						},
						Permissions: []*models.ResourcePermissionDTO{
							{
								UserLogin:  "alice",
								Permission: "Editor",
							},
						},
					},
				}
				testSvc.EXPECT().ListDashboardPermissions(mock.Anything).Return(resp, nil)
			},
		},
		{
			name: "ServiceAccountPermissionTest",
			validateFn: func(t *testing.T, output string) {
				assert.True(t, strings.Contains(output, "service:svc-account"))
			},
			setupMocks: func(testSvc *mocks.GrafanaService) {
				testSvc.EXPECT().Login().Return()
				testSvc.EXPECT().InitOrganizations().Return()
				resp := []domain.DashboardAndPermissions{
					{
						Dashboard: &domain.NestedHit{
							Hit:        &models.Hit{ID: 2, UID: "dashboard-uid-2", Title: "Another Dashboard"},
							NestedPath: "General",
						},
						Permissions: []*models.ResourcePermissionDTO{
							{
								UserLogin:        "svc-account",
								IsServiceAccount: true,
								Permission:       "Viewer",
							},
						},
					},
				}
				testSvc.EXPECT().ListDashboardPermissions(mock.Anything).Return(resp, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			slog.Info("Running test", slog.Any("testName", tc.name))
			testSvc := new(mocks.GrafanaService)
			if tc.setupMocks != nil {
				tc.setupMocks(testSvc)
			}
			optionMockSvc := GetOptionMockSvc(testSvc)
			r, w, cleanup := test_tooling.InterceptStdout()
			defer cleanup()

			rootSvc := cli.NewRootService()
			err := cli.Execute(rootSvc, listCmd, optionMockSvc())
			assert.Nil(t, err)
			assert.NoError(t, w.Close())

			out, _ := io.ReadAll(r)
			outStr := string(out)
			assert.NotNil(t, tc.validateFn)
			tc.validateFn(t, outStr)
		})
	}
}

// ── download ─────────────────────────────────────────────────────────────────

func TestDownloadDashboardPermissions(t *testing.T) {
	downloadCmd := []string{"backup", "dashboards", "permission", "download"}
	testCases := []struct {
		name       string
		validateFn func(t *testing.T, output string)
		setupMocks func(testSvc *mocks.GrafanaService)
	}{
		{
			name: "NoDataTest",
			validateFn: func(t *testing.T, output string) {
				assert.True(t, strings.Contains(output, "No Dashboard permissions"))
			},
			setupMocks: func(testSvc *mocks.GrafanaService) {
				testSvc.EXPECT().Login().Return()
				testSvc.EXPECT().InitOrganizations().Return()
				testSvc.EXPECT().DownloadDashboardPermissions(mock.Anything).Return(nil, nil)
			},
		},
		{
			name: "SuccessDownload",
			validateFn: func(t *testing.T, output string) {
				assert.True(t, strings.Contains(output, "permissions_file.json"))
			},
			setupMocks: func(testSvc *mocks.GrafanaService) {
				testSvc.EXPECT().Login().Return()
				testSvc.EXPECT().InitOrganizations().Return()
				testSvc.EXPECT().DownloadDashboardPermissions(mock.Anything).Return([]string{"permissions_file.json"}, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			slog.Info("Running test", slog.Any("testName", tc.name))
			testSvc := new(mocks.GrafanaService)
			if tc.setupMocks != nil {
				tc.setupMocks(testSvc)
			}
			optionMockSvc := GetOptionMockSvc(testSvc)
			r, w, cleanup := test_tooling.InterceptStdout()
			defer cleanup()

			rootSvc := cli.NewRootService()
			err := cli.Execute(rootSvc, downloadCmd, optionMockSvc())
			assert.Nil(t, err)
			assert.NoError(t, w.Close())

			out, _ := io.ReadAll(r)
			outStr := string(out)
			assert.NotNil(t, tc.validateFn)
			tc.validateFn(t, outStr)
		})
	}
}

// ── upload ───────────────────────────────────────────────────────────────────

func TestUploadDashboardPermissions(t *testing.T) {
	uploadCmd := []string{"backup", "dashboards", "permission", "upload"}
	testCases := []struct {
		name       string
		validateFn func(t *testing.T, output string)
		setupMocks func(testSvc *mocks.GrafanaService)
	}{
		{
			name: "NoDataTest",
			validateFn: func(t *testing.T, output string) {
				assert.True(t, strings.Contains(output, "No permissions found"))
			},
			setupMocks: func(testSvc *mocks.GrafanaService) {
				testSvc.EXPECT().Login().Return()
				testSvc.EXPECT().InitOrganizations().Return()
				testSvc.EXPECT().UploadDashboardPermissions(mock.Anything).Return(nil, nil)
			},
		},
		{
			name: "SuccessUpload",
			validateFn: func(t *testing.T, output string) {
				assert.True(t, strings.Contains(output, "dashboard-uid"))
			},
			setupMocks: func(testSvc *mocks.GrafanaService) {
				testSvc.EXPECT().Login().Return()
				testSvc.EXPECT().InitOrganizations().Return()
				testSvc.EXPECT().UploadDashboardPermissions(mock.Anything).Return([]string{"dashboard-uid"}, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			slog.Info("Running test", slog.Any("testName", tc.name))
			testSvc := new(mocks.GrafanaService)
			if tc.setupMocks != nil {
				tc.setupMocks(testSvc)
			}
			optionMockSvc := GetOptionMockSvc(testSvc)
			r, w, cleanup := test_tooling.InterceptStdout()
			defer cleanup()

			rootSvc := cli.NewRootService()
			err := cli.Execute(rootSvc, uploadCmd, optionMockSvc())
			assert.Nil(t, err)
			assert.NoError(t, w.Close())

			out, _ := io.ReadAll(r)
			outStr := string(out)
			assert.NotNil(t, tc.validateFn)
			tc.validateFn(t, outStr)
		})
	}
}

// ── clear ────────────────────────────────────────────────────────────────────

// TestClearDashboardPermissions exercises the "clear" subcommand without
// --skip-confirmation, which blocks on tools.GetUserConfirmation reading
// os.Stdin. We redirect stdin via os.Pipe and feed "y\n" so the confirmation
// resolves without hanging the test process — the same pattern used in
// pkg/tools/prompt_helpers_test.go.
func TestClearDashboardPermissions(t *testing.T) {
	clearCmd := []string{"backup", "dashboards", "permission", "clear"}
	testCases := []struct {
		name       string
		validateFn func(t *testing.T, output string)
		setupMocks func(testSvc *mocks.GrafanaService)
	}{
		{
			name: "SuccessClear",
			validateFn: func(t *testing.T, output string) {
				assert.True(t, strings.Contains(output, "All dashboard permissions have been cleared"))
			},
			setupMocks: func(testSvc *mocks.GrafanaService) {
				testSvc.EXPECT().Login().Return()
				testSvc.EXPECT().InitOrganizations().Return()
				testSvc.EXPECT().ClearDashboardPermissions(mock.Anything).Return(nil)
			},
		},
		{
			name: "ErrorClear",
			validateFn: func(t *testing.T, output string) {
				assert.True(t, strings.Contains(output, "Failed to retrieve Dashboard Permissions"))
			},
			setupMocks: func(testSvc *mocks.GrafanaService) {
				testSvc.EXPECT().Login().Return()
				testSvc.EXPECT().InitOrganizations().Return()
				testSvc.EXPECT().ClearDashboardPermissions(mock.Anything).Return(fmt.Errorf("boom"))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			slog.Info("Running test", slog.Any("testName", tc.name))
			testSvc := new(mocks.GrafanaService)
			if tc.setupMocks != nil {
				tc.setupMocks(testSvc)
			}
			optionMockSvc := GetOptionMockSvc(testSvc)

			// Redirect stdin: GetUserConfirmation(..., terminate=true) will
			// call log.Fatal on anything but "y", killing the test process.
			r, w, _ := os.Pipe()
			origStdin := os.Stdin
			os.Stdin = r
			defer func() { os.Stdin = origStdin }()
			_, _ = io.WriteString(w, "y\n")
			w.Close()

			outR, outW, cleanup := test_tooling.InterceptStdout()
			defer cleanup()

			rootSvc := cli.NewRootService()
			err := cli.Execute(rootSvc, clearCmd, optionMockSvc())
			assert.Nil(t, err)
			assert.NoError(t, outW.Close())

			out, _ := io.ReadAll(outR)
			outStr := string(out)
			assert.NotNil(t, tc.validateFn)
			tc.validateFn(t, outStr)
		})
	}
}

// TestClearDashboardPermissions_SkipConfirmation verifies that --skip-confirmation
// (a PersistentFlag on the parent `dashboards` command, shared with
// `dashboards clear`) bypasses the GetUserConfirmation stdin prompt entirely —
// no stdin redirection is needed here, and if the flag were not honored this
// test would hang/fail on the blocking stdin read.
func TestClearDashboardPermissions_SkipConfirmation(t *testing.T) {
	clearCmd := []string{"backup", "dashboards", "permission", "clear", "--skip-confirmation"}

	testSvc := new(mocks.GrafanaService)
	testSvc.EXPECT().Login().Return()
	testSvc.EXPECT().InitOrganizations().Return()
	testSvc.EXPECT().ClearDashboardPermissions(mock.Anything).Return(nil)
	optionMockSvc := GetOptionMockSvc(testSvc)

	outR, outW, cleanup := test_tooling.InterceptStdout()
	defer cleanup()

	rootSvc := cli.NewRootService()
	err := cli.Execute(rootSvc, clearCmd, optionMockSvc())
	assert.Nil(t, err)
	assert.NoError(t, outW.Close())

	out, _ := io.ReadAll(outR)
	assert.True(t, strings.Contains(string(out), "All dashboard permissions have been cleared"))
}
