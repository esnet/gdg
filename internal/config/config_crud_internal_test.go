// config_crud_internal_test.go tests unexported pure-logic helpers that live in
// config_crud.go. Using package config (white-box) gives access to unexported
// symbols while keeping the test file in the same package.
package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/esnet/gdg/internal/config/config_domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── encodeFolderName ──────────────────────────────────────────────────────────

func TestEncodeFolderName_Plain(t *testing.T) {
	// Plain ASCII names with no special chars should be returned unchanged.
	assert.Equal(t, "General", encodeFolderName("General"))
	assert.Equal(t, "MyDashboards", encodeFolderName("MyDashboards"))
}

func TestEncodeFolderName_Spaces(t *testing.T) {
	// Spaces are encoded as '+' by URL encoding, then the '+' is regex-escaped to '\+'.
	result := encodeFolderName("Linux Data")
	assert.Contains(t, result, "Linux")
	assert.Contains(t, result, "Data")
	// Should not contain a raw space.
	assert.NotContains(t, result, " ")
}

func TestEncodeFolderName_SpecialChars(t *testing.T) {
	// Ampersands must be encoded.
	result := encodeFolderName("k&r")
	assert.NotContains(t, result, "&")
	// EncodePath treats '/' as an OS path separator: it splits on '/', encodes
	// each component individually, then rejoins with '/'.  The slash itself is
	// therefore preserved — only the content between slashes is encoded.
	result2 := encodeFolderName("a/b")
	assert.Contains(t, result2, "/", "path separator should be preserved by EncodePath")
}

func TestEncodeFolderName_TrimsWhitespace(t *testing.T) {
	// Leading/trailing whitespace is trimmed before encoding.
	assert.Equal(t, encodeFolderName("General"), encodeFolderName("  General  "))
}

// ── looksLikeRegex ────────────────────────────────────────────────────────────

func TestLooksLikeRegex_LiteralNames(t *testing.T) {
	// Plain folder names must NOT be flagged as regexes.
	assert.False(t, looksLikeRegex("General"))
	assert.False(t, looksLikeRegex("Happy Gilmore"))
	assert.False(t, looksLikeRegex("v1.0 Dashboards")) // dot alone is not flagged
	assert.False(t, looksLikeRegex("Linux Data"))
	assert.False(t, looksLikeRegex("Production Metrics"))
}

func TestLooksLikeRegex_RegexPatterns(t *testing.T) {
	// Inputs containing regex metacharacters must be flagged.
	assert.True(t, looksLikeRegex("Stardust/.*"))  // '*' triggers detection
	assert.True(t, looksLikeRegex("DEV-.*"))        // '*'
	assert.True(t, looksLikeRegex("(prod|staging)")) // '(' and '|' and ')'
	assert.True(t, looksLikeRegex("[Dd]ev"))         // '['
	assert.True(t, looksLikeRegex("node_\\d+"))      // '\'
	assert.True(t, looksLikeRegex("^prod"))          // '^'
	assert.True(t, looksLikeRegex("prod$"))          // '$'
	assert.True(t, looksLikeRegex("node?"))          // '?'
}

func TestLooksLikeRegex_EdgeCases(t *testing.T) {
	// Empty string and whitespace-only are not regexes.
	assert.False(t, looksLikeRegex(""))
	assert.False(t, looksLikeRegex("   "))
	// Ampersand and percent are URL-special but not regex-special; not flagged.
	assert.False(t, looksLikeRegex("k&r"))
	assert.False(t, looksLikeRegex("100%"))
}

// ── testFolderRegexMatch ──────────────────────────────────────────────────────

func TestFolderRegexMatch_ExactMatch(t *testing.T) {
	folders := []string{"General"}
	matched, matches := testFolderRegexMatch(folders, "General")
	assert.True(t, matched)
	assert.Contains(t, matches, "General")
}

