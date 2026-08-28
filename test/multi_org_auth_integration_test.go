package test

// Closes BUG_FIX_TODO.md gap #1 ("Multi-org token/anonymous integration
// test"), gap #2 ("No CI leg for anonymous auth"), and gap #3 ("Upload/delete
// not exercised against a non-default org namespace"). All three are the
// exact scenario that let the original App Platform namespace bugs ship:
// every other integration test operates against the default org (org 1 /
// "default"), which is precisely the case the old
// k8sNamespace()/InitOrganizations() bugs could never surface, since a
// broken org lookup that silently falls back to "default" is invisible when
// "default" is also the correct answer.
//
// These tests build a real multi-org Grafana container (via the existing
// UploadOrganizations fixture machinery in test_tooling), then pin a client
// to a NON-default org ("testing", org id 4) two different ways:
//   - a service-account token, created while the admin client's org header
//     is pinned to "testing" (CreateServiceAccount/CreateServiceAccountToken
//     have no orgId parameter -- for basic auth, the org a service account
//     is created in is whatever GetDefaultGrafanaConfig().OrganizationName
//     resolves to via the X-Grafana-Org-Id header baseService attaches to
//     every basic-auth request; it is NOT influenced by SetUserOrganizations/
//     UserSetUsingOrg, which only changes the user's server-side "current
//     org" pointer, not the header-driven org context basic-auth API calls
//     actually use)
//   - anonymous access, pinned to "testing" via GF_AUTH_ANONYMOUS_ORG_NAME at
//     container start (the org itself is created afterwards, server-side --
//     Grafana resolves GF_AUTH_ANONYMOUS_ORG_NAME by name per-request, not
//     just at boot)
//
// Each test then calls InitOrganizations() (exercising resolvePinnedOrgIdentity)
// and lists/uploads/deletes dashboards and/or folders (exercising
// k8sNamespace() and the folder equivalent), and cross-checks against the
// admin's own default-org view to prove genuine namespace isolation rather
// than a coincidental match.
//
// On gap #2 specifically: no changes to .github/workflows/go.yml were
// needed. Its matrix already runs the whole ./... suite once per
// (grafana_version, use_tokens) cell, and every test here calls
// test_tooling.SkipTokenBasedTests(t) up front (same reason
// TestOrganizationCrud skips under token config -- InitTest returns a
// token, not admin, client when TEST_TOKEN_CONFIG=1, and these tests need a
// basic-auth admin session to bootstrap orgs/service-accounts). So they
// already run automatically whenever use_tokens=0, on whichever Grafana
// versions their own version gate allows -- no separate "anonymous" matrix
// dimension is needed, since anonymous is built by hand inside the test
// body rather than selected via TEST_TOKEN_CONFIG.
//
// The dashboard-scoped tests are gated to v13+ (skipLegacyVersion) because
// they exercise the App Platform v2 dashboard API, which doesn't exist on
// v12. TestAnonymousAuth_MultiOrg_FoldersScopeToConfiguredOrg deliberately
// has no such gate: ListFolders/UploadFolders/DeleteAllFolders go through
// the plain /api/search endpoint, not the v2 App Platform client, so they
// have no v13 dependency -- that's what gives anonymous auth real coverage
// on v12 too.

import (
	"context"
	"log/slog"
	"testing"

	"github.com/esnet/gdg/internal/adapter/grafana/api"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/config/config_domain"
	"github.com/esnet/gdg/internal/ports/outbound"
	"github.com/esnet/gdg/pkg/test_tooling"
	"github.com/esnet/gdg/pkg/test_tooling/common"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
)

// pinAdminOrgContext points the basic-auth admin config at orgName (which
// drives the X-Grafana-Org-Id header baseService attaches to every
// basic-auth request -- see getOrgNameClientOpts in base_service.go) and
// rebuilds the client so the change takes effect, returning the rebuilt
// client and that org's ID. Pass "" to point back at the default org.
func pinAdminOrgContext(t *testing.T, cfg *config_domain.GDGAppConfiguration, container testcontainers.Container, orgName string) (outbound.GrafanaService, int64) {
	t.Helper()
	cfg.GetDefaultGrafanaConfig().OrganizationName = orgName
	rebuilt := test_tooling.CreateSimpleClientWithConfig(t, cfg, container)
	if orgName == "" {
		return rebuilt, config_domain.DefaultOrganizationId
	}
	orgs := rebuilt.ListOrganizations(api.NewOrganizationFilter(orgName), false)
	assert.Equal(t, 1, len(orgs), "expected exactly one org named %q", orgName)
	if len(orgs) != 1 {
		t.FailNow()
	}
	return rebuilt, orgs[0].Organization.ID
}

