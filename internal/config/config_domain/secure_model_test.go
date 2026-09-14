package config_domain

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── SecureModel.Empty ─────────────────────────────────────────────────────────

func TestSecureModel_Empty_NilIsEmpty(t *testing.T) {
	var sm *SecureModel
	assert.True(t, sm.Empty())
}

func TestSecureModel_Empty_ZeroValueIsEmpty(t *testing.T) {
	sm := &SecureModel{}
	assert.True(t, sm.Empty())
}

func TestSecureModel_Empty_PasswordOnlyIsNotEmpty(t *testing.T) {
	sm := &SecureModel{Password: "secret"}
	assert.False(t, sm.Empty())
}

func TestSecureModel_Empty_TokenOnlyIsNotEmpty(t *testing.T) {
	sm := &SecureModel{Token: "mytoken"}
	assert.False(t, sm.Empty())
}

func TestSecureModel_Empty_BothSetIsNotEmpty(t *testing.T) {
	sm := &SecureModel{Password: "pw", Token: "tok"}
	assert.False(t, sm.Empty())
}

// ── SecureModel.UpdateSecureModel ─────────────────────────────────────────────

func TestUpdateSecureModel_TransformsTokenAndPassword(t *testing.T) {
	sm := &SecureModel{Password: "plain-pass", Token: "plain-token"}
	prefix := func(s string) (string, error) {
		return "enc:" + s, nil
	}
	sm.UpdateSecureModel(prefix, stubLookupSvc{})
	assert.Equal(t, "enc:plain-pass", sm.Password)
	assert.Equal(t, "enc:plain-token", sm.Token)
}

func TestUpdateSecureModel_ErrorLeavesFieldUnchanged(t *testing.T) {
	sm := &SecureModel{Password: "original", Token: "tok"}
	failFn := func(s string) (string, error) {
		return "", errors.New("encode failed")
	}
	sm.UpdateSecureModel(failFn, stubLookupSvc{})
	// Both fields should be unchanged after errors
	assert.Equal(t, "original", sm.Password)
	assert.Equal(t, "tok", sm.Token)
}

func TestUpdateSecureModel_EmptyTokenSkipped(t *testing.T) {
	sm := &SecureModel{Password: "pw", Token: ""}
	called := 0
	sm.UpdateSecureModel(func(s string) (string, error) {
		called++
		return "x", nil
	}, stubLookupSvc{})
	// Only Password is non-empty, so fn is called exactly once
	assert.Equal(t, 1, called)
	assert.Equal(t, "x", sm.Password)
}

func TestUpdateSecureModel_EmptyPasswordSkipped(t *testing.T) {
	sm := &SecureModel{Password: "", Token: "tok"}
	called := 0
	sm.UpdateSecureModel(func(s string) (string, error) {
		called++
		return "x", nil
	}, stubLookupSvc{})
	assert.Equal(t, 1, called)
	assert.Equal(t, "x", sm.Token)
}

func TestUpdateSecureModel_BothEmptyNoCalls(t *testing.T) {
	sm := &SecureModel{}
	called := 0
	sm.UpdateSecureModel(func(s string) (string, error) {
		called++
		return s, nil
	}, stubLookupSvc{})
	assert.Equal(t, 0, called)
}

func TestUpdateSecureModel_LookupRefTokenSkipped(t *testing.T) {
	sm := &SecureModel{Password: "pw", Token: "lookup:gsm:projects/1/secrets/x/versions/1"}
	called := 0
	sm.UpdateSecureModel(func(s string) (string, error) {
		called++
		return "encoded:" + s, nil
	}, stubLookupSvc{})
	// Token is a lookup reference, so it must be left untouched by UpdateSecureModel.
	assert.Equal(t, "lookup:gsm:projects/1/secrets/x/versions/1", sm.Token)
	assert.Equal(t, "encoded:pw", sm.Password)
	assert.Equal(t, 1, called)
}

func TestUpdateSecureModel_LookupRefPasswordSkipped(t *testing.T) {
	sm := &SecureModel{Password: "lookup:gsm:projects/1/secrets/x/versions/1", Token: "tok"}
	called := 0
	sm.UpdateSecureModel(func(s string) (string, error) {
		called++
		return "encoded:" + s, nil
	}, stubLookupSvc{})
	assert.Equal(t, "lookup:gsm:projects/1/secrets/x/versions/1", sm.Password)
	assert.Equal(t, "encoded:tok", sm.Token)
	assert.Equal(t, 1, called)
}

