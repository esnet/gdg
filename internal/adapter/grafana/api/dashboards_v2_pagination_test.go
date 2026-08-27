package api

// Regression test for the dashboard v2 pagination bug: the App Platform list
// endpoint paginates server-side and returns a metadata.continue token
// whenever more items remain, but a single unpaginated List() call silently
// truncates to just the first page -- on a real staging server this dropped
// 359 of 403 dashboards (89%) with no error and no log line, because the
// truncated items were never even seen by the filter step. listDashboardsV2
// was fixed to loop on GetContinue() until the server returns none; this
// test drives a fake App Platform server across three pages (via a real
// httptest.Server, not a client-go fake/reactor, so the exact HTTP-level
// continue-token contract is what's under test) and fails if that loop ever
// regresses back to a single call.

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// pagedDashboardServer serves `total` dashboards across pages of `pageSize`
// items via the App Platform list contract: each response includes
// metadata.continue whenever items remain, empty ("") on the final page.
// It also serves /api/org (for k8sNamespace, token auth) and /api/search
// (for the folder-UID map that listDashboardsV2 builds) so the whole
// DashboardServiceImpl.listDashboardsV2 call path resolves without a real
// Grafana instance. It records every continue token it received so the test
// can assert the client actually walked every page rather than looping
// forever or stopping early.
func pagedDashboardServer(t *testing.T, total, pageSize int) (*httptest.Server, *[]string) {
	t.Helper()
	var seenContinues []string

	mux := http.NewServeMux()
	mux.HandleFunc("/api/org", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"id": 1, "name": "Main Org."})
	})
	mux.HandleFunc("/api/search", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]any{}) // no folders needed for this test
	})
	mux.HandleFunc("/apis/dashboard.grafana.app/v2/namespaces/default/dashboards", func(w http.ResponseWriter, r *http.Request) {
		cont := r.URL.Query().Get("continue")
		seenContinues = append(seenContinues, cont)

		start := 0
		if cont != "" {
			_, _ = fmt.Sscanf(cont, "offset-%d", &start)
		}
		end := min(start+pageSize, total)

		items := make([]map[string]any, 0, end-start)
		for i := start; i < end; i++ {
			items = append(items, map[string]any{
				"apiVersion": "dashboard.grafana.app/v2",
				"kind":       "Dashboard",
				"metadata": map[string]any{
					"name": fmt.Sprintf("dash-%03d", i),
				},
				"spec": map[string]any{
					"title": fmt.Sprintf("Dashboard %03d", i),
				},
			})
		}

		nextContinue := ""
		if end < total {
			nextContinue = fmt.Sprintf("offset-%d", end)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"apiVersion": "dashboard.grafana.app/v2",
			"kind":       "DashboardList",
			"metadata":   map[string]any{"continue": nextContinue},
			"items":      items,
		})
	})

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return server, &seenContinues
}

// TestListDashboardsV2_WalksEveryPage is the direct regression test: with a
// server page size smaller than the total dashboard count (mirroring the
// real staging server's observed 44-item page against 403 real dashboards),
// listDashboardsV2 must return every item across every page, not just the
// first one.
func TestListDashboardsV2_WalksEveryPage(t *testing.T) {
	const total = 101
	const pageSize = 44 // matches the page size observed on the real staging server

	server, seenContinues := pagedDashboardServer(t, total, pageSize)
	svc := newTestDashboardService(t, server.URL, authToken)

	results := svc.listDashboardsV2(nil)

	require.Len(t, results, total, "must return every dashboard across every page, not just the first")

	seen := make(map[string]bool, total)
	for _, r := range results {
		seen[r.Resource.Name] = true
	}
	assert.Len(t, seen, total, "no dashboard should be lost or duplicated across the pagination loop")

	// 101 items at 44/page is 3 pages: continues of "", "offset-44", "offset-88".
	require.Len(t, *seenContinues, 3, "expected exactly 3 requests (one per page) for 101 items at 44/page")
	assert.Equal(t, []string{"", "offset-44", "offset-88"}, *seenContinues)
}

// TestListDashboardsV2_SinglePage_NoExtraRequest guards the other direction:
// when everything fits on one page, the loop must stop as soon as the server
// returns an empty continue token, not keep polling.
func TestListDashboardsV2_SinglePage_NoExtraRequest(t *testing.T) {
	const total = 5
	const pageSize = 44

	server, seenContinues := pagedDashboardServer(t, total, pageSize)
	svc := newTestDashboardService(t, server.URL, authToken)

	results := svc.listDashboardsV2(nil)

	require.Len(t, results, total)
	require.Len(t, *seenContinues, 1, "a single short page must not trigger a second request")
	assert.Equal(t, []string{""}, *seenContinues)
}