// buildAnonymousConfig returns a config pinned to no auth at all (no
// username/password, no token). config.NewConfig(common.DefaultTestConfig)
// bakes in the shared admin/admin basic-auth credentials used elsewhere in
// this suite, so they must be explicitly cleared -- same as InitTest does
// when building a token client from a basic-auth one.
func buildAnonymousConfig(t *testing.T) *config_domain.GDGAppConfiguration {
	t.Helper()
	anonCfg := config.NewConfig(common.DefaultTestConfig)
	anonGrafana := anonCfg.GetDefaultGrafanaConfig()
	anonGrafana.UserName = ""
	anonGrafana.Apply(config_domain.WithSecureAuth(config_domain.SecureModel{}))
	assert.False(t, anonGrafana.IsBasicAuth(), "sanity: anonymous client must not be basic auth")
	assert.Empty(t, anonGrafana.GetAPIToken(), "sanity: anonymous client must not carry a token")
	return anonCfg
}

// TestTokenAuth_MultiOrg_ScopesToConfiguredOrg exercises
// InitOrganizations()/k8sNamespace() end-to-end for a service-account token
// pinned to a non-default org, and closes gap #3 by driving upload+list+delete
// through that token against that org.
func TestTokenAuth_MultiOrg_ScopesToConfiguredOrg(t *testing.T) {
	skipLegacyVersion(t)
	// This test manages auth explicitly and needs a basic-auth admin session
	// to create orgs/service-accounts -- same reason TestOrganizationCrud
	// skips under token config (InitTest returns a token, not admin, client
	// when TEST_TOKEN_CONFIG=1).
	test_tooling.SkipTokenBasedTests(t)

	cfg := config.NewConfig(common.DefaultTestConfig)
	var r *test_tooling.InitContainerResult
	err := Retry(context.Background(), DefaultRetryAttempts, func() error {
		r = test_tooling.InitTest(t, cfg, nil)
		return r.Err
	})
	assert.NotNil(t, r)
	assert.NoError(t, err)
	defer func() {
		if cleanupErr := r.CleanUp(); cleanupErr != nil {
			slog.Warn("Unable to clean up after test", "test", t.Name())
		}
	}()
	apiClient := r.ApiClient

	newOrgs := apiClient.UploadOrganizations(api.NewOrganizationFilter())
	assert.Contains(t, newOrgs, "testing")

	apiClient, _ = pinAdminOrgContext(t, cfg, r.Container, "testing")

	serviceName, _ := uuid.NewUUID()
	serviceAccnt, err := apiClient.CreateServiceAccount(serviceName.String(), "admin", 0)
	assert.NoError(t, err, "unable to create service account scoped to the 'testing' org")
	newKey, err := apiClient.CreateServiceAccountToken(serviceAccnt.ID, "admin", 0)
	assert.NoError(t, err)

	// Point the admin client back at the default org so the isolation check
	// below reflects org 1, not "testing".
	apiClient, _ = pinAdminOrgContext(t, cfg, r.Container, "")

	// Build a fresh client authenticated with the token only (no username/password),
	// pinned to whatever org the token itself resolves to.
	tokenCfg := config.NewConfig(common.DefaultTestConfig)
	tokenGrafana := tokenCfg.GetDefaultGrafanaConfig()
	tokenGrafana.UserName = ""
	tokenGrafana.Apply(config_domain.WithSecureAuth(config_domain.SecureModel{Token: newKey.Key}))
	assert.False(t, tokenGrafana.IsBasicAuth(), "sanity: token client must not be basic auth")
	tokenClient := test_tooling.CreateSimpleClientWithConfig(t, tokenCfg, r.Container)

	// Exercises resolvePinnedOrgIdentity() against a real, non-default org.
	tokenClient.InitOrganizations()
	assert.Equal(t, "testing", tokenGrafana.GetOrganizationName(),
		"token pinned to a non-default org must resolve its real org name via InitOrganizations, not fall back to a default")

	filtersEntity := api.NewDashboardFilter(tokenCfg, "", "", "")

	// Exercises k8sNamespace() (upload path) against org-2, closing gap #3.
	uploadedFiles, err := tokenClient.UploadDashboards(filtersEntity)
	assert.NoError(t, err)
	assert.NotEmpty(t, uploadedFiles, "expected the 'testing' org fixture dashboards to upload")

	// Exercises k8sNamespace() (list path) against org-2.
	boards := tokenClient.ListDashboards(filtersEntity)
	assert.Equal(t, len(uploadedFiles), len(boards))

	// Isolation check: the admin client, back in the default org, must not
	// see these dashboards. A namespace-resolution regression that silently
	// fell back to "default" for the token client would instead show these
	// dashboards landing in org 1, and this assertion would catch it.
	adminDefaultOrgBoards := apiClient.ListDashboards(api.NewDashboardFilter(cfg, "", "", ""))
	assert.Empty(t, adminDefaultOrgBoards, "dashboards uploaded via the org-scoped token must not be visible in the default org")

	// Exercises k8sNamespace() (delete path) against org-2, closing gap #3.
	deleteList := tokenClient.DeleteAllDashboards(filtersEntity)
	assert.Equal(t, len(uploadedFiles), len(deleteList))
	boards = tokenClient.ListDashboards(filtersEntity)
	assert.Empty(t, boards)
}

