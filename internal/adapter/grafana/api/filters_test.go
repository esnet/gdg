package api

// Unit tests for the filter constructors in the api package:
// NewTeamFilter, NewUserFilter, NewOrganizationFilter, NewLibraryElementFilter.
//
// Each constructor calls a setupXxxReaders helper that registers typed readers
// with v2.BaseFilter.  These tests exercise the registration + validation
// logic without any network calls by using plain model structs as input.

import (
	"context"
	"testing"

	"github.com/esnet/gdg/internal/config/config_domain"
	"github.com/esnet/gdg/internal/domain"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// NewTeamFilter
// ---------------------------------------------------------------------------

func TestNewTeamFilter_NonNil(t *testing.T) {
	f := NewTeamFilter("")
	require.NotNil(t, f)
}

func TestNewTeamFilter_EmptyName_AcceptsAnyTeam(t *testing.T) {
	f := NewTeamFilter("")
	ctx := context.Background()
	name := "any-team"
	team := models.TeamDTO{Name: &name}
	// Empty expected value → validation passes for everything.
	assert.True(t, f.Validate(ctx, domain.Name, team))
}

func TestNewTeamFilter_MatchingTeamName_ReturnsTrue(t *testing.T) {
	f := NewTeamFilter("ops")
	ctx := context.Background()
	name := "ops"
	team := models.TeamDTO{Name: &name}
	assert.True(t, f.Validate(ctx, domain.Name, team))
}

func TestNewTeamFilter_NonMatchingTeamName_ReturnsFalse(t *testing.T) {
	f := NewTeamFilter("ops")
	ctx := context.Background()
	name := "dev"
	team := models.TeamDTO{Name: &name}
	assert.False(t, f.Validate(ctx, domain.Name, team))
}

func TestNewTeamFilter_JSONReader_MatchingName(t *testing.T) {
	f := NewTeamFilter("my-team")
	ctx := context.Background()
	raw := []byte(`{"name":"my-team"}`)
	assert.True(t, f.Validate(ctx, domain.Name, raw))
}

func TestNewTeamFilter_JSONReader_NonMatchingName(t *testing.T) {
	f := NewTeamFilter("my-team")
	ctx := context.Background()
	raw := []byte(`{"name":"other-team"}`)
	assert.False(t, f.Validate(ctx, domain.Name, raw))
}

// ---------------------------------------------------------------------------
// NewUserFilter
// ---------------------------------------------------------------------------

func TestNewUserFilter_NonNil(t *testing.T) {
	f := NewUserFilter("")
	require.NotNil(t, f)
}

func TestNewUserFilter_EmptyLabel_AcceptsAnyUser(t *testing.T) {
	f := NewUserFilter("")
	ctx := context.Background()
	user := models.UserSearchHitDTO{Login: "alice", AuthLabels: []string{"ldap"}}
	// Empty label list → validation passes.
	assert.True(t, f.Validate(ctx, domain.AuthLabel, user))
}

func TestNewUserFilter_MatchingLabel_ReturnsTrue(t *testing.T) {
	f := NewUserFilter("ldap")
	ctx := context.Background()
	user := models.UserSearchHitDTO{Login: "alice", AuthLabels: []string{"ldap"}}
	assert.True(t, f.Validate(ctx, domain.AuthLabel, user))
}

func TestNewUserFilter_NonMatchingLabel_ReturnsFalse(t *testing.T) {
	f := NewUserFilter("ldap")
	ctx := context.Background()
	user := models.UserSearchHitDTO{Login: "bob", AuthLabels: []string{"saml"}}
	assert.False(t, f.Validate(ctx, domain.AuthLabel, user))
}

func TestNewUserFilter_JSONReader_MatchingLabel(t *testing.T) {
	f := NewUserFilter("ldap")
	ctx := context.Background()
	raw := []byte(`{"authLabels":["ldap","saml"]}`)
	assert.True(t, f.Validate(ctx, domain.AuthLabel, raw))
}

func TestNewUserFilter_JSONReader_MissingLabels_ReturnsFalse(t *testing.T) {
	f := NewUserFilter("ldap")
	ctx := context.Background()
	// authLabels key absent — reader returns an error → Validate returns false.
	raw := []byte(`{"login":"charlie"}`)
	assert.False(t, f.Validate(ctx, domain.AuthLabel, raw))
}

// ---------------------------------------------------------------------------
// NewOrganizationFilter
// ---------------------------------------------------------------------------

func TestNewOrganizationFilter_NonNil(t *testing.T) {
	f := NewOrganizationFilter()
	require.NotNil(t, f)
}

func TestNewOrganizationFilter_NoArgs_AcceptsAnyOrg(t *testing.T) {
	f := NewOrganizationFilter()
	ctx := context.Background()
	org := models.OrgDTO{Name: "any org"}
	// No filter arg → no AddValidation called → ValidateAll is always true.
	assert.True(t, f.ValidateAll(ctx, org))
}

func TestNewOrganizationFilter_MatchingName_ReturnsTrue(t *testing.T) {
	f := NewOrganizationFilter("Main Org.")
	ctx := context.Background()
	org := models.OrgDTO{Name: "Main Org."}
	assert.True(t, f.Validate(ctx, domain.OrgFilter, org))
}

func TestNewOrganizationFilter_NonMatchingName_ReturnsFalse(t *testing.T) {
	f := NewOrganizationFilter("Main Org.")
	ctx := context.Background()
	org := models.OrgDTO{Name: "Secondary Org"}
	assert.False(t, f.Validate(ctx, domain.OrgFilter, org))
}

func TestNewOrganizationFilter_JSONReader_MatchingName(t *testing.T) {
	f := NewOrganizationFilter("Main Org.")
	ctx := context.Background()
	raw := []byte(`{"organization":{"name":"Main Org."}}`)
	assert.True(t, f.Validate(ctx, domain.OrgFilter, raw))
}

func TestNewOrganizationFilter_JSONReader_NonMatchingName(t *testing.T) {
	f := NewOrganizationFilter("Main Org.")
	ctx := context.Background()
	raw := []byte(`{"organization":{"name":"Other Org"}}`)
	assert.False(t, f.Validate(ctx, domain.OrgFilter, raw))
}

func TestNewOrganizationFilter_JSONReader_MissingField_ReturnsFalse(t *testing.T) {
	f := NewOrganizationFilter("Main Org.")
	ctx := context.Background()
	// "organization.name" path absent → reader returns error → false.
	raw := []byte(`{"something":"else"}`)
	assert.False(t, f.Validate(ctx, domain.OrgFilter, raw))
}

// ---------------------------------------------------------------------------
// NewLibraryElementFilter
// ---------------------------------------------------------------------------

func TestNewLibraryElementFilter_NonNil(t *testing.T) {
	cfg := &config_domain.GDGAppConfiguration{}
	cfg.Contexts = map[string]*config_domain.GrafanaConfig{
		"default": {MonitoredFolders: []string{"General"}},
	}
	cfg.ContextName = "default"
	f := NewLibraryElementFilter(cfg)
	require.NotNil(t, f)
}

func TestNewLibraryElementFilter_WithNestedReader_FolderPath(t *testing.T) {
	// Build a minimal config with one monitored folder so folder-filter passes.
	cfg := &config_domain.GDGAppConfiguration{}
	cfg.Contexts = map[string]*config_domain.GrafanaConfig{
		"default": {MonitoredFolders: []string{"General"}},
	}
	cfg.ContextName = "default"

	f := NewLibraryElementFilter(cfg)
	ctx := context.Background()

	// The WithNested reader extracts NestedPath for domain.FolderFilter.
	entry := &domain.WithNested[models.LibraryElementDTO]{
		Entity:     &models.LibraryElementDTO{Name: "My Panel"},
		NestedPath: "General",
	}
	// Validate against the FolderFilter type — should pass for "General".
	assert.True(t, f.Validate(ctx, domain.FolderFilter, entry))
}

func TestNewLibraryElementFilter_MapReader_FolderPath(t *testing.T) {
	cfg := &config_domain.GDGAppConfiguration{}
	cfg.Contexts = map[string]*config_domain.GrafanaConfig{
		"default": {MonitoredFolders: []string{"General"}},
	}
	cfg.ContextName = "default"

	f := NewLibraryElementFilter(cfg)
	ctx := context.Background()

	m := map[string]any{NestedDashFolderName: "General"}
	assert.True(t, f.Validate(ctx, domain.FolderFilter, m))
}


