// config_crud_helpers_test.go exercises the pure-logic helper functions defined
// in config_crud.go.  Because these functions are unexported the test file uses
// package config (white-box) rather than package config_test.
package config

import (
	"testing"

	"github.com/esnet/gdg/internal/config/config_domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── looksLikeRegex ────────────────────────────────────────────────────────────

func TestLooksLikeRegex_MetaCharsReturnTrue(t *testing.T) {
	cases := []struct {
		input string
		want  bool
	}{
		{"linux.*", true},          // '*'
		{"foo?bar", true},          // '?'
		{"[0-9]+", true},           // '['
		{"(a|b)", true},            // '(' and '|'
		{"^start", true},           // '^'
		{"end$", true},             // '$'
		{"back\\slash", true},      // '\'
		{"My Folder", false},       // plain name with space — no metacharacters
		{"v1.0 Dashboards", false}, // dot is intentionally excluded
		{"General", false},         // clean name
		{"ES+net", false},          // literal plus — not a regex metacharacter
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			assert.Equal(t, tc.want, looksLikeRegex(tc.input))
		})
	}
}

// ── encodeFolderName ──────────────────────────────────────────────────────────

func TestEncodeFolderName_CleanNamePassesThrough(t *testing.T) {
	// "General" contains no characters that need encoding.
	assert.Equal(t, "General", encodeFolderName("General"))
}

func TestEncodeFolderName_SpacesEncoded(t *testing.T) {
	// url.QueryEscape converts space → '+'; QuoteMeta then escapes '+' → '\+'
	result := encodeFolderName("Linux Gnu")
	assert.Equal(t, `Linux\+Gnu`, result)
}

func TestEncodeFolderName_LiteralPlusEncoded(t *testing.T) {
	// url.QueryEscape encodes '+' as '%2B'; QuoteMeta leaves alphanumerics/% alone.
	result := encodeFolderName("ES+net")
	assert.Equal(t, "ES%2Bnet", result)
}

func TestEncodeFolderName_SlashSegmented(t *testing.T) {
	// EncodePath splits on the OS path separator ('/') and encodes each segment
	// independently, then rejoins with '/'.
	result := encodeFolderName("ES+net/LHC+Data")
	assert.Equal(t, "ES%2Bnet/LHC%2BData", result)
}

func TestEncodeFolderName_LeadingTrailingSpacesTrimmed(t *testing.T) {
	assert.Equal(t, "General", encodeFolderName("  General  "))
}

// ── testFolderRegexMatch ──────────────────────────────────────────────────────

func TestFolderRegexMatch_EmptyListNeverMatches(t *testing.T) {
	matched, patterns := testFolderRegexMatch([]string{}, "General")
	assert.False(t, matched)
	assert.Empty(t, patterns)
}

func TestFolderRegexMatch_LiteralPatternMatchesExact(t *testing.T) {
	// "General" stored as-is (no encoding needed); value "General" should match.
	matched, patterns := testFolderRegexMatch([]string{"General"}, "General")
	assert.True(t, matched)
	assert.Equal(t, []string{"General"}, patterns)
}

func TestFolderRegexMatch_EncodedPatternMatchesRawValue(t *testing.T) {
	// Store an encoded folder name as the pattern, then test against the raw name.
	// encodeFolderName("Linux Gnu") → "Linux\+Gnu"
	// testFolderRegexMatch URL-encodes "Linux Gnu" → "Linux+Gnu",
	// then regexp "Linux\+Gnu" matches "Linux+Gnu" (literal +).
	stored := encodeFolderName("Linux Gnu")
	matched, _ := testFolderRegexMatch([]string{stored}, "Linux Gnu")
	assert.True(t, matched)
}

func TestFolderRegexMatch_NoMatchReturnsEmpty(t *testing.T) {
	matched, patterns := testFolderRegexMatch([]string{"General"}, "Metrics")
	assert.False(t, matched)
	assert.Empty(t, patterns)
}

func TestFolderRegexMatch_RegexPatternMatchesMultiple(t *testing.T) {
	// A wildcard pattern should match multiple values.
	pattern := ".*"
	folders := []string{pattern}
	matched, hits := testFolderRegexMatch(folders, "anything")
	assert.True(t, matched)
	assert.Contains(t, hits, pattern)
}

