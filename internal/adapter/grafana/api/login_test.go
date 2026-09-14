package api

// Unit tests for DashNGoImpl.Login's lookup-resolution wiring. These exercise
// only the lookup-related branch added to Login(); the BasicAuth branch is
// avoided by ensuring GetPassword() is empty (via an explicit SecureModel
// override), so no network call is made.

import (
	"testing"

	"github.com/esnet/gdg/internal/config/config_domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeLookupResolver is a minimal outbound.LookupResolver test double that
// records whether it was invoked and returns a fixed value.
type fakeLookupResolver struct {
	called bool
	value  string
	err    error
}

func (f *fakeLookupResolver) Resolve(raw string) (string, error) {
	f.called = true
	if f.err != nil {
		return "", f.err
	}
	return f.value, nil
}

func TestLogin_ResolvesLookupsWhenResolverPresent(t *testing.T) {
	svc := newTestSvc(t)
	require.NoError(t, svc.grafanaConf.TestSetSecureAuth(config_domain.SecureModel{
		Token: "lookup:gsm:projects/1/secrets/x/versions/1",
	}))
	resolver := &fakeLookupResolver{value: "resolved-token"}
	svc.lookupResolver = resolver

	svc.Login()

	assert.True(t, resolver.called, "Login should invoke the configured lookup resolver")
	assert.Equal(t, "resolved-token", svc.grafanaConf.GetAPIToken())
}

func TestLogin_NilResolverSkipsResolution(t *testing.T) {
	svc := newTestSvc(t)
	require.NoError(t, svc.grafanaConf.TestSetSecureAuth(config_domain.SecureModel{
		Token: "lookup:gsm:projects/1/secrets/x/versions/1",
	}))
	svc.lookupResolver = nil

	svc.Login()

	assert.Equal(t, "lookup:gsm:projects/1/secrets/x/versions/1", svc.grafanaConf.GetAPIToken(),
		"with no resolver configured, the lookup reference must be left untouched")
}

func TestLogin_ResolverPresentButNoLookupRefsNoChange(t *testing.T) {
	svc := newTestSvc(t)
	require.NoError(t, svc.grafanaConf.TestSetSecureAuth(config_domain.SecureModel{
		Token: "plain-token",
	}))
	resolver := &fakeLookupResolver{value: "should-not-be-used"}
	svc.lookupResolver = resolver

	svc.Login()

	assert.False(t, resolver.called, "resolver should not be invoked when there are no lookup refs")
	assert.Equal(t, "plain-token", svc.grafanaConf.GetAPIToken())
}