// TestAnonymousAuth_MultiOrg_ScopesToConfiguredOrg exercises
// InitOrganizations()/k8sNamespace() end-to-end for anonymous access pinned
// to a non-default org via GF_AUTH_ANONYMOUS_ORG_NAME -- the exact scenario
// (fix #4 in BUG_FIX_TODO.md) that previously either got stuck reporting
// "unknown" (InitOrganizations) or hit a hard "namespace mismatch" API error
// (k8sNamespace/listDashboardsV2) instead of resolving correctly.
//
// v13+ only (skipLegacyVersion): this exercises the App Platform v2
// dashboard API, which doesn't exist on v12.
func TestAnonymousAuth_MultiOrg_ScopesToConfiguredOrg(t *testing.T) {
	skipLegacyVersion(t)
	test_tooling.SkipTokenBasedTests(t)

	cfg := config.NewConfig(common.DefaultTestConfig)
	var r *test_tooling.InitContainerResult
	err := Retry(context.Background(), DefaultRetryAttempts, func() error {
		// GF_AUTH_ANONYMOUS_ORG_NAME is read by name per anonymous request,
		// not only at container boot, so it's safe to point it at an org
		// that doesn't exist yet -- it's created below, before any
		// anonymous request is made.
		r = test_tooling.InitTest(t, cfg, map[string]string{"GF_AUTH_ANONYMOUS_ORG_NAME": "testing"})
		return r.Err
	})
	assert.NotNil(t, r)
	assert.NoError(t, err)
	defer func() {
		if cleanupErr := r.CleanUp(); cleanupErr != nil {
			slog.Warn("Unable to clean up after test", "test", t.Name())
		}
	}()
	apiClient := r.ApiClient

	newOrgs := apiClient.UploadOrganizations(api.NewOrganizationFilter())
	assert.Contains(t, newOrgs, "testing")

	// Point the admin at "testing" too, so UploadDashboards reads the
	// org_testing fixture set and uploads into the org the admin's requests
	// are now scoped to.
	apiClient, _ = pinAdminOrgContext(t, cfg, r.Container, "testing")

	filtersEntity := api.NewDashboardFilter(cfg, "", "", "")
	uploadedFiles, err := apiClient.UploadDashboards(filtersEntity)
	assert.NoError(t, err)
	assert.NotEmpty(t, uploadedFiles, "expected the 'testing' org fixture dashboards to upload")

	anonCfg := buildAnonymousConfig(t)
	anonGrafana := anonCfg.GetDefaultGrafanaConfig()
	anonClient := test_tooling.CreateSimpleClientWithConfig(t, anonCfg, r.Container)

	// Exercises resolvePinnedOrgIdentity() for anonymous access against a
	// real, non-default org -- previously stuck reporting "unknown" forever.
	anonClient.InitOrganizations()
	assert.Equal(t, "testing", anonGrafana.GetOrganizationName(),
		"anonymous access pinned via GF_AUTH_ANONYMOUS_ORG_NAME must resolve its real org name via InitOrganizations")

	// Exercises k8sNamespace() (list path) for anonymous access against
	// org-2 -- previously either silently dropped everything (fix #1's bug
	// class) or hit a hard "namespace mismatch" error (fix #4's original
	// report) instead of returning the org's real dashboards.
	anonBoards := anonClient.ListDashboards(api.NewDashboardFilter(anonCfg, "", "", ""))
	assert.Equal(t, len(uploadedFiles), len(anonBoards))

	// Isolation check: point the admin back at the default org and confirm
	// these dashboards aren't visible there either -- proving the anonymous
	// client's view above came from genuine org-2 resolution, not a
	// coincidental default-org match.
	apiClient, _ = pinAdminOrgContext(t, cfg, r.Container, "")
	adminDefaultOrgBoards := apiClient.ListDashboards(api.NewDashboardFilter(cfg, "", "", ""))
	assert.Empty(t, adminDefaultOrgBoards, "dashboards uploaded into the 'testing' org must not be visible in the default org")
}

