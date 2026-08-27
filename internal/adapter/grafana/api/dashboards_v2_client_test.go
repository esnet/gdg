package api

// Unit tests for k8sNamespace — the App Platform namespace resolution used by
// the v2 dashboard listing/upload path.
//
// Background: a Grafana service-account token is permanently scoped to a
// single org and cannot switch orgs. The legacy /api/user/orgs endpoint that
// GetConfiguredOrgId uses to resolve organization_name -> org ID is a
// signed-in-user-session endpoint; tokens can't use it reliably. Before this
// fix, k8sNamespace() always went through that lookup regardless of auth
// mode, so for a token whose real org wasn't org 1, the lookup would fail and
// silently fall back to "default" (org 1) -- querying the wrong namespace and
// returning zero dashboards even though dashboards exist in the token's real
// org. Because GDG's own test Grafana instance only ever has a single org,
// the previous integration tests could never catch this: the broken fallback
// and the correct answer were both "default".
//
// These tests exercise k8sNamespace() directly against a fake Grafana server
// so the org-2 case is actually reachable without a real multi-org server.

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	configDomain "github.com/esnet/gdg/internal/config/config_domain"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestDashboardService builds a minimal DashboardServiceImpl pointed at a
// fake Grafana server, with no config file / viper involvement.
func newTestDashboardService(t *testing.T, serverURL string, token string) *DashboardServiceImpl {
	t.Helper()

	grafanaConf := &configDomain.GrafanaConfig{
		URL:              serverURL,
		OrganizationName: "Main Org.",
	}
	if token != "" {
		grafanaConf.Apply(configDomain.WithSecureAuth(configDomain.SecureModel{Token: token}))
	}

	gdgConfig := &configDomain.GDGAppConfiguration{
		ViperConfig: viper.New(),
		ContextName: "test",
		Contexts: map[string]*configDomain.GrafanaConfig{
			"test": grafanaConf,
		},
		Global: &configDomain.AppGlobals{RetryCount: 1, RetryDelay: "5ms"},
	}

	return &DashboardServiceImpl{baseService: baseService{
		gdgConfig:   gdgConfig,
		grafanaConf: grafanaConf,
	}}
}

// TestK8sNamespace_TokenAuth_ResolvesViaCurrentOrg verifies that, for token
// auth, k8sNamespace resolves the namespace from the token's own /api/org
// identity (which works for service account tokens) instead of the legacy
// /api/user/orgs by-name lookup (which does not). A token belonging to org 2
// must resolve to "org-2", not silently fall back to "default".
func TestK8sNamespace_TokenAuth_ResolvesViaCurrentOrg(t *testing.T) {
	var userOrgsHit bool

	mux := http.NewServeMux()
	mux.HandleFunc("/api/org", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"id": 2, "name": "Second Org"})
	})
	// A token can't use /api/user/orgs reliably; simulate that by failing it.
	// If k8sNamespace still depends on this for token auth, the test below
	// will observe the wrong ("default") namespace.
	mux.HandleFunc("/api/user/orgs", func(w http.ResponseWriter, r *http.Request) {
		userOrgsHit = true
		w.WriteHeader(http.StatusForbidden)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	svc := newTestDashboardService(t, server.URL, "glsa_faketoken")

	ns := svc.k8sNamespace()

	assert.Equal(t, "org-2", ns, "token belonging to org 2 must resolve to the org-2 namespace")
	assert.False(t, userOrgsHit, "token auth must not depend on the user-session /api/user/orgs endpoint")
}

// TestK8sNamespace_TokenAuth_DefaultOrg verifies the common case: a token
// belonging to org 1 resolves to "default".
func TestK8sNamespace_TokenAuth_DefaultOrg(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/org", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"id": 1, "name": "Main Org."})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	svc := newTestDashboardService(t, server.URL, "glsa_faketoken")

	require.Equal(t, "default", svc.k8sNamespace())
}

// TestK8sNamespace_BasicAuth_UsesConfiguredOrgName verifies that basic/admin
// auth still resolves the namespace by looking up the configured
// organization_name via /api/user/orgs, unchanged from before.
func TestK8sNamespace_BasicAuth_UsesConfiguredOrgName(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/user/orgs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{"orgId": 3, "name": "Main Org."},
		})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	svc := newTestDashboardService(t, server.URL, "")

	assert.Equal(t, "org-3", svc.k8sNamespace())
}
