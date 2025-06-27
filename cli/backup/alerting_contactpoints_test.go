package backup_test

import (
	"fmt"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/esnet/gdg/internal/tools/ptr"
	"github.com/grafana/grafana-openapi-client-go/models"

	"github.com/esnet/gdg/cli"
	"github.com/esnet/gdg/internal/service/mocks"
	"github.com/esnet/gdg/pkg/test_tooling"
	"github.com/stretchr/testify/assert"
)

func TestUploadContactPoints(t *testing.T) {
	listCmd := []string{"backup", "alerting", "contactpoint", "upload"}
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
				assert.True(t, strings.Contains(output, "WRN GDG does not manage the 'email receiver' entity."))
				assert.True(t, strings.Contains(output, "ERR unable to upload contact points err=\"Unable to download data data\""))
			},
			setupMocks: func(testSvc *mocks.GrafanaService) {
				testSvc.EXPECT().InitOrganizations().Return()
				testSvc.EXPECT().UploadContactPoints().Return(nil, fmt.Errorf("Unable to download data data"))
			},
		},
		{
			name: "SuccessUpload",
			validateFn: func(t *testing.T, output string) {
				assert.True(t, strings.Contains(output, "WRN GDG does not manage the 'email receiver' entity."))
				assert.True(t, strings.Contains(output, "NAME"))
				assert.True(t, strings.Contains(output, "─────"), "table structure not found")
				assert.True(t, strings.Contains(output, "discord"))
				assert.True(t, strings.Contains(output, "slack"))
			},
			setupMocks: func(testSvc *mocks.GrafanaService) {
				testSvc.EXPECT().InitOrganizations().Return()
				testSvc.EXPECT().UploadContactPoints().Return([]string{"discord", "slack"}, nil)
			},
		},
	}
	for _, tc := range testCases {
		if tc.skip {
			slog.Debug("Skipping test", slog.Any("testName", tc.name))
			continue
		}
		slog.Info("Running test", slog.Any("testName", tc.name))
		testSvc := new(mocks.GrafanaService)
		if tc.setupMocks != nil {
			tc.setupMocks(testSvc)
		}
		optionMockSvc := GetOptionMockSvc(testSvc)
		r, w, cleanup := test_tooling.InterceptStdout()
		defer cleanup()

		err := cli.Execute(listCmd, optionMockSvc())
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

func TestDownloadContactPoints(t *testing.T) {
	listCmd := []string{"backup", "alerting", "contactpoint", "download"}
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
				assert.True(t, strings.Contains(output, "WRN GDG does not manage the 'email receiver' entity."))
				assert.True(t, strings.Contains(output, "ERR unable to download contact points"))
			},
			setupMocks: func(testSvc *mocks.GrafanaService) {
				testSvc.EXPECT().InitOrganizations().Return()
				testSvc.EXPECT().DownloadContactPoints().Return("", fmt.Errorf("Unable to download data data"))
			},
		},
		{
			name: "SuccessDownload",
			validateFn: func(t *testing.T, output string) {
				assert.True(t, strings.Contains(output, "WRN GDG does not manage the 'email receiver' entity."))
				assert.True(t, strings.Contains(output, "INF contact points successfully downloaded file=fileName"))
			},
			setupMocks: func(testSvc *mocks.GrafanaService) {
				testSvc.EXPECT().InitOrganizations().Return()
				testSvc.EXPECT().DownloadContactPoints().Return("fileName", nil)
			},
		},
	}
	for _, tc := range testCases {
		if tc.skip {
			slog.Debug("Skipping test", slog.Any("testName", tc.name))
			continue
		}
		slog.Info("Running test", slog.Any("testName", tc.name))
		testSvc := new(mocks.GrafanaService)
		if tc.setupMocks != nil {
			tc.setupMocks(testSvc)
		}
		optionMockSvc := GetOptionMockSvc(testSvc)
		r, w, cleanup := test_tooling.InterceptStdout()
		defer cleanup()

		err := cli.Execute(listCmd, optionMockSvc())
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

func TestListContactPoints(t *testing.T) {
	listCmd := []string{"backup", "alerting", "contactpoint", "list"}
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
				assert.True(t, strings.Contains(output, "WRN GDG does not manage the 'email receiver' entity."))
				assert.True(t, strings.Contains(output, "No contact points found"))
			},
			setupMocks: func(testSvc *mocks.GrafanaService) {
				testSvc.EXPECT().InitOrganizations().Return()
				testSvc.EXPECT().ListContactPoints().Return(nil, nil)
			},
		},
		{
			name: "ListingTest",
			validateFn: func(t *testing.T, output string) {
				assert.True(t, strings.Contains(output, "WRN GDG does not manage the 'email receiver' entity."))
				assert.True(t, strings.Contains(output, "discordUid"))
				assert.True(t, strings.Contains(output, "slackUid"))
				assert.True(t, strings.Contains(output, "Discord"))
				assert.True(t, strings.Contains(output, "Slack"))
				// validate Type
				assert.True(t, strings.Contains(output, "discordType"))
				assert.True(t, strings.Contains(output, "slackType"))
			},
			setupMocks: func(testSvc *mocks.GrafanaService) {
				testSvc.EXPECT().InitOrganizations().Return()
				resp := []*models.EmbeddedContactPoint{
					{
						UID:      "discordUid",
						Name:     "Discord",
						Type:     ptr.Of("discordType"),
						Settings: map[string]any{"token": "secret", "someValue": "result"},
					},
					{
						UID:      "slackUid",
						Name:     "Slack",
						Type:     ptr.Of("slackType"),
						Settings: map[string]any{"token": "secret", "slack": "rocks"},
					},
				}

				testSvc.EXPECT().ListContactPoints().Return(resp, nil)
			},
		},
	}
	for _, tc := range testCases {
		if tc.skip {
			slog.Debug("Skipping test", slog.Any("testName", tc.name))
			continue
		}
		slog.Info("Running test", slog.Any("testName", tc.name))
		testSvc := new(mocks.GrafanaService)
		if tc.setupMocks != nil {
			tc.setupMocks(testSvc)
		}
		optionMockSvc := GetOptionMockSvc(testSvc)
		r, w, cleanup := test_tooling.InterceptStdout()
		defer cleanup()

		err := cli.Execute(listCmd, optionMockSvc())
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

func TestClearContactPoints(t *testing.T) {
	clearCmd := []string{"backup", "alerting", "contactpoint", "clear"}
	testCases := []struct {
		skip       bool
		name       string
		validateFn func(t *testing.T, output string)
		setupMocks func(testSvc *mocks.GrafanaService)
	}{
		{
			name: "ErrDataTest",
			validateFn: func(t *testing.T, output string) {
				assert.True(t, strings.Contains(output, "WRN GDG does not manage the 'email receiver' entity."))
				assert.True(t, strings.Contains(output, "ERR unable to clear Contact Points"))
			},
			setupMocks: func(testSvc *mocks.GrafanaService) {
				testSvc.EXPECT().InitOrganizations().Return()
				testSvc.EXPECT().ClearContactPoints().Return(nil, fmt.Errorf("Errror!!!"))
			},
		},
		{
			name: "NoDataTest",
			validateFn: func(t *testing.T, output string) {
				assert.True(t, strings.Contains(output, "WRN GDG does not manage the 'email receiver' entity."))
				assert.True(t, strings.Contains(output, "Contact Points successfully removed"))
			},
			setupMocks: func(testSvc *mocks.GrafanaService) {
				testSvc.EXPECT().InitOrganizations().Return()
				testSvc.EXPECT().ClearContactPoints().Return(nil, nil)
			},
		},
		{
			name: "ListingTest",
			validateFn: func(t *testing.T, output string) {
				assert.True(t, strings.Contains(output, "WRN GDG does not manage the 'email receiver' entity."))
				assert.True(t, strings.Contains(output, "Contact Points successfully removed"))
			},
			setupMocks: func(testSvc *mocks.GrafanaService) {
				testSvc.EXPECT().ClearContactPoints().Return([]string{"discord", "slack"}, nil)
				testSvc.EXPECT().InitOrganizations().Return()
			},
		},
	}

	for _, tc := range testCases {
		if tc.skip {
			slog.Debug("Skipping test", slog.Any("testName", tc.name))
			continue
		}
		slog.Info("Running test", slog.Any("testName", tc.name))
		testSvc := new(mocks.GrafanaService)
		if tc.setupMocks != nil {
			tc.setupMocks(testSvc)
		}

		optionMockSvc := GetOptionMockSvc(testSvc)
		r, w, cleanup := test_tooling.InterceptStdout()
		defer cleanup()

		err := cli.Execute(clearCmd, optionMockSvc())
		assert.Nil(t, err)
		defer cleanup()
		assert.NoError(t, w.Close())

		out, _ := io.ReadAll(r)
		outStr := string(out)
		assert.NotNil(t, tc.validateFn)
		tc.validateFn(t, outStr)
	}
}
