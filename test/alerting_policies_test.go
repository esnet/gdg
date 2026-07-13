package test

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/samber/lo"

	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/pkg/test_tooling"
	"github.com/esnet/gdg/pkg/test_tooling/common"
	"github.com/esnet/gdg/pkg/test_tooling/path"
	"github.com/stretchr/testify/assert"
)

func TestPoliciesCrud(t *testing.T) {
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
		cleanupErr := r.CleanUp()
		if cleanupErr != nil {
			slog.Warn("Unable to clean up after test", "test", t.Name())
		}
	}()
	// Grafana v13 tightened validation in PUT /v1/provisioning/policies: the built-in
	// "grafana-default-email" system receiver is no longer accepted as a valid receiver
	// in the policy tree for any auth mode. This test is covered by the v12+token CI
	// matrix run. v13 coverage is provided by TestPoliciesCrudV13.
	if isV13() {
		t.Skip("Grafana v13: grafana-default-email is no longer a valid provisioned " +
			"receiver in PUT /v1/provisioning/policies. See TestPoliciesCrudV13.")
	}
	apiClient := r.ApiClient
	// Upload Contact points first.
	_, err = apiClient.UploadContactPoints()
	assert.NoError(t, err)
	policies, err := apiClient.ListAlertNotifications()
	assert.NoError(t, err)
	assert.Equal(t, len(policies.Routes), 0, "Validate initial contact list is empty")
	policiesListing, err := apiClient.UploadAlertNotifications()
	assert.NoError(t, err)
	assert.Equal(t, len(policiesListing.Routes), 2)
	route := lo.FindOrElse(policiesListing.Routes, nil, func(item *models.Route) bool {
		return item.Receiver == "slack"
	})
	assert.NotNil(t, route)
	assert.Equal(t, len(route.ObjectMatchers[0]), 3)
	assert.Equal(t, route.ObjectMatchers[0][2], "23")

	policies, err = apiClient.ListAlertNotifications()
	assert.NoError(t, err)
	assert.Equal(t, len(policies.Routes), 2)
	data, err := apiClient.DownloadAlertNotifications()
	assert.NoError(t, err)
	assert.Equal(t, "test/data/org_main-org/alerting/policies.json", data)
	rawData, err := os.ReadFile(data)
	assert.NoError(t, err)
	assert.True(t, bytes.Contains(rawData, []byte("grafana_folder")))
	assert.True(t, bytes.Contains(rawData, []byte("alertname")))
	err = apiClient.ClearAlertNotifications()
	assert.NoError(t, err)
	policies, err = apiClient.ListAlertNotifications()
	assert.NoError(t, err)
	assert.Equal(t, len(policies.Routes), 0)
}

// TestPoliciesCrudV13 exercises the full policy CRUD workflow on Grafana v13+.
// It uses only explicitly provisioned receivers (slack, discord) because v13
// no longer accepts the built-in "grafana-default-email" system receiver in
// PUT /v1/provisioning/policies. The policy JSON is written to an isolated temp
// directory so no fixture files are ever modified.
func TestPoliciesCrudV13(t *testing.T) {
	skipLegacyVersion(t)

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
		cleanupErr := r.CleanUp()
		if cleanupErr != nil {
			slog.Warn("Unable to clean up after test", "test", t.Name())
		}
	}()

	// Upload contact points using the original client (reads contacts.json from test/data/).
	apiClient := r.ApiClient
	_, err = apiClient.UploadContactPoints()
	assert.NoError(t, err)

	// Build a v13-compatible policy tree in a temp dir. Only references receivers
	// that were just provisioned (slack, discord) — no built-in system receivers.
	policyJSON := []byte(`{
	"group_by": ["grafana_folder", "alertname"],
	"receiver": "slack",
	"routes": [
		{
			"object_matchers": [["foo", "=", "22"]],
			"receiver": "discord"
		},
		{
			"continue": true,
			"object_matchers": [["moo", "=", "23"]],
			"receiver": "slack"
		}
	]
}`)
	policyTmpDir := t.TempDir()
	policyDir := filepath.Join(policyTmpDir, "org_main-org", "alerting")
	assert.NoError(t, os.MkdirAll(policyDir, 0o750))
	assert.NoError(t, os.WriteFile(filepath.Join(policyDir, "policies.json"), policyJSON, 0o644))

	// Copy the secure credentials dir into policyTmpDir so the upload client
	// can resolve auth. SecureLocation() resolves to {OutputPath}/secure, so
	// without this copy the client falls back to anonymous access and gets 403.
	secureDir := filepath.Join(policyTmpDir, "secure")
	assert.NoError(t, os.MkdirAll(secureDir, 0o750))
	assert.NoError(t, os.CopyFS(secureDir, os.DirFS("test/data/secure")))

	// Upload client reads the policy from policyTmpDir (auth + policy file present).
	uploadCfg := config.NewConfig(common.DefaultTestConfig)
	uploadCfg.GetDefaultGrafanaConfig().OutputPath = policyTmpDir
	uploadClient := test_tooling.CreateSimpleClientWithConfig(t, uploadCfg, r.Container)

	// Verify clean initial state.
	policies, err := uploadClient.ListAlertNotifications()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(policies.Routes), "initial policy tree should have no sub-routes")

	// Upload the v13-compatible policy tree.
	policiesListing, err := uploadClient.UploadAlertNotifications()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(policiesListing.Routes))

	// The first route uses discord as the receiver.
	discordRoute := lo.FindOrElse(policiesListing.Routes, nil, func(item *models.Route) bool {
		return item.Receiver == "discord"
	})
	assert.NotNil(t, discordRoute)
	assert.Equal(t, 3, len(discordRoute.ObjectMatchers[0]))
	assert.Equal(t, "22", discordRoute.ObjectMatchers[0][2])

	// The second route uses slack and has continue=true.
	slackRoute := lo.FindOrElse(policiesListing.Routes, nil, func(item *models.Route) bool {
		return item.Receiver == "slack"
	})
	assert.NotNil(t, slackRoute)
	assert.True(t, slackRoute.Continue)
	assert.Equal(t, "23", slackRoute.ObjectMatchers[0][2])

	// List confirms 2 sub-routes are active.
	policies, err = uploadClient.ListAlertNotifications()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(policies.Routes))

	// Download to a separate authenticated temp dir — never overwrites fixture files.
	dlClient := createDownloadClient(t, cfg, r.Container)
	data, err := dlClient.DownloadAlertNotifications()
	assert.NoError(t, err)
	rawData, err := os.ReadFile(data)
	assert.NoError(t, err)
	assert.True(t, bytes.Contains(rawData, []byte("grafana_folder")))
	assert.True(t, bytes.Contains(rawData, []byte("slack")))

	// Clear resets the policy tree back to the default (empty routes).
	err = uploadClient.ClearAlertNotifications()
	assert.NoError(t, err)
	policies, err = uploadClient.ListAlertNotifications()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(policies.Routes))
}
