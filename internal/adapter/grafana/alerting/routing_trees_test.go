package alerting

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/esnet/gdg/internal/adapter/grafana/resources"
	"github.com/esnet/gdg/internal/adapter/storage"
	"github.com/esnet/gdg/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// routingTreeServer fakes the App Platform list contract for routing trees,
// serving the given items. It also stubs /api/org so the full
// ListRoutingTrees path runs without a real Grafana.
func routingTreeServer(t *testing.T, items []map[string]any) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/org", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"id": 1, "name": "Main Org."})
	})
	mux.HandleFunc("/apis/notifications.alerting.grafana.app/v1beta1/namespaces/default/routingtrees", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"apiVersion": "notifications.alerting.grafana.app/v1beta1",
			"kind":       "RoutingTreeList",
			"metadata":   map[string]any{},
			"items":      items,
		})
	})

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return server
}

func routingTreeItem(name, receiver string, routeCount int) map[string]any {
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

func TestListRoutingTrees_ReturnsEveryTree(t *testing.T) {
	items := []map[string]any{
		routingTreeItem("moo", "devnull", 1),
		routingTreeItem("user-defined", "grafana-default-email", 3),
	}
	server := routingTreeServer(t, items)
	svc := newTestAlertingService(t, server.URL)

	result, err := svc.ListRoutingTrees("")

	require.NoError(t, err)
	require.Len(t, result, 2)
	assert.Equal(t, "moo", result[0].Name())
	assert.Equal(t, "devnull", result[0].DefaultReceiver())
	assert.Equal(t, 1, result[0].RouteCount())
	assert.Equal(t, domain.GdgRoutingTreeApiVersion, result[0].GdgApiVersion)

	assert.Equal(t, "user-defined", result[1].Name())
	assert.Equal(t, "grafana-default-email", result[1].DefaultReceiver())
	assert.Equal(t, 3, result[1].RouteCount())
}

func TestListRoutingTrees_Empty(t *testing.T) {
	server := routingTreeServer(t, []map[string]any{})
	svc := newTestAlertingService(t, server.URL)

	result, err := svc.ListRoutingTrees("")

	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestListRoutingTrees_FiltersByName(t *testing.T) {
	items := []map[string]any{
		routingTreeItem("moo", "devnull", 1),
		routingTreeItem("user-defined", "grafana-default-email", 3),
	}
	server := routingTreeServer(t, items)
	svc := newTestAlertingService(t, server.URL)

	result, err := svc.ListRoutingTrees("user-defined")

	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, "user-defined", result[0].Name())
}

func TestListRoutingTrees_NameFilterNoMatch(t *testing.T) {
	items := []map[string]any{
		routingTreeItem("moo", "devnull", 1),
	}
	server := routingTreeServer(t, items)
	svc := newTestAlertingService(t, server.URL)

	result, err := svc.ListRoutingTrees("does-not-exist")

	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestDownloadRoutingTrees_WritesOneFilePerTree(t *testing.T) {
	items := []map[string]any{
		routingTreeItem("moo", "devnull", 1),
		routingTreeItem("user-defined", "grafana-default-email", 3),
	}
	server := routingTreeServer(t, items)
	svc := newTestAlertingService(t, server.URL)

	tmpDir := t.TempDir()
	svc.GrafanaConf.OutputPath = tmpDir
	svc.Storage = storage.NewLocalStorage(t.Context())

	files, err := svc.DownloadRoutingTrees("")

	require.NoError(t, err)
	require.Len(t, files, 2)

	expected := []string{
		filepath.Join(tmpDir, "org_main-org", "alerting-routing-trees", "moo.json"),
		filepath.Join(tmpDir, "org_main-org", "alerting-routing-trees", "user-defined.json"),
	}
	assert.ElementsMatch(t, expected, files)

	for _, f := range files {
		raw, readErr := os.ReadFile(f)
		require.NoError(t, readErr)
		var gdg domain.RoutingTreeGdg
		require.NoError(t, json.Unmarshal(raw, &gdg))
		assert.Equal(t, domain.GdgRoutingTreeApiVersion, gdg.GdgApiVersion)
		assert.NotEmpty(t, gdg.Name())
	}
}

func TestDownloadRoutingTrees_NoTrees(t *testing.T) {
	server := routingTreeServer(t, []map[string]any{})
	svc := newTestAlertingService(t, server.URL)

	tmpDir := t.TempDir()
	svc.GrafanaConf.OutputPath = tmpDir
	svc.Storage = storage.NewLocalStorage(t.Context())

	files, err := svc.DownloadRoutingTrees("")

	require.NoError(t, err)
	assert.Empty(t, files)
}

func TestDownloadRoutingTrees_FiltersByName(t *testing.T) {
	items := []map[string]any{
		routingTreeItem("moo", "devnull", 1),
		routingTreeItem("user-defined", "grafana-default-email", 3),
	}
	server := routingTreeServer(t, items)
	svc := newTestAlertingService(t, server.URL)

	tmpDir := t.TempDir()
	svc.GrafanaConf.OutputPath = tmpDir
	svc.Storage = storage.NewLocalStorage(t.Context())

	files, err := svc.DownloadRoutingTrees("moo")

	require.NoError(t, err)
	require.Len(t, files, 1)
	assert.Contains(t, files[0], "moo.json")
}

// ---------------------------------------------------------------------------
// ClearRoutingTrees
// ---------------------------------------------------------------------------

// routingTreeClearServer fakes the App Platform list/delete/update contract.
// deletable controls which names a DELETE call succeeds for; any other name
// returns 403 Forbidden, simulating a protected/default tree that cannot be
// removed outright. Every successful PUT (reset) is recorded in resets.
func routingTreeClearServer(t *testing.T, items []map[string]any, deletable map[string]bool) (*httptest.Server, *[]map[string]any) {
	t.Helper()
	var resets []map[string]any

	mux := http.NewServeMux()
	mux.HandleFunc("/api/org", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"id": 1, "name": "Main Org."})
	})
	basePath := "/apis/notifications.alerting.grafana.app/v1beta1/namespaces/default/routingtrees"
	mux.HandleFunc(basePath, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"apiVersion": "notifications.alerting.grafana.app/v1beta1",
			"kind":       "RoutingTreeList",
			"metadata":   map[string]any{},
			"items":      items,
		})
	})
	mux.HandleFunc(basePath+"/", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Path[len(basePath)+1:]
		switch r.Method {
		case http.MethodDelete:
			if deletable[name] {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{}`))
				return
			}
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"kind":    "Status",
				"status":  "Failure",
				"message": "the default routing tree cannot be deleted",
				"reason":  "Forbidden",
				"code":    403,
			})
		case http.MethodPut:
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			resets = append(resets, body)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(body)
		case http.MethodGet:
			for _, item := range items {
				md, _ := item["metadata"].(map[string]any)
				if md["name"] == name {
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(item)
					return
				}
			}
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return server, &resets
}

func TestClearRoutingTrees_DeletesEverythingDeletable(t *testing.T) {
	items := []map[string]any{
		routingTreeItem("moo", "devnull", 1),
		routingTreeItem("other", "devnull", 2),
	}
	server, resets := routingTreeClearServer(t, items, map[string]bool{"moo": true, "other": true})
	svc := newTestAlertingService(t, server.URL)

	cleared, err := svc.ClearRoutingTrees("")

	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"moo", "other"}, cleared)
	assert.Empty(t, *resets, "deletable trees must not be reset")
}

func TestClearRoutingTrees_ResetsProtectedTreeOnDeleteFailure(t *testing.T) {
	items := []map[string]any{
		routingTreeItem("moo", "devnull", 1),
		routingTreeItem("user-defined", "grafana-default-email", 3),
	}
	server, resets := routingTreeClearServer(t, items, map[string]bool{"moo": true})
	svc := newTestAlertingService(t, server.URL)

	cleared, err := svc.ClearRoutingTrees("")

	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"moo", "user-defined"}, cleared)
	require.Len(t, *resets, 1, "the protected tree must be reset via PUT instead of deleted")
	spec, _ := (*resets)[0]["spec"].(map[string]any)
	assert.Equal(t, domain.DefaultRoutingTreeSpec(), spec)
	md, _ := (*resets)[0]["metadata"].(map[string]any)
	assert.Equal(t, "user-defined", md["name"])
}

func TestClearRoutingTrees_FiltersByName(t *testing.T) {
	items := []map[string]any{
		routingTreeItem("moo", "devnull", 1),
		routingTreeItem("other", "devnull", 2),
	}
	server, _ := routingTreeClearServer(t, items, map[string]bool{"moo": true, "other": true})
	svc := newTestAlertingService(t, server.URL)

	cleared, err := svc.ClearRoutingTrees("moo")

	require.NoError(t, err)
	assert.Equal(t, []string{"moo"}, cleared)
}

func TestClearRoutingTrees_NoTrees(t *testing.T) {
	server, _ := routingTreeClearServer(t, []map[string]any{}, map[string]bool{})
	svc := newTestAlertingService(t, server.URL)

	cleared, err := svc.ClearRoutingTrees("")

	require.NoError(t, err)
	assert.Empty(t, cleared)
}

// ---------------------------------------------------------------------------
// UploadRoutingTrees
// ---------------------------------------------------------------------------

// routingTreeState is a stateful fake of the App Platform routing tree
// resource store, letting upload tests exercise the full Get-then-Create/Update
// reconcile flow and the delete/reset pruning step against realistic sequential
// behavior instead of canned per-name responses.
type routingTreeState struct {
	mu        sync.Mutex
	items     map[string]map[string]any
	protected map[string]bool
	created   []string
	updated   []string
	deleted   []string
	reset     []string
}

func newRoutingTreeState(initial []map[string]any, protected map[string]bool) *routingTreeState {
	items := make(map[string]map[string]any, len(initial))
	for _, item := range initial {
		md, _ := item["metadata"].(map[string]any)
		name, _ := md["name"].(string)
		items[name] = item
	}
	return &routingTreeState{items: items, protected: protected}
}

func (s *routingTreeState) server(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/org", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"id": 1, "name": "Main Org."})
	})
	basePath := "/apis/notifications.alerting.grafana.app/v1beta1/namespaces/default/routingtrees"
	mux.HandleFunc(basePath, func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		defer s.mu.Unlock()
		switch r.Method {
		case http.MethodGet:
			items := make([]map[string]any, 0, len(s.items))
			for _, item := range s.items {
				items = append(items, item)
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"apiVersion": "notifications.alerting.grafana.app/v1beta1",
				"kind":       "RoutingTreeList",
				"metadata":   map[string]any{},
				"items":      items,
			})
		case http.MethodPost:
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			md, _ := body["metadata"].(map[string]any)
			name, _ := md["name"].(string)
			s.items[name] = body
			s.created = append(s.created, name)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(body)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc(basePath+"/", func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		defer s.mu.Unlock()
		name := r.URL.Path[len(basePath)+1:]
		switch r.Method {
		case http.MethodGet:
			item, ok := s.items[name]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(map[string]any{
					"kind": "Status", "status": "Failure", "reason": "NotFound", "code": 404,
				})
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(item)
		case http.MethodPut:
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			s.items[name] = body
			s.updated = append(s.updated, name)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(body)
		case http.MethodDelete:
			if s.protected[name] {
				w.WriteHeader(http.StatusForbidden)
				_ = json.NewEncoder(w).Encode(map[string]any{
					"kind": "Status", "status": "Failure", "reason": "Forbidden", "code": 403,
				})
				return
			}
			delete(s.items, name)
			s.deleted = append(s.deleted, name)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return server
}

func writeRoutingTreeFile(t *testing.T, svc *serviceImpl, name, receiver string) {
	t.Helper()
	gdg := &domain.RoutingTreeGdg{
		Resource:      routingTreeItem(name, receiver, 1),
		GdgApiVersion: domain.GdgRoutingTreeApiVersion,
	}
	raw, err := json.Marshal(gdg)
	require.NoError(t, err)
	path := svc.Resources.BuildResourcePath(svc.GrafanaConf, name, domain.RoutingTreeResource, true, false)
	require.NoError(t, svc.Storage.WriteFile(path, raw))
}

func newTestUploadService(t *testing.T, serverURL string) *serviceImpl {
	t.Helper()
	svc := newTestAlertingService(t, serverURL)
	svc.GrafanaConf.OutputPath = t.TempDir()
	svc.Storage = storage.NewLocalStorage(t.Context())
	svc.Resources = resources.NewHelpers()
	return svc
}

func TestUploadRoutingTrees_CreatesNewTree(t *testing.T) {
	state := newRoutingTreeState(nil, nil)
	server := state.server(t)
	svc := newTestUploadService(t, server.URL)

	writeRoutingTreeFile(t, svc, "new-tree", "devnull")

	uploaded, err := svc.UploadRoutingTrees("")

	require.NoError(t, err)
	assert.Equal(t, []string{"new-tree"}, uploaded)
	assert.Equal(t, []string{"new-tree"}, state.created)
	assert.Empty(t, state.updated)
}

func TestUploadRoutingTrees_UpdatesExistingTree(t *testing.T) {
	state := newRoutingTreeState([]map[string]any{
		routingTreeItem("already-exists", "devnull", 0),
	}, nil)
	server := state.server(t)
	svc := newTestUploadService(t, server.URL)

	writeRoutingTreeFile(t, svc, "already-exists", "devnull")

	uploaded, err := svc.UploadRoutingTrees("")

	require.NoError(t, err)
	assert.Equal(t, []string{"already-exists"}, uploaded)
	assert.Empty(t, state.created)
	assert.Equal(t, []string{"already-exists"}, state.updated)
}

func TestUploadRoutingTrees_FiltersByName(t *testing.T) {
	state := newRoutingTreeState(nil, nil)
	server := state.server(t)
	svc := newTestUploadService(t, server.URL)

	writeRoutingTreeFile(t, svc, "tree-a", "devnull")
	writeRoutingTreeFile(t, svc, "tree-b", "devnull")

	uploaded, err := svc.UploadRoutingTrees("tree-a")

	require.NoError(t, err)
	assert.Equal(t, []string{"tree-a"}, uploaded)
	assert.Equal(t, []string{"tree-a"}, state.created)
}

func TestUploadRoutingTrees_NoLocalFiles(t *testing.T) {
	state := newRoutingTreeState(nil, nil)
	server := state.server(t)
	svc := newTestUploadService(t, server.URL)

	// Pre-create the empty routing-trees directory, mirroring the real-world
	// case where a prior "download" created it but it has since been emptied,
	// rather than the directory never having existed at all.
	svc.Resources.BuildResourceFolder(svc.GrafanaConf, "", domain.RoutingTreeResource, true, false)

	uploaded, err := svc.UploadRoutingTrees("")

	require.NoError(t, err)
	assert.Empty(t, uploaded)
}

// TestUploadRoutingTrees_PrunesServerTreesNotInLocalStorage verifies the core
// "server reflects local" guarantee: a tree that exists on the server but has
// no corresponding local file must be removed, not left behind.
func TestUploadRoutingTrees_PrunesServerTreesNotInLocalStorage(t *testing.T) {
	state := newRoutingTreeState([]map[string]any{
		routingTreeItem("keep-me", "devnull", 0),
		routingTreeItem("remove-me", "devnull", 0),
	}, nil)
	server := state.server(t)
	svc := newTestUploadService(t, server.URL)

	writeRoutingTreeFile(t, svc, "keep-me", "devnull")

	uploaded, err := svc.UploadRoutingTrees("")

	require.NoError(t, err)
	assert.Equal(t, []string{"keep-me"}, uploaded)
	assert.Equal(t, []string{"remove-me"}, state.deleted)
	_, stillExists := state.items["remove-me"]
	assert.False(t, stillExists, "server-side tree with no local file must be removed")
	_, keptExists := state.items["keep-me"]
	assert.True(t, keptExists)
}

// TestUploadRoutingTrees_ResetsProtectedTreeOnPrune verifies that when the
// server refuses to delete an orphaned tree (the protected/default tree),
// upload resets it to the default spec instead of leaving it untouched.
func TestUploadRoutingTrees_ResetsProtectedTreeOnPrune(t *testing.T) {
	state := newRoutingTreeState([]map[string]any{
		routingTreeItem("user-defined", "grafana-default-email", 3),
	}, map[string]bool{"user-defined": true})
	server := state.server(t)
	svc := newTestUploadService(t, server.URL)
	svc.Resources.BuildResourceFolder(svc.GrafanaConf, "", domain.RoutingTreeResource, true, false)

	// No local file for "user-defined" at all.
	uploaded, err := svc.UploadRoutingTrees("")

	require.NoError(t, err)
	assert.Empty(t, uploaded)
	assert.Equal(t, []string{"user-defined"}, state.updated, "protected tree must be reset via PUT")
	spec, _ := state.items["user-defined"]["spec"].(map[string]any)
	assert.Equal(t, domain.DefaultRoutingTreeSpec(), spec)
}

// TestUploadRoutingTrees_PruneRespectsNameFilter verifies that when --name
// scopes the upload to a single tree, pruning only considers that tree too —
// other orphaned server trees outside the filter must be left alone.
func TestUploadRoutingTrees_PruneRespectsNameFilter(t *testing.T) {
	state := newRoutingTreeState([]map[string]any{
		routingTreeItem("orphan-in-scope", "devnull", 0),
		routingTreeItem("orphan-out-of-scope", "devnull", 0),
	}, nil)
	server := state.server(t)
	svc := newTestUploadService(t, server.URL)
	svc.Resources.BuildResourceFolder(svc.GrafanaConf, "", domain.RoutingTreeResource, true, false)

	uploaded, err := svc.UploadRoutingTrees("orphan-in-scope")

	require.NoError(t, err)
	assert.Empty(t, uploaded)
	assert.Equal(t, []string{"orphan-in-scope"}, state.deleted)
	_, stillExists := state.items["orphan-out-of-scope"]
	assert.True(t, stillExists, "trees outside the --name filter must not be pruned")
}
