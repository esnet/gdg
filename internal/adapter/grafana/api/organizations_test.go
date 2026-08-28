package api

// Unit tests for InitOrganizations' org-name resolution for auth modes that
// cannot switch orgs (token and anonymous access).
//
// Background: GrafanaConfig.GetOrganizationName() falls back to "unknown" for
// these modes when organization_name isn't set in the config -- this showed
// up throughout GDG's logs ("orgName=unknown") even though a service-account
// token or an anonymous request is permanently scoped to exactly one org, so
// its identity is knowable via GET /api/org. The original fix only handled
// this for tokens with no configured organization_name; it missed anonymous
// access entirely (silently stuck at "unknown" forever, since it has no
// token to trigger the resolution) and never checked a configured name
// against reality even for tokens (a misconfigured organization_name for a
// token just got a generic "can't verify" warning, no actual comparison).
//
// resolvePinnedOrgIdentity replaces all of that with a single path for any
// non-basic-auth identity: always ask GET /api/org, and treat that answer as
// authoritative -- caching it when nothing was configured, and overriding
// (with a loud warning) when the configured name disagrees with reality.
// That org name isn't just cosmetic: it feeds on-disk backup/restore paths
// via GetPath(..., GetOrganizationName()).

import (
	"encoding/json"
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

// orgHandler stands up a fake /api/org endpoint reporting the given org.
func orgHandler(t *testing.T, id int64, name string) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/org", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"id": id, "name": name})
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return server
}

// TestInitOrganizations_TokenAuth_NoConfiguredName_ResolvesFromToken verifies
// that when a token is configured with no organization_name set, GDG queries
// the token's own /api/org identity and caches the real org name, instead of
// leaving GetOrganizationName() at "unknown" for the rest of the run.
func TestInitOrganizations_TokenAuth_NoConfiguredName_ResolvesFromToken(t *testing.T) {
	server := orgHandler(t, 1, "Main Org.")

	svc := newTestDashNGo(t, server.URL, "glsa_faketoken")
	assert.Equal(t, "unknown", svc.grafanaConf.GetOrganizationName(), "sanity: unresolved token context starts at unknown")

	svc.InitOrganizations()

	assert.Equal(t, "Main Org.", svc.grafanaConf.GetOrganizationName())
}

// TestInitOrganizations_TokenAuth_ConfiguredNameMatches_NoOverride verifies
// that when organization_name already matches the token's real org, it's
// left exactly as configured (and, incidentally, actually gets checked now).
func TestInitOrganizations_TokenAuth_ConfiguredNameMatches_NoOverride(t *testing.T) {
	server := orgHandler(t, 1, "Main Org.")

	svc := newTestDashNGo(t, server.URL, "glsa_faketoken")
	svc.grafanaConf.OrganizationName = "Main Org."

	svc.InitOrganizations()

	assert.Equal(t, "Main Org.", svc.grafanaConf.GetOrganizationName())
}

// TestInitOrganizations_TokenAuth_ConfiguredNameMismatch_OverridesToActual
// verifies the reliability fix: previously a token with an explicitly
// configured organization_name that didn't match its real org was never
// checked against reality -- it just got a generic warning and kept running
// under the wrong name for the rest of the session (which feeds real
// on-disk backup/restore paths). Now the live identity wins.
func TestInitOrganizations_TokenAuth_ConfiguredNameMismatch_OverridesToActual(t *testing.T) {
	server := orgHandler(t, 2, "Second Org")

	svc := newTestDashNGo(t, server.URL, "glsa_faketoken")
	svc.grafanaConf.OrganizationName = "Configured Org"

	svc.InitOrganizations()

	assert.Equal(t, "Second Org", svc.grafanaConf.GetOrganizationName(), "the token's actual org must win over a stale/incorrect configured name")
}

// TestInitOrganizations_Anonymous_NoConfiguredName_ResolvesFromCurrentOrg
// verifies that anonymous access (no token, no basic auth) gets the same
// resolution as token auth -- the gap that let an anonymous context stay
// stuck reporting "unknown" forever, since the old code only checked
// GetAPIToken() != "" and anonymous access has none.
func TestInitOrganizations_Anonymous_NoConfiguredName_ResolvesFromCurrentOrg(t *testing.T) {
	server := orgHandler(t, 2, "Public")

	svc := newTestDashNGo(t, server.URL, "")
	assert.Equal(t, "unknown", svc.grafanaConf.GetOrganizationName(), "sanity: unresolved anonymous context starts at unknown")

	svc.InitOrganizations()

	assert.Equal(t, "Public", svc.grafanaConf.GetOrganizationName())
}

// TestInitOrganizations_TokenAuth_ApiOrgUnavailable_KeepsConfiguredName
// verifies the fallback when /api/org itself can't be reached: GDG must not
// crash or silently blank out an already-configured name, just warn and
// leave the config as-is. This is also what stops the new always-verify
// behavior from turning a transient network hiccup into a hard crash at
// startup -- resolvePinnedOrgIdentity deliberately avoids
// GetTokenOrganization()/getAssociatedActiveOrg(), which log.Fatal on any
// /api/org error.
func TestInitOrganizations_TokenAuth_ApiOrgUnavailable_KeepsConfiguredName(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/org", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	svc := newTestDashNGo(t, server.URL, "glsa_faketoken")
	svc.grafanaConf.OrganizationName = "Configured Org"

	svc.InitOrganizations()

	assert.Equal(t, "Configured Org", svc.grafanaConf.GetOrganizationName(), "an unreachable /api/org must not wipe out an already-configured name")
}
