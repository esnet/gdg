package api

// Regression test for the folder pagination bug: baseService.listFolders()
// and DashNGoImpl.ListFolders() called the legacy /api/search endpoint with
// no Limit/Page set, so they got exactly one page at whatever Grafana's
// server-side default happens to be and never looped -- the same bug shape
// that silently dropped 89% of dashboards on the v2 App Platform path,
// applied to folders instead. searchAllPages() was added to loop Limit/Page
// (requesting the documented max of 5000 per page) until a page comes back
// short, mirroring the pattern listDashboardsV1 already used correctly for
// dashboards. This test drives a fake /api/search that honors the requested
// limit as its own page size -- the real contract per the client's doc
// comment ("Limit the number of returned results (max 5000)"), unlike the
// v2 App Platform's independent server-enforced page size -- across the
// 5000-item page boundary, and fails if the loop ever regresses back to a
// single unpaginated call.

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/go-openapi/strfmt"
	"github.com/grafana/grafana-openapi-client-go/client"
	"github.com/grafana/grafana-openapi-client-go/client/search"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// pagedSearchServer serves `total` dash-folder hits from /api/search, one
// page per request sized to whatever `limit` the client actually requests
// (the real /api/search contract -- unlike the v2 App Platform, which
// enforces its own page size regardless of the requested limit). It records
// every (limit, page) pair it received so the test can assert the client
// walked every page rather than stopping early or looping forever.
func pagedSearchServer(t *testing.T, total int) (*httptest.Server, *[][2]string) {
	t.Helper()
	var seenPages [][2]string

	mux := http.NewServeMux()
	mux.HandleFunc("/api/search", func(w http.ResponseWriter, r *http.Request) {
		limitStr := r.URL.Query().Get("limit")
		pageStr := r.URL.Query().Get("page")
		seenPages = append(seenPages, [2]string{limitStr, pageStr})

		limit, err := strconv.Atoi(limitStr)
		require.NoError(t, err, "listFolders/ListFolders must always send an explicit limit param")
		page, err := strconv.Atoi(pageStr)
		require.NoError(t, err, "listFolders/ListFolders must always send an explicit page param")
		require.GreaterOrEqual(t, page, 1)

		start := (page - 1) * limit
		end := min(start+limit, total)

		hits := make([]*models.Hit, 0, max(end-start, 0))
		for i := start; i < end; i++ {
			hits = append(hits, &models.Hit{
				UID:   "folder-" + strconv.Itoa(i),
				Title: "Folder " + strconv.Itoa(i),
				Type:  "dash-folder",
			})
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(hits)
	})

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return server, &seenPages
}

// newTestGrafanaClient builds a bare grafana-openapi-client-go client pointed
// at a fake server, bypassing GDG's config/auth plumbing entirely -- all
// searchAllPages needs is a *client.GrafanaHTTPAPI.
func newTestGrafanaClient(t *testing.T, serverURL string) *client.GrafanaHTTPAPI {
	t.Helper()
	u, err := url.Parse(serverURL)
	require.NoError(t, err)
	cfg := &client.TransportConfig{
		Host:     u.Host,
		BasePath: "/api",
		Schemes:  []string{u.Scheme},
	}
	return client.NewHTTPClientWithConfig(strfmt.Default, cfg)
}

var folderSearchType = "dash-folder"

// TestSearchAllPages_WalksEveryPage verifies searchAllPages loops past the
// documented 5000-item page boundary instead of stopping at the first page
// -- 5001 total items forces exactly two requests (a full 5000-item page,
// then a 1-item short page that ends the loop).
func TestSearchAllPages_WalksEveryPage(t *testing.T) {
	const total = 5001

	server, seenPages := pagedSearchServer(t, total)
	apiClient := newTestGrafanaClient(t, server.URL)

	hits := searchAllPages(apiClient, func(p *search.SearchParams) {
		p.Type = &folderSearchType
	})

	require.Len(t, hits, total, "must return every folder across every page, not just the first")

	seen := make(map[string]bool, total)
	for _, h := range hits {
		seen[h.UID] = true
	}
	assert.Len(t, seen, total, "no folder should be lost or duplicated across the pagination loop")

	require.Len(t, *seenPages, 2, "5001 items at 5000/page must take exactly 2 requests")
	assert.Equal(t, [][2]string{{"5000", "1"}, {"5000", "2"}}, *seenPages)
}

// TestSearchAllPages_ShortPage_StopsLooping guards the other direction: when
// everything fits on one (short) page, the loop must stop immediately, not
// keep requesting empty pages forever.
func TestSearchAllPages_ShortPage_StopsLooping(t *testing.T) {
	const total = 3

	server, seenPages := pagedSearchServer(t, total)
	apiClient := newTestGrafanaClient(t, server.URL)

	hits := searchAllPages(apiClient, func(p *search.SearchParams) {
		p.Type = &folderSearchType
	})

	require.Len(t, hits, total)
	require.Len(t, *seenPages, 1, "a single short page must not trigger a second request")
	assert.Equal(t, [][2]string{{"5000", "1"}}, *seenPages)
}