func TestFolderRegexMatch_BadPatternSkipped(t *testing.T) {
	// A pattern that does not compile should be silently ignored.
	bad := "[unclosed"
	matched, patterns := testFolderRegexMatch([]string{bad, "General"}, "General")
	assert.True(t, matched)
	assert.NotContains(t, patterns, bad)
}

func TestFolderRegexMatch_MultiplePatternsSomeMatch(t *testing.T) {
	folders := []string{encodeFolderName("Linux Gnu"), "General", encodeFolderName("Metrics")}
	matched, hits := testFolderRegexMatch(folders, "Linux Gnu")
	assert.True(t, matched)
	assert.Len(t, hits, 1)
}

// ── validateFilter ────────────────────────────────────────────────────────────

func TestValidateFilter_BlankFieldReturnsError(t *testing.T) {
	_, err := validateFilter("", ".*", true)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "required")
}

func TestValidateFilter_BlankRegexReturnsError(t *testing.T) {
	_, err := validateFilter("name", "", true)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "required")
}

func TestValidateFilter_InvalidRegexReturnsError(t *testing.T) {
	_, err := validateFilter("name", "[bad", true)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid regex")
}

func TestValidateFilter_ValidInclusiveRule(t *testing.T) {
	rule, err := validateFilter("name", "prod-.*", true)
	require.NoError(t, err)
	require.NotNil(t, rule)
	assert.Equal(t, "name", rule.Field)
	assert.Equal(t, "prod-.*", rule.Regex)
	assert.True(t, rule.Inclusive)
}

func TestValidateFilter_ValidExclusiveRule(t *testing.T) {
	rule, err := validateFilter("type", "mysql", false)
	require.NoError(t, err)
	require.NotNil(t, rule)
	assert.False(t, rule.Inclusive)
}

func TestValidateFilter_WhitespaceIsTrimmed(t *testing.T) {
	rule, err := validateFilter("  name  ", "  prod-.*  ", true)
	require.NoError(t, err)
	assert.Equal(t, "name", rule.Field)
	assert.Equal(t, "prod-.*", rule.Regex)
}

// ── validateCredentialRule ────────────────────────────────────────────────────

func TestValidateCredentialRule_AlwaysExclusive(t *testing.T) {
	rule, err := validateCredentialRule("name", ".*")
	require.NoError(t, err)
	assert.False(t, rule.Inclusive, "credential rules must always be exclusive (Inclusive=false)")
}

func TestValidateCredentialRule_ReturnsErrorForBlankField(t *testing.T) {
	_, err := validateCredentialRule("", ".*")
	require.Error(t, err)
}

// ── testRegexMatch ────────────────────────────────────────────────────────────

func TestTestRegexMatch_MatchingValueReturnsTrue(t *testing.T) {
	matched, err := testRegexMatch("prod-.*", "prod-grafana")
	require.NoError(t, err)
	assert.True(t, matched)
}

func TestTestRegexMatch_NonMatchingValueReturnsFalse(t *testing.T) {
	matched, err := testRegexMatch("prod-.*", "staging-grafana")
	require.NoError(t, err)
	assert.False(t, matched)
}

func TestTestRegexMatch_InvalidRegexReturnsError(t *testing.T) {
	_, err := testRegexMatch("[bad", "anything")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid regex")
}

func TestTestRegexMatch_LeadingTrailingSpacesTrimmedFromValue(t *testing.T) {
	matched, err := testRegexMatch("^grafana$", "  grafana  ")
	require.NoError(t, err)
	// After TrimSpace the value becomes "grafana" and should match.
	assert.True(t, matched)
}

func TestTestRegexMatch_CatchAllMatchesAnything(t *testing.T) {
	matched, err := testRegexMatch(".*", "any-datasource-name")
	require.NoError(t, err)
	assert.True(t, matched)
}

// ── appendDefaultCredentialRule ───────────────────────────────────────────────

func TestAppendDefaultCredentialRule_AppendsToEmptySlice(t *testing.T) {
	result := appendDefaultCredentialRule(nil)
	require.Len(t, result, 1)
	r := result[0]
	require.Len(t, r.Rules, 1)
	assert.Equal(t, "name", r.Rules[0].Field)
	assert.Equal(t, ".*", r.Rules[0].Regex)
	assert.Equal(t, "default.yaml", r.SecureData)
}

