package alerting

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/esnet/gdg/internal/adapter/grafana/core"
	"github.com/esnet/gdg/internal/adapter/grafana/resources"
	configDomain "github.com/esnet/gdg/internal/config/config_domain"
	"github.com/esnet/gdg/internal/domain"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestAlertingService builds a minimal serviceImpl pointed at a fake
// Grafana server, with no config file / viper involvement.
func newTestAlertingService(t *testing.T, serverURL string) *serviceImpl {
	t.Helper()

	grafanaConf := &configDomain.GrafanaConfig{
		URL:              serverURL,
		OrganizationName: "Main Org.",
	}
	grafanaConf.Apply(configDomain.WithSecureAuth(configDomain.SecureModel{Token: "glsa_faketoken"}))

	gdgConfig := &configDomain.GDGAppConfiguration{
		ViperConfig: viper.New(),
		ContextName: "test",
		Contexts: map[string]*configDomain.GrafanaConfig{
			"test": grafanaConf,
		},
		Global: &configDomain.AppGlobals{RetryCount: 1, RetryDelay: "5ms"},
	}

	return &serviceImpl{
		Service: core.Service{
			GdgConfig:   gdgConfig,
			GrafanaConf: grafanaConf,
			Resources:   resources.NewHelpers(),
		},
	}
}

// TestRoutingTreeClient_DiscoversPreferredVersion verifies that the dynamic
// client is built against the version returned by App Platform discovery
// (GET /apis), not hardcoded to the fallback constant.
func TestRoutingTreeClient_DiscoversPreferredVersion(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/org", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"id": 1, "name": "Main Org."})
	})
	mux.HandleFunc("/apis", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"groups": []map[string]any{
				{
					"name":             domain.RoutingTreeGroup,
					"preferredVersion": map[string]any{"version": "v2future"},
				},
			},
		})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	svc := newTestAlertingService(t, server.URL)

	client, ns, apiVersion, err := svc.routingTreeClient()

	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "default", ns)
	assert.Equal(t, domain.RoutingTreeGroup+"/v2future", apiVersion)
}

// TestRoutingTreeClient_FallsBackOnDiscoveryFailure verifies that when App
// Platform discovery fails, the client falls back to the hardcoded fallback version.
func TestRoutingTreeClient_FallsBackOnDiscoveryFailure(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/org", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"id": 1, "name": "Main Org."})
	})
	mux.HandleFunc("/apis", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	svc := newTestAlertingService(t, server.URL)

	client, ns, apiVersion, err := svc.routingTreeClient()

	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "default", ns)
	assert.Equal(t, domain.GdgRoutingTreeApiVersion, apiVersion)
}
