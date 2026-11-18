package backup_test

import (
	"fmt"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/esnet/gdg/cli"
	"github.com/esnet/gdg/internal/domain"
	"github.com/esnet/gdg/internal/ports/outbound/mocks"
	"github.com/esnet/gdg/pkg/test_tooling"
	"github.com/stretchr/testify/assert"
)

func routingTreeResource(name, receiver string, routeCount int) map[string]any {
	routes := make([]any, routeCount)
	for i := range routes {
		routes[i] = map[string]any{"receiver": fmt.Sprintf("r-%d", i)}
	}
	return map[string]any{
		"kind":       "RoutingTree",
		"apiVersion": "notifications.alerting.grafana.app/v1beta1",
		"metadata": map[string]any{
			"name":      name,
			"namespace": "default",
		},
		"spec": map[string]any{
			"defaults": map[string]any{"receiver": receiver},
			"routes":   routes,
		},
	}
}

func TestListRoutingTrees(t *testing.T) {
	listCmd := []string{"backup", "alerting", "routingtree", "list"}
	testCases := []struct {
		skip       bool
		name       string
		validateFn func(t *testing.T, output string)
		setupMocks func(testSvc *mocks.GrafanaService)
		expectErr  bool
	}{
		{
			name: "ErrDataTest",
			validateFn: func(t *testing.T, output string) {
				assert.True(t, strings.Contains(output, "ERR unable to list routing trees"))
			},
			setupMocks: func(testSvc *mocks.GrafanaService) {
				testSvc.EXPECT().Login().Return()
				testSvc.EXPECT().InitOrganizations().Return()
				testSvc.EXPECT().ListRoutingTrees("").Return(nil, fmt.Errorf("boom"))
			},
		},
		{
			name: "NoDataTest",
			validateFn: func(t *testing.T, output string) {
				assert.True(t, strings.Contains(output, "No routing trees found"))
			},
			setupMocks: func(testSvc *mocks.GrafanaService) {
				testSvc.EXPECT().Login().Return()
				testSvc.EXPECT().InitOrganizations().Return()
				testSvc.EXPECT().ListRoutingTrees("").Return(nil, nil)
			},
		},
		{
			name: "ListingTest",
			validateFn: func(t *testing.T, output string) {
				assert.True(t, strings.Contains(output, "moo"))
				assert.True(t, strings.Contains(output, "user-defined"))
				assert.True(t, strings.Contains(output, "devnull"))
				assert.True(t, strings.Contains(output, "grafana-default-email"))
			},
			setupMocks: func(testSvc *mocks.GrafanaService) {
				testSvc.EXPECT().Login().Return()
				testSvc.EXPECT().InitOrganizations().Return()
				resp := []*domain.RoutingTreeGdg{
					{Resource: routingTreeResource("moo", "devnull", 1), GdgApiVersion: domain.GdgRoutingTreeApiVersion},
					{Resource: routingTreeResource("user-defined", "grafana-default-email", 3), GdgApiVersion: domain.GdgRoutingTreeApiVersion},
				}
				testSvc.EXPECT().ListRoutingTrees("").Return(resp, nil)
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

func TestListRoutingTreesWithNameFilter(t *testing.T) {
	listCmd := []string{"backup", "alerting", "routingtree", "list", "--name", "user-defined"}
	testSvc := new(mocks.GrafanaService)
	testSvc.EXPECT().Login().Return()
	testSvc.EXPECT().InitOrganizations().Return()
	resp := []*domain.RoutingTreeGdg{
		{Resource: routingTreeResource("user-defined", "grafana-default-email", 3), GdgApiVersion: domain.GdgRoutingTreeApiVersion},
	}
	testSvc.EXPECT().ListRoutingTrees("user-defined").Return(resp, nil)

	optionMockSvc := GetOptionMockSvc(testSvc)
	r, w, cleanup := test_tooling.InterceptStdout()
	defer cleanup()

	rootSvc := cli.NewRootService()
	err := cli.Execute(rootSvc, listCmd, optionMockSvc())
	assert.Nil(t, err)
	assert.NoError(t, w.Close())

	out, _ := io.ReadAll(r)
	assert.Contains(t, string(out), "user-defined")
}

func TestDownloadRoutingTrees(t *testing.T) {
	downloadCmd := []string{"backup", "alerting", "routingtree", "download"}
	testCases := []struct {
		skip       bool
		name       string
		validateFn func(t *testing.T, output string)
		setupMocks func(testSvc *mocks.GrafanaService)
	}{
		{
			name: "ErrDataTest",
			validateFn: func(t *testing.T, output string) {
				assert.True(t, strings.Contains(output, "ERR unable to download routing trees"))
			},
			setupMocks: func(testSvc *mocks.GrafanaService) {
				testSvc.EXPECT().Login().Return()
				testSvc.EXPECT().InitOrganizations().Return()
				testSvc.EXPECT().DownloadRoutingTrees("").Return(nil, fmt.Errorf("boom"))
			},
		},
		{
			name: "SuccessDownload",
			validateFn: func(t *testing.T, output string) {
				assert.True(t, strings.Contains(output, "moo.json"))
			},
			setupMocks: func(testSvc *mocks.GrafanaService) {
				testSvc.EXPECT().Login().Return()
				testSvc.EXPECT().InitOrganizations().Return()
				testSvc.EXPECT().DownloadRoutingTrees("").Return([]string{"alerting-routing-trees/moo.json"}, nil)
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

		rootSvc := cli.NewRootService()
		err := cli.Execute(rootSvc, downloadCmd, optionMockSvc())
		assert.Nil(t, err)
		defer cleanup()
		assert.NoError(t, w.Close())

		out, _ := io.ReadAll(r)
		outStr := string(out)
		assert.NotNil(t, tc.validateFn)
		tc.validateFn(t, outStr)
	}
}

func TestClearRoutingTrees(t *testing.T) {
	clearCmd := []string{"backup", "alerting", "routingtree", "clear"}
	testCases := []struct {
		skip       bool
		name       string
		validateFn func(t *testing.T, output string)
		setupMocks func(testSvc *mocks.GrafanaService)
	}{
		{
			name: "ErrDataTest",
			validateFn: func(t *testing.T, output string) {
				assert.True(t, strings.Contains(output, "ERR unable to clear routing trees"))
			},
			setupMocks: func(testSvc *mocks.GrafanaService) {
				testSvc.EXPECT().Login().Return()
				testSvc.EXPECT().InitOrganizations().Return()
				testSvc.EXPECT().ClearRoutingTrees("").Return(nil, fmt.Errorf("boom"))
			},
		},
		{
			name: "NoDataTest",
			validateFn: func(t *testing.T, output string) {
				assert.True(t, strings.Contains(output, "No routing trees found"))
			},
			setupMocks: func(testSvc *mocks.GrafanaService) {
				testSvc.EXPECT().Login().Return()
				testSvc.EXPECT().InitOrganizations().Return()
				testSvc.EXPECT().ClearRoutingTrees("").Return(nil, nil)
			},
		},
		{
			name: "SuccessClear",
			validateFn: func(t *testing.T, output string) {
				assert.True(t, strings.Contains(output, "moo"))
				assert.True(t, strings.Contains(output, "user-defined"))
			},
			setupMocks: func(testSvc *mocks.GrafanaService) {
				testSvc.EXPECT().Login().Return()
				testSvc.EXPECT().InitOrganizations().Return()
				testSvc.EXPECT().ClearRoutingTrees("").Return([]string{"moo", "user-defined"}, nil)
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

		rootSvc := cli.NewRootService()
		err := cli.Execute(rootSvc, clearCmd, optionMockSvc())
		assert.Nil(t, err)
		defer cleanup()
		assert.NoError(t, w.Close())

		out, _ := io.ReadAll(r)
		outStr := string(out)
		assert.NotNil(t, tc.validateFn)
		tc.validateFn(t, outStr)
	}
}

func TestClearRoutingTreesWithNameFilter(t *testing.T) {
	clearCmd := []string{"backup", "alerting", "routingtree", "clear", "--name", "moo"}
	testSvc := new(mocks.GrafanaService)
	testSvc.EXPECT().Login().Return()
	testSvc.EXPECT().InitOrganizations().Return()
	testSvc.EXPECT().ClearRoutingTrees("moo").Return([]string{"moo"}, nil)

	optionMockSvc := GetOptionMockSvc(testSvc)
	r, w, cleanup := test_tooling.InterceptStdout()
	defer cleanup()

	rootSvc := cli.NewRootService()
	err := cli.Execute(rootSvc, clearCmd, optionMockSvc())
	assert.Nil(t, err)
	assert.NoError(t, w.Close())

	out, _ := io.ReadAll(r)
	assert.Contains(t, string(out), "moo")
}

func TestUploadRoutingTrees(t *testing.T) {
	uploadCmd := []string{"backup", "alerting", "routingtree", "upload"}
	testCases := []struct {
		skip       bool
		name       string
		validateFn func(t *testing.T, output string)
		setupMocks func(testSvc *mocks.GrafanaService)
	}{
		{
			name: "ErrDataTest",
			validateFn: func(t *testing.T, output string) {
				assert.True(t, strings.Contains(output, "ERR unable to upload routing trees"))
			},
			setupMocks: func(testSvc *mocks.GrafanaService) {
				testSvc.EXPECT().Login().Return()
				testSvc.EXPECT().InitOrganizations().Return()
				testSvc.EXPECT().UploadRoutingTrees("").Return(nil, fmt.Errorf("boom"))
			},
		},
		{
			name: "NoDataTest",
			validateFn: func(t *testing.T, output string) {
				assert.True(t, strings.Contains(output, "No routing trees found"))
			},
			setupMocks: func(testSvc *mocks.GrafanaService) {
				testSvc.EXPECT().Login().Return()
				testSvc.EXPECT().InitOrganizations().Return()
				testSvc.EXPECT().UploadRoutingTrees("").Return(nil, nil)
			},
		},
		{
			name: "SuccessUpload",
			validateFn: func(t *testing.T, output string) {
				assert.True(t, strings.Contains(output, "moo"))
			},
			setupMocks: func(testSvc *mocks.GrafanaService) {
				testSvc.EXPECT().Login().Return()
				testSvc.EXPECT().InitOrganizations().Return()
				testSvc.EXPECT().UploadRoutingTrees("").Return([]string{"moo"}, nil)
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

		rootSvc := cli.NewRootService()
		err := cli.Execute(rootSvc, uploadCmd, optionMockSvc())
		assert.Nil(t, err)
		defer cleanup()
		assert.NoError(t, w.Close())

		out, _ := io.ReadAll(r)
		outStr := string(out)
		assert.NotNil(t, tc.validateFn)
		tc.validateFn(t, outStr)
	}
}

// TestUploadRoutingTreesWarnsAboutPruning verifies the upload command always
// surfaces a warning that it will remove server-side routing trees with no
// local counterpart, since upload makes the server fully reflect local state.
func TestUploadRoutingTreesWarnsAboutPruning(t *testing.T) {
	uploadCmd := []string{"backup", "alerting", "routingtree", "upload"}
	testSvc := new(mocks.GrafanaService)
	testSvc.EXPECT().Login().Return()
	testSvc.EXPECT().InitOrganizations().Return()
	testSvc.EXPECT().UploadRoutingTrees("").Return([]string{"moo"}, nil)

	optionMockSvc := GetOptionMockSvc(testSvc)
	r, w, cleanup := test_tooling.InterceptStdout()
	defer cleanup()

	rootSvc := cli.NewRootService()
	err := cli.Execute(rootSvc, uploadCmd, optionMockSvc())
	assert.Nil(t, err)
	assert.NoError(t, w.Close())

	out, _ := io.ReadAll(r)
	outStr := string(out)
	assert.True(t, strings.Contains(outStr, "WRN"))
	assert.True(t, strings.Contains(outStr, "will be deleted"))
}

func TestUploadRoutingTreesWithNameFilter(t *testing.T) {
	uploadCmd := []string{"backup", "alerting", "routingtree", "upload", "--name", "moo"}
	testSvc := new(mocks.GrafanaService)
	testSvc.EXPECT().Login().Return()
	testSvc.EXPECT().InitOrganizations().Return()
	testSvc.EXPECT().UploadRoutingTrees("moo").Return([]string{"moo"}, nil)

	optionMockSvc := GetOptionMockSvc(testSvc)
	r, w, cleanup := test_tooling.InterceptStdout()
	defer cleanup()

	rootSvc := cli.NewRootService()
	err := cli.Execute(rootSvc, uploadCmd, optionMockSvc())
	assert.Nil(t, err)
	assert.NoError(t, w.Close())

	out, _ := io.ReadAll(r)
	assert.Contains(t, string(out), "moo")
}
