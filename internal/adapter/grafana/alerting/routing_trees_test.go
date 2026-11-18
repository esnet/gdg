package alerting

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

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

	result, err := svc.ListRoutingTrees()

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

	result, err := svc.ListRoutingTrees()

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

	files, err := svc.DownloadRoutingTrees()

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

	files, err := svc.DownloadRoutingTrees()

	require.NoError(t, err)
	assert.Empty(t, files)
}
