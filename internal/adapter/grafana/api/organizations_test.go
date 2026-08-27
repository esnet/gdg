package api

// Unit tests for InitOrganizations' token-auth org-name resolution.
//
// Background: GrafanaConfig.GetOrganizationName() falls back to "unknown" for
// token auth when organization_name isn't set in the config -- this shows up
// throughout GDG's logs ("orgName=unknown") even though a service-account
// token is permanently scoped to exactly one org, so its identity is knowable
// via GET /api/org. InitOrganizations now resolves and caches that name once
// at startup instead of leaving every subsequent log line showing "unknown".

import (
	"net/http"
	"net/http/httptest"
	"testing"

	configDomain "github.com/esnet/gdg/internal/config/config_domain"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func newTestDashNGo(t *testing.T, serverURL string, token string) *DashNGoImpl {
	t.Helper()

	grafanaConf := &configDomain.GrafanaConfig{
		URL: serverURL,
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

	return &DashNGoImpl{baseService: baseService{
		gdgConfig:   gdgConfig,
		grafanaConf: grafanaConf,
	}}
}

// TestInitOrganizations_TokenAuth_NoConfiguredName_ResolvesFromToken verifies
// that when a token is configured with no organization_name set, GDG queries
// the token's own /api/org identity and caches the real org name, instead of
// leaving GetOrganizationName() at "unknown" for the rest of the run.
func TestInitOrganizations_TokenAuth_NoConfiguredName_ResolvesFromToken(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/org", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":1,"name":"Main Org."}`))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	svc := newTestDashNGo(t, server.URL, "glsa_faketoken")
	assert.Equal(t, "unknown", svc.grafanaConf.GetOrganizationName(), "sanity: unresolved token context starts at unknown")

	svc.InitOrganizations()

	assert.Equal(t, "Main Org.", svc.grafanaConf.GetOrganizationName())
}

// TestInitOrganizations_TokenAuth_ExplicitName_KeepsWarningPath verifies that
// when organization_name IS explicitly configured for a token, the existing
// "tokens don't operate across multiple orgs" warning path is preserved and
// no live /api/org lookup happens (the configured name is trusted as-is).
func TestInitOrganizations_TokenAuth_ExplicitName_KeepsWarningPath(t *testing.T) {
	var orgHit bool
	mux := http.NewServeMux()
	mux.HandleFunc("/api/org", func(w http.ResponseWriter, r *http.Request) {
		orgHit = true
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":2,"name":"Second Org"}`))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	svc := newTestDashNGo(t, server.URL, "glsa_faketoken")
	svc.grafanaConf.OrganizationName = "Configured Org"

	svc.InitOrganizations()

	assert.Equal(t, "Configured Org", svc.grafanaConf.GetOrganizationName(), "explicitly configured org name must not be overwritten")
	assert.False(t, orgHit, "no live lookup should happen when organization_name is already configured")
}