func TestUpdateSecureModel_BothLookupRefsNoCalls(t *testing.T) {
	sm := &SecureModel{Password: "lookup:gsm:a", Token: "lookup:gsm:b"}
	called := 0
	sm.UpdateSecureModel(func(s string) (string, error) {
		called++
		return s, nil
	}, stubLookupSvc{})
	assert.Equal(t, 0, called)
	assert.Equal(t, "lookup:gsm:a", sm.Password)
	assert.Equal(t, "lookup:gsm:b", sm.Token)
}

// ── SecureModel.ResolveLookups ────────────────────────────────────────────────

func TestResolveLookups_TransformsLookupRefsOnly(t *testing.T) {
	sm := &SecureModel{Password: "plain-pass", Token: "lookup:gsm:projects/1/secrets/x/versions/1"}
	called := 0
	sm.ResolveLookups(func(s string) (string, error) {
		called++
		return "resolved-value", nil
	}, stubLookupSvc{})
	// Token is a lookup ref, so it should be resolved; Password is plain and must be untouched.
	assert.Equal(t, "resolved-value", sm.Token)
	assert.Equal(t, "plain-pass", sm.Password)
	assert.Equal(t, 1, called)
}

func TestResolveLookups_ErrorLeavesFieldUnchanged(t *testing.T) {
	sm := &SecureModel{Token: "lookup:gsm:projects/1/secrets/x/versions/1"}
	sm.ResolveLookups(func(s string) (string, error) {
		return "", errors.New("resolve failed")
	}, stubLookupSvc{})
	assert.Equal(t, "lookup:gsm:projects/1/secrets/x/versions/1", sm.Token)
}

func TestResolveLookups_BothFieldsResolved(t *testing.T) {
	sm := &SecureModel{
		Password: "lookup:gsm:a",
		Token:    "lookup:gsm:b",
	}
	sm.ResolveLookups(func(s string) (string, error) {
		return "resolved:" + s, nil
	}, stubLookupSvc{})
	assert.Equal(t, "resolved:lookup:gsm:a", sm.Password)
	assert.Equal(t, "resolved:lookup:gsm:b", sm.Token)
}

func TestResolveLookups_NoLookupRefsNoCalls(t *testing.T) {
	sm := &SecureModel{Password: "plain-pass", Token: "plain-token"}
	called := 0
	sm.ResolveLookups(func(s string) (string, error) {
		called++
		return s, nil
	}, stubLookupSvc{})
	assert.Equal(t, 0, called)
	assert.Equal(t, "plain-pass", sm.Password)
	assert.Equal(t, "plain-token", sm.Token)
}

func TestResolveLookups_EmptyFieldsNoCalls(t *testing.T) {
	sm := &SecureModel{}
	called := 0
	sm.ResolveLookups(func(s string) (string, error) {
		called++
		return s, nil
	}, stubLookupSvc{})
	assert.Equal(t, 0, called)
}

// ── GrafanaConnection ─────────────────────────────────────────────────────────

func TestGrafanaConnection_UserAndPassword(t *testing.T) {
	conn := GrafanaConnection{
		"user":              "ds_user",
		"basicAuthPassword": "ds_pass",
	}
	assert.Equal(t, "ds_user", conn.User())
	assert.Equal(t, "ds_pass", conn.Password())
}

func TestGrafanaConnection_MissingKeysReturnEmpty(t *testing.T) {
	conn := GrafanaConnection{}
	assert.Equal(t, "", conn.User())
	assert.Equal(t, "", conn.Password())
}

// ── PluginEntity.SetPluginConfig ──────────────────────────────────────────────

func TestPluginEntity_SetPluginConfig_ResetsProcessed(t *testing.T) {
	pe := &PluginEntity{
		PluginConfig: map[string]string{"key": "old"},
		processed:    true,
	}
	pe.SetPluginConfig(map[string]string{"key": "new"})
	assert.False(t, pe.processed, "SetPluginConfig should mark entity as unprocessed")
	assert.Equal(t, "new", pe.PluginConfig["key"])
}

// ── SecureModelConfig.SecureFieldNames ────────────────────────────────────────

func TestSecureFieldNames_ReturnsSortedKeys(t *testing.T) {
	smc := &SecureModelConfig{
		SecureEntities: map[string]SecureEntity{
			"zebra": {Patterns: []string{".*"}},
			"alpha": {Patterns: []string{"a.*"}},
			"mango": {Patterns: []string{"m.*"}},
		},
	}
	names := smc.SecureFieldNames()
	require.Len(t, names, 3)
	assert.Equal(t, []string{"alpha", "mango", "zebra"}, names)
}

func TestSecureFieldNames_EmptyMapReturnsEmpty(t *testing.T) {
	smc := &SecureModelConfig{SecureEntities: map[string]SecureEntity{}}
	names := smc.SecureFieldNames()
	assert.Empty(t, names)
}
