package test

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"slices"
	"testing"

	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/config/config_domain"
	"github.com/esnet/gdg/pkg/test_tooling"
	"github.com/esnet/gdg/pkg/test_tooling/common"
	"github.com/esnet/gdg/pkg/test_tooling/path"
	"github.com/stretchr/testify/assert"
)

// makeAlertFilter builds an AlertSettings with the provided MatchingRules applied
// to contact points. Pass no rules to get a no-op filter (FiltersEnabled == false).
func makeAlertFilter(rules ...config_domain.MatchingRule) *config_domain.AlertSettings {
	return &config_domain.AlertSettings{
		ContactSettings: config_domain.ContactPointSettings{
			FilterRules: rules,
		},
	}
}

func TestContactsCrud(t *testing.T) {
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
	contactPoints, err := apiClient.ListContactPoints()
	assert.NoError(t, err)
	assert.Equal(t, len(contactPoints), 0, "Validate initial contact list is empty")
	contacts, err := apiClient.UploadContactPoints()
	assert.NoError(t, err)
	assert.Equal(t, len(contacts), 2)
	assert.True(t, slices.Contains(contacts, "discord"))
	contactPoints, err = apiClient.ListContactPoints()
	assert.NoError(t, err)
	assert.Equal(t, len(contactPoints), 2)
	data, err := apiClient.DownloadContactPoints()
	assert.NoError(t, err)
	assert.Equal(t, "test/data/org_main-org/alerting/contacts.json", data)
	rawData, err := os.ReadFile(data)
	assert.NoError(t, err)
	assert.True(t, bytes.Contains(rawData, []byte("discord")))
	assert.True(t, bytes.Contains(rawData, []byte("slack")))
	assert.False(t, bytes.Contains(rawData, []byte("email receiver")))
	contacts, err = apiClient.ClearContactPoints()
	assert.NoError(t, err)
	assert.Equal(t, len(contacts), 2)
}