// TestAnonymousAuth_MultiOrg_FoldersScopeToConfiguredOrg closes gap #2 ("No
// CI leg for anonymous auth") for the case that actually runs on v12:
// folders. Unlike dashboards, ListFolders/UploadFolders/DeleteAllFolders go
// through the plain /api/search endpoint rather than the v13+ App Platform
// v2 client, so this test has no skipLegacyVersion gate and gives anonymous
// access real CI coverage on both Grafana versions the matrix runs
// (12.0.0-ubuntu and 13.2.0-ubuntu), whenever use_tokens=0.
func TestAnonymousAuth_MultiOrg_FoldersScopeToConfiguredOrg(t *testing.T) {
	test_tooling.SkipTokenBasedTests(t)

	cfg := config.NewConfig(common.DefaultTestConfig)
	var r *test_tooling.InitContainerResult
	err := Retry(context.Background(), DefaultRetryAttempts, func() error {
		r = test_tooling.InitTest(t, cfg, map[string]string{"GF_AUTH_ANONYMOUS_ORG_NAME": "testing"})
		return r.Err
	})
	assert.NotNil(t, r)
	assert.NoError(t, err)
	defer func() {
		if cleanupErr := r.CleanUp(); cleanupErr != nil {
			slog.Warn("Unable to clean up after test", "test", t.Name())
		}
	}()
	apiClient := r.ApiClient

	newOrgs := apiClient.UploadOrganizations(api.NewOrganizationFilter())
	assert.Contains(t, newOrgs, "testing")

	// Point the admin at "testing" so UploadFolders reads the org_testing
	// fixture set (4 folders -- the same fixture TestFolderNestedCRUD uses
	// for basic auth) and uploads into the org the admin's requests are now
	// scoped to.
	apiClient, _ = pinAdminOrgContext(t, cfg, r.Container, "testing")
	uploadedFolders := apiClient.UploadFolders(nil)
	assert.NotEmpty(t, uploadedFolders, "expected the 'testing' org fixture folders to upload")

	anonCfg := buildAnonymousConfig(t)
	anonGrafana := anonCfg.GetDefaultGrafanaConfig()
	anonClient := test_tooling.CreateSimpleClientWithConfig(t, anonCfg, r.Container)

	// Exercises resolvePinnedOrgIdentity() for anonymous access against a
	// real, non-default org, on a Grafana version where the App Platform v2
	// dashboard client isn't even in play.
	anonClient.InitOrganizations()
	assert.Equal(t, "testing", anonGrafana.GetOrganizationName(),
		"anonymous access pinned via GF_AUTH_ANONYMOUS_ORG_NAME must resolve its real org name via InitOrganizations")

	// Exercises the folder-listing namespace resolution for anonymous access
	// against org-2. Anonymous access defaults to the Viewer role, so this
	// only exercises list, not upload/delete -- an anonymous client can't
	// write folders without GF_AUTH_ANONYMOUS_ORG_ROLE being raised, which
	// this test deliberately doesn't do (matching real-world anonymous
	// deployments, which are read-only).
	anonFolders := anonClient.ListFolders(nil)
	assert.Equal(t, len(uploadedFolders), len(anonFolders))

	// Isolation check: point the admin back at the default org and confirm
	// these folders aren't visible there either.
	apiClient, _ = pinAdminOrgContext(t, cfg, r.Container, "")
	adminDefaultOrgFolders := apiClient.ListFolders(nil)
	assert.Empty(t, adminDefaultOrgFolders, "folders uploaded into the 'testing' org must not be visible in the default org")
}