func TestFolderRegexMatch_NoMatch(t *testing.T) {
	folders := []string{"General", "Production"}
	matched, matches := testFolderRegexMatch(folders, "Staging")
	assert.False(t, matched)
	assert.Empty(t, matches)
}

func TestFolderRegexMatch_WildcardPattern(t *testing.T) {
	// A ".*" pattern matches everything.
	folders := []string{".*"}
	matched, _ := testFolderRegexMatch(folders, "AnyFolder")
	assert.True(t, matched)
}

func TestFolderRegexMatch_MultiplePatterns_OneMatches(t *testing.T) {
	folders := []string{"General", "Prod.*"}
	matched, matches := testFolderRegexMatch(folders, "Production")
	assert.True(t, matched)
	assert.Len(t, matches, 1)
	assert.Contains(t, matches, "Prod.*")
}

func TestFolderRegexMatch_MultiplePatterns_BothMatch(t *testing.T) {
	folders := []string{".*", "General"}
	matched, matches := testFolderRegexMatch(folders, "General")
	assert.True(t, matched)
	assert.Len(t, matches, 2)
}

func TestFolderRegexMatch_InvalidPattern_SkippedGracefully(t *testing.T) {
	// An invalid regex in the pattern list is silently skipped; valid ones still match.
	folders := []string{"[invalid", "General"}
	matched, matches := testFolderRegexMatch(folders, "General")
	assert.True(t, matched)
	assert.Contains(t, matches, "General")
}

func TestFolderRegexMatch_EmptyFolderList(t *testing.T) {
	matched, matches := testFolderRegexMatch([]string{}, "General")
	assert.False(t, matched)
	assert.Empty(t, matches)
}

func TestFolderRegexMatch_EncodesTestValue(t *testing.T) {
	// encodeFolderName("Linux Data") → "Linux\+Data"  (URL-encoded then regex-escaped)
	// testFolderRegexMatch URL-encodes the raw value to "Linux+Data" (no QuoteMeta),
	// so the stored pattern "Linux\+Data" (regex: literal '+') correctly matches it.
	encoded := encodeFolderName("Linux Data")
	folders := []string{encoded}
	matched, _ := testFolderRegexMatch(folders, "Linux Data")
	assert.True(t, matched, "URL-encoded test value should match the regex-escaped stored pattern")
}

func TestFolderRegexMatch_RawRegexPattern(t *testing.T) {
	// When the user stores a raw regex (looksLikeRegex → true, user picks "Yes"),
	// the pattern is kept verbatim.  testFolderRegexMatch should still work: it
	// URL-encodes the test value and the raw regex matches against it.
	//
	// "Stardust/.*" stored as-is:
	//   test value "Stardust/Metrics" → URL-encoded per component → "Stardust/Metrics"
	//   regex "Stardust/.*" compiled → matches "Stardust/Metrics" ✓
	folders := []string{"Stardust/.*"}
	matched, matches := testFolderRegexMatch(folders, "Stardust/Metrics")
	assert.True(t, matched, "raw regex pattern should match subfolders")
	assert.Contains(t, matches, "Stardust/.*")

	// A completely unrelated folder must not match.
	notMatched, _ := testFolderRegexMatch(folders, "General")
	assert.False(t, notMatched)
}

// ── validateFilter ────────────────────────────────────────────────────────────

func TestValidateFilter_ValidExclusive(t *testing.T) {
	rule, err := validateFilter("name", `DEV-.*`, false)
	require.NoError(t, err)
	require.NotNil(t, rule)
	assert.Equal(t, "name", rule.Field)
	assert.Equal(t, `DEV-.*`, rule.Regex)
	assert.False(t, rule.Inclusive)
}

func TestValidateFilter_ValidInclusive(t *testing.T) {
	rule, err := validateFilter("type", "elasticsearch", true)
	require.NoError(t, err)
	require.NotNil(t, rule)
	assert.True(t, rule.Inclusive)
}

func TestValidateFilter_EmptyField(t *testing.T) {
	_, err := validateFilter("", ".*", false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "required")
}