// TestContactsFilterCrud exercises the full CRUD cycle (Upload → List → Download → Clear)
// for each contact-point filter scenario. All sub-tests share a single Grafana container;
// each sub-test clears Grafana at the end so the next starts from a clean slate.
//
// Test data (test/data/org_main-org/alerting/contacts.json) has two contact points:
//
//	"discord"  — receivers[0].type = "discord"  (no recipient field)
//	"slack"    — receivers[0].type = "slack",   settings.recipient = "testing"
//
// Grafana v13 also injects a built-in "email receiver" contact point with an empty UID.
// ListContactPoints (and therefore all CRUD ops) silently skips it, so it never
// appears in upload/list/download/clear counts.
func TestContactsFilterCrud(t *testing.T) {
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
		if cleanErr := r.CleanUp(); cleanErr != nil {
			slog.Warn("Unable to clean up after test", "test", t.Name())
		}
	}()

	type testCase struct {
		name              string
		filter            *config_domain.AlertSettings // nil = no filter
		wantUploadCount   int
		wantUploadNames   []string
		wantListCount     int
		wantInDownload    []string
		wantNotInDownload []string
		wantClearCount    int
	}

	tests := []testCase{
		{
			// Baseline: no filter — both contact points pass through every operation.
			name:            "NoFilter",
			filter:          nil,
			wantUploadCount: 2,
			wantUploadNames: []string{"discord", "slack"},
			wantListCount:   2,
			wantInDownload:  []string{"discord", "slack"},
			wantClearCount:  2,
		},
		{
			// Exclusive match on name: "discord" is excluded → only "slack" survives.
			name:              "ExcludeByName_discord",
			filter:            makeAlertFilter(config_domain.MatchingRule{Field: "name", Regex: "discord"}),
			wantUploadCount:   1,
			wantUploadNames:   []string{"slack"},
			wantListCount:     1,
			wantInDownload:    []string{"slack"},
			wantNotInDownload: []string{"discord"},
			wantClearCount:    1,
		},
		{
			// Inclusive match on name: keep only entries whose name matches "discord".
			// "slack" does not match → excluded.
			name:              "KeepOnlyByName_discord",
			filter:            makeAlertFilter(config_domain.MatchingRule{Field: "name", Regex: "discord", Inclusive: true}),
			wantUploadCount:   1,
			wantUploadNames:   []string{"discord"},
			wantListCount:     1,
			wantInDownload:    []string{"discord"},
			wantNotInDownload: []string{"slack"},
			wantClearCount:    1,
		},
		{
			// Exclude by nested receiver type: contacts whose receivers contain type "slack"
			// are excluded → only "discord" survives.
			name:              "ExcludeByReceiverType_slack",
			filter:            makeAlertFilter(config_domain.MatchingRule{Field: "receivers.#.type", Regex: "slack"}),
			wantUploadCount:   1,
			wantUploadNames:   []string{"discord"},
			wantListCount:     1,
			wantInDownload:    []string{"discord"},
			wantNotInDownload: []string{"slack"},
			wantClearCount:    1,
		},
		{
			// Exclude by deeply nested settings field: contacts where any receiver has
			// settings.recipient matching "testing" are excluded → only "discord" survives.
			name:              "ExcludeByNestedField_recipient",
			filter:            makeAlertFilter(config_domain.MatchingRule{Field: "receivers.#.settings.recipient", Regex: "testing"}),
			wantUploadCount:   1,
			wantUploadNames:   []string{"discord"},
			wantListCount:     1,
			wantInDownload:    []string{"discord"},
			wantNotInDownload: []string{"slack"},
			wantClearCount:    1,
		},
	}

	grafanaCfg := cfg.GetDefaultGrafanaConfig()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Inject (or clear) the filter for this sub-test.
			grafanaCfg.AlertSettings = tc.filter
			defer func() { grafanaCfg.AlertSettings = nil }()

			apiClient := r.ApiClient

			// ── Upload ────────────────────────────────────────────────────────
			// UploadContactPoints reads from the local JSON file and applies the
			// active filter — only non-excluded contact points are pushed to Grafana.
			uploaded, uploadErr := apiClient.UploadContactPoints()
			assert.NoError(t, uploadErr)
			assert.Equal(t, tc.wantUploadCount, len(uploaded), "upload count mismatch")
			for _, name := range tc.wantUploadNames {
				assert.True(t, slices.Contains(uploaded, name), "expected %q in uploaded list", name)
			}

			// ── List ──────────────────────────────────────────────────────────
			// ListContactPoints fetches from Grafana and applies the same filter.
			listed, listErr := apiClient.ListContactPoints()
			assert.NoError(t, listErr)
			assert.Equal(t, tc.wantListCount, len(listed), "list count mismatch")

			// ── Download ──────────────────────────────────────────────────────
			// DownloadContactPoints overwrites contacts.json with the filtered
			// Grafana state. Snapshot the original bytes first and restore
			// immediately after asserting — UploadContactPoints in the next
			// sub-test depends on the full two-entry fixture being present.
			const contactsPath = "test/data/org_main-org/alerting/contacts.json"
			originalContacts, snapErr := os.ReadFile(contactsPath)
			assert.NoError(t, snapErr, "failed to snapshot contacts.json before download")

			filePath, dlErr := apiClient.DownloadContactPoints()
			assert.NoError(t, dlErr)
			rawData, readErr := os.ReadFile(filePath)
			assert.NoError(t, readErr)
			for _, s := range tc.wantInDownload {
				assert.True(t, bytes.Contains(rawData, []byte(s)), "expected %q in downloaded file", s)
			}
			for _, s := range tc.wantNotInDownload {
				assert.False(t, bytes.Contains(rawData, []byte(s)), "did not expect %q in downloaded file", s)
			}

			// Restore immediately — don't wait for sub-test exit.
			assert.NoError(t, os.WriteFile(contactsPath, originalContacts, 0o644), "failed to restore contacts.json")

			// ── Clear ─────────────────────────────────────────────────────────
			// ClearContactPoints calls ListContactPoints internally, so the filter
			// applies: only the currently visible (non-excluded) contacts are removed.
			cleared, clearErr := apiClient.ClearContactPoints()
			assert.NoError(t, clearErr)
			assert.Equal(t, tc.wantClearCount, len(cleared), "clear count mismatch")

			// Confirm Grafana is empty for the next sub-test.
			remaining, err := apiClient.ListContactPoints()
			assert.NoError(t, err)
			assert.Empty(t, remaining, "Grafana should be empty after Clear")
		})
	}
}
