package api

// Unit tests for k8sNamespace — the App Platform namespace resolution used by
// the v2 dashboard listing/upload path.
//
// Background: a Grafana service-account token is permanently scoped to a
// single org and cannot switch orgs, and an anonymous request is likewise
// pinned to whatever org auth.anonymous configures server-side. The legacy
// /api/user/orgs endpoint that GetConfiguredOrgId uses to resolve
// organization_name -> org ID is a signed-in-user-session endpoint; neither
// tokens nor anonymous requests can use it reliably (401/403). Before the
// first fix, k8sNamespace() always went through that lookup regardless of
// auth mode, so for a token whose real org wasn't org 1, the lookup would
// fail and silently fall back to "default" (org 1) -- querying the wrong
// namespace and returning zero dashboards even though dashboards exist in
// the token's real org. Because GDG's own test Grafana instance only ever
// has a single org, the previous integration tests could never catch this:
// the broken fallback and the correct answer were both "default".
//
// A second gap surfaced against a real multi-org server: anonymous access
// (no token, no basic auth) was still routed through the by-name
// /api/user/orgs lookup, which anonymous requests also can't use (401) --
// and unlike the token case, requesting the wrong namespace isn't silently
// wrong, it's rejected outright by the App Platform ("invalid namespace" /
// "namespace mismatch"), breaking the dashboard list entirely. Anonymous
// access needed the same /api/org-based resolution as tokens.
//
// These tests exercise k8sNamespace() directly against a fake Grafana server
// so both the org-2 case and the anonymous case are reachable without a real
// multi-org server.

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

// authMode selects which authentication the test's GrafanaConfig simulates.
type authMode int

const (
	authToken authMode = iota
	authAnonymous
	authBasic
)

// newTestDashboardService builds a minimal DashboardServiceImpl pointed at a
// fake Grafana server, with no config file / viper involvement.
func newTestDashboardService(t *testing.T, serverURL string, mode authMode) *DashboardServiceImpl {
	t.Helper()

	grafanaConf := &configDomain.GrafanaConfig{
		URL:              serverURL,
		OrganizationName: "Main Org.",
	}
	switch mode {
	case authToken:
		grafanaConf.Apply(configDomain.WithSecureAuth(configDomain.SecureModel{Token: "glsa_faketoken"}))
	case authBasic:
		grafanaConf.UserName = "admin"
		grafanaConf.Apply(configDomain.WithSecureAuth(configDomain.SecureModel{Password: "admin"}))
	case authAnonymous:
		// no token, no username/password -- IsBasicAuth() and GetAPIToken()
		// are both false, matching a real anonymous-access config.
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

	svc := newTestDashboardService(t, server.URL, authToken)

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

	svc := newTestDashboardService(t, server.URL, authToken)

	require.Equal(t, "default", svc.k8sNamespace())
}

// TestK8sNamespace_Anonymous_ResolvesViaCurrentOrg verifies that anonymous
// access (no token, no basic auth) resolves the namespace via /api/org, the
// same way token auth does, instead of the by-name /api/user/orgs lookup
// that anonymous requests can't use either. This is the exact failure mode
// seen against a real server: an anonymous context configured for a
// non-default org ("Public", org 2) was requesting the "default" namespace
// (org 1) instead, and the App Platform rejected it outright rather than
// silently returning the wrong org's dashboards.
func TestK8sNamespace_Anonymous_ResolvesViaCurrentOrg(t *testing.T) {
	var userOrgsHit bool

	mux := http.NewServeMux()
	mux.HandleFunc("/api/org", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"id": 2, "name": "Public"})
	})
	mux.HandleFunc("/api/user/orgs", func(w http.ResponseWriter, r *http.Request) {
		userOrgsHit = true
		w.WriteHeader(http.StatusUnauthorized)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	svc := newTestDashboardService(t, server.URL, authAnonymous)

	ns := svc.k8sNamespace()

	assert.Equal(t, "org-2", ns, "anonymous access scoped to org 2 must resolve to the org-2 namespace")
	assert.False(t, userOrgsHit, "anonymous access must not depend on the user-session /api/user/orgs endpoint")
}

// TestK8sNamespace_BasicAuth_UsesConfiguredOrgName verifies that basic/admin
// auth still resolves the namespace by looking up the configured
// organization_name via /api/user/orgs, unchanged from before -- basic auth
// is the one mode that can actually switch orgs, so the by-name lookup is
// still correct and preferred there.
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

	svc := newTestDashboardService(t, server.URL, authBasic)

	assert.Equal(t, "org-3", svc.k8sNamespace())
}