func TestAppendDefaultCredentialRule_AppendsWhenDefaultMissing(t *testing.T) {
	existing := []*config_domain.RegexMatchesList{
		{
			Rules:      []config_domain.MatchingRule{{Field: "type", Regex: "mysql"}},
			SecureData: "mysql-creds.yaml",
		},
	}
	result := appendDefaultCredentialRule(existing)
	assert.Len(t, result, 2)
	last := result[len(result)-1]
	assert.Equal(t, ".*", last.Rules[0].Regex)
}

func TestAppendDefaultCredentialRule_IdempotentWhenDefaultPresent(t *testing.T) {
	existing := []*config_domain.RegexMatchesList{
		{
			Rules:      []config_domain.MatchingRule{{Field: "name", Regex: ".*"}},
			SecureData: "default.yaml",
		},
	}
	result := appendDefaultCredentialRule(existing)
	// Should not grow.
	assert.Len(t, result, 1, "should not append a second default rule")
}

func TestAppendDefaultCredentialRule_DefaultIsAlwaysLast(t *testing.T) {
	existing := []*config_domain.RegexMatchesList{
		{
			Rules:      []config_domain.MatchingRule{{Field: "type", Regex: "postgres"}},
			SecureData: "pg-creds.yaml",
		},
	}
	result := appendDefaultCredentialRule(existing)
	last := result[len(result)-1]
	assert.Equal(t, ".*", last.Rules[0].Regex, "catch-all should always be the last rule")
}

// ── summariseFilters ──────────────────────────────────────────────────────────

func TestSummariseFilters_EmptySliceReturnsNone(t *testing.T) {
	assert.Equal(t, "none", summariseFilters(nil))
	assert.Equal(t, "none", summariseFilters([]config_domain.MatchingRule{}))
}

func TestSummariseFilters_InclusiveRule(t *testing.T) {
	filters := []config_domain.MatchingRule{
		{Field: "name", Regex: "prod-.*", Inclusive: true},
	}
	assert.Equal(t, "name/prod-.*(incl)", summariseFilters(filters))
}

func TestSummariseFilters_ExclusiveRule(t *testing.T) {
	filters := []config_domain.MatchingRule{
		{Field: "type", Regex: "mysql", Inclusive: false},
	}
	assert.Equal(t, "type/mysql(excl)", summariseFilters(filters))
}

func TestSummariseFilters_MultipleRulesJoined(t *testing.T) {
	filters := []config_domain.MatchingRule{
		{Field: "name", Regex: "prod-.*", Inclusive: true},
		{Field: "type", Regex: "mysql", Inclusive: false},
	}
	result := summariseFilters(filters)
	assert.Contains(t, result, "name/prod-.*")
	assert.Contains(t, result, "type/mysql")
	// Multiple entries should be comma-separated.
	assert.Contains(t, result, ", ")
}

// ── validateGrafanaURL ────────────────────────────────────────────────────────

func TestValidateGrafanaURL_ValidHTTP(t *testing.T) {
	assert.NoError(t, validateGrafanaURL("http://grafana.example.com"))
}

func TestValidateGrafanaURL_ValidHTTPS(t *testing.T) {
	assert.NoError(t, validateGrafanaURL("https://grafana.example.com:3000"))
}

func TestValidateGrafanaURL_ValidLocalhost(t *testing.T) {
	assert.NoError(t, validateGrafanaURL("http://localhost:3000"))
}

func TestValidateGrafanaURL_EmptyReturnsError(t *testing.T) {
	err := validateGrafanaURL("")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "required")
}

func TestValidateGrafanaURL_WhitespaceOnlyReturnsError(t *testing.T) {
	err := validateGrafanaURL("   ")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "required")
}

func TestValidateGrafanaURL_MissingSchemeReturnsError(t *testing.T) {
	err := validateGrafanaURL("grafana.example.com")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "http")
}

func TestValidateGrafanaURL_WrongSchemeReturnsError(t *testing.T) {
	err := validateGrafanaURL("ftp://grafana.example.com")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "http")
}

func TestValidateGrafanaURL_SchemeOnlyReturnsError(t *testing.T) {
	err := validateGrafanaURL("http://")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "host")
}

func TestValidateGrafanaURL_LeadingTrailingSpacesAccepted(t *testing.T) {
	// Whitespace is trimmed before validation; the URL itself is valid.
	assert.NoError(t, validateGrafanaURL("  http://localhost:3000  "))
}