func TestValidateFilter_EmptyRegex(t *testing.T) {
	_, err := validateFilter("name", "", false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "required")
}

func TestValidateFilter_BothEmpty(t *testing.T) {
	_, err := validateFilter("  ", "  ", false)
	require.Error(t, err)
}

func TestValidateFilter_InvalidRegex(t *testing.T) {
	_, err := validateFilter("name", "[invalid", false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid regex")
}

func TestValidateFilter_TrimsWhitespace(t *testing.T) {
	rule, err := validateFilter("  name  ", "  .*  ", false)
	require.NoError(t, err)
	assert.Equal(t, "name", rule.Field)
	assert.Equal(t, ".*", rule.Regex)
}

// ── validateCredentialRule ────────────────────────────────────────────────────

func TestValidateCredentialRule_Valid(t *testing.T) {
	rule, err := validateCredentialRule("url", `.*esproxy.*`)
	require.NoError(t, err)
	require.NotNil(t, rule)
	assert.Equal(t, "url", rule.Field)
	assert.Equal(t, `.*esproxy.*`, rule.Regex)
	// Credential rules are never inclusive.
	assert.False(t, rule.Inclusive)
}

func TestValidateCredentialRule_EmptyField(t *testing.T) {
	_, err := validateCredentialRule("", ".*")
	require.Error(t, err)
}

func TestValidateCredentialRule_InvalidRegex(t *testing.T) {
	_, err := validateCredentialRule("name", "[bad")
	require.Error(t, err)
}

// ── testRegexMatch ────────────────────────────────────────────────────────────

func TestRegexMatch_Match(t *testing.T) {
	matched, err := testRegexMatch(`DEV-.*`, "DEV-frontend")
	require.NoError(t, err)
	assert.True(t, matched)
}

func TestRegexMatch_NoMatch(t *testing.T) {
	matched, err := testRegexMatch(`DEV-.*`, "PROD-frontend")
	require.NoError(t, err)
	assert.False(t, matched)
}

func TestRegexMatch_InvalidRegex(t *testing.T) {
	_, err := testRegexMatch(`[invalid`, "anything")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid regex")
}

func TestRegexMatch_TrimsTestValue(t *testing.T) {
	// Leading/trailing whitespace in value should be trimmed.
	matched, err := testRegexMatch(`General`, "  General  ")
	require.NoError(t, err)
	assert.True(t, matched)
}

func TestRegexMatch_CatchAllPattern(t *testing.T) {
	matched, err := testRegexMatch(`.*`, "anything_at_all")
	require.NoError(t, err)
	assert.True(t, matched)
}

// ── appendDefaultCredentialRule ───────────────────────────────────────────────

func TestAppendDefaultCredentialRule_EmptyInput(t *testing.T) {
	result := appendDefaultCredentialRule(nil)
	require.Len(t, result, 1)
	assert.Equal(t, "name", result[0].Rules[0].Field)
	assert.Equal(t, ".*", result[0].Rules[0].Regex)
	assert.Equal(t, "default.yaml", result[0].SecureData)
}

func TestAppendDefaultCredentialRule_AlreadyPresent(t *testing.T) {
	existing := []*config_domain.RegexMatchesList{
		{
			Rules:      []config_domain.MatchingRule{{Field: "name", Regex: ".*"}},
			SecureData: "default.yaml",
		},
	}
	result := appendDefaultCredentialRule(existing)
	// Must not duplicate the default rule.
	assert.Len(t, result, 1)
}

func TestAppendDefaultCredentialRule_AppendsAfterCustomRules(t *testing.T) {
	custom := []*config_domain.RegexMatchesList{
		{
			Rules:      []config_domain.MatchingRule{{Field: "name", Regex: "misc"}},
			SecureData: "special.yaml",
		},
		{
			Rules:      []config_domain.MatchingRule{{Field: "url", Regex: `.*esproxy.*`}},
			SecureData: "proxy.yaml",
		},
	}
	result := appendDefaultCredentialRule(custom)
	require.Len(t, result, 3)
	last := result[len(result)-1]
	assert.Equal(t, "name", last.Rules[0].Field)
	assert.Equal(t, ".*", last.Rules[0].Regex)
	assert.Equal(t, "default.yaml", last.SecureData)
}

func TestAppendDefaultCredentialRule_PreservesOrder(t *testing.T) {
	r1 := &config_domain.RegexMatchesList{
		Rules:      []config_domain.MatchingRule{{Field: "name", Regex: "alpha"}},
		SecureData: "a.yaml",
	}
	r2 := &config_domain.RegexMatchesList{
		Rules:      []config_domain.MatchingRule{{Field: "name", Regex: "beta"}},
		SecureData: "b.yaml",
	}
	result := appendDefaultCredentialRule([]*config_domain.RegexMatchesList{r1, r2})
	require.Len(t, result, 3)
	assert.Equal(t, r1, result[0])
	assert.Equal(t, r2, result[1])
	assert.Equal(t, ".*", result[2].Rules[0].Regex)
}

// ── summariseFilters ──────────────────────────────────────────────────────────

func TestSummariseFilters_Empty(t *testing.T) {
	assert.Equal(t, "none", summariseFilters(nil))
	assert.Equal(t, "none", summariseFilters([]config_domain.MatchingRule{}))
}

func TestSummariseFilters_SingleExclusive(t *testing.T) {
	filters := []config_domain.MatchingRule{
		{Field: "name", Regex: `DEV-.*`, Inclusive: false},
	}
	result := summariseFilters(filters)
	assert.Contains(t, result, "name")
	assert.Contains(t, result, `DEV-.*`)
	assert.Contains(t, result, "excl")
}

func TestSummariseFilters_SingleInclusive(t *testing.T) {
	filters := []config_domain.MatchingRule{
		{Field: "type", Regex: "elasticsearch", Inclusive: true},
	}
	result := summariseFilters(filters)
	assert.Contains(t, result, "type")
	assert.Contains(t, result, "elasticsearch")
	assert.Contains(t, result, "incl")
}

func TestSummariseFilters_Multiple(t *testing.T) {
	filters := []config_domain.MatchingRule{
		{Field: "name", Regex: `DEV-.*`, Inclusive: false},
		{Field: "type", Regex: "elasticsearch", Inclusive: true},
	}
	result := summariseFilters(filters)
	// Both entries must appear in the summary.
	assert.Contains(t, result, "name")
	assert.Contains(t, result, "type")
	assert.Contains(t, result, "excl")
	assert.Contains(t, result, "incl")
}

// ── writeSecureFileData ───────────────────────────────────────────────────────

func TestWriteSecureFileData_WritesYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "creds.yaml")

	type testCred struct {
		User     string `yaml:"user"`
		Password string `yaml:"password"`
	}
	cred := testCred{User: "admin", Password: "secret"}

	err := writeSecureFileData(cred, path)
	require.NoError(t, err)

	raw, readErr := os.ReadFile(path)
	require.NoError(t, readErr)
	content := string(raw)
	assert.Contains(t, content, "admin")
	assert.Contains(t, content, "secret")
}

func TestWriteSecureFileData_FilePermissions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "secure.yaml")

	err := writeSecureFileData(map[string]string{"key": "value"}, path)
	require.NoError(t, err)

	info, statErr := os.Stat(path)
	require.NoError(t, statErr)
	// File must be owner-read/write only (0600).
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
}

func TestWriteSecureFileData_GrafanaConnection(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "default.yaml")

	conn := config_domain.GrafanaConnection{
		"user":              "testuser",
		"basicAuthPassword": "testpass",
	}

	err := writeSecureFileData(conn, path)
	require.NoError(t, err)

	raw, _ := os.ReadFile(path)
	content := string(raw)
	assert.Contains(t, content, "testuser")
	assert.Contains(t, content, "testpass")
}
