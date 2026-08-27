package config_tooling

import (
	"testing"

	"github.com/esnet/gdg/internal/config/config_domain"
	"github.com/stretchr/testify/assert"
)

// testConn is a JSON-serialisable entity used across filter tests.
type testConn struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// testContactPoint mirrors the shape of a Grafana contact point.
type testContactPoint struct {
	Name  string `json:"name"`
	OrgID int64  `json:"orgId"`
}

// testReceiver and testContactPointFull mirror the real contacts.json shape,
// with a nested receivers array, to exercise gjson array paths (e.g. receivers.#.type).
type testReceiver struct {
	Type string `json:"type"`
	UID  string `json:"uid"`
}

type testContactPointFull struct {
	Name      string         `json:"name"`
	OrgID     int64          `json:"orgId"`
	Receivers []testReceiver `json:"receivers"`
}

// ── nil / empty rules ─────────────────────────────────────────────────────────

func TestIsExcluded_NilRulesNeverExcludes(t *testing.T) {
	assert.False(t, IsExcluded(testConn{Name: "prod-db", Type: "postgres"}, nil))
}

func TestIsExcluded_EmptyRulesNeverExcludes(t *testing.T) {
	assert.False(t, IsExcluded(testConn{Name: "prod-db"}, []config_domain.MatchingRule{}))
}

// ── exclusive rules (Inclusive: false) ───────────────────────────────────────

func TestIsExcluded_ExclusiveMatchExcludes(t *testing.T) {
	rules := []config_domain.MatchingRule{
		{Field: "name", Regex: "dev-.*", Inclusive: false},
	}
	// "dev-db" matches the exclusive rule → excluded
	assert.True(t, IsExcluded(testConn{Name: "dev-db", Type: "mysql"}, rules))
}

func TestIsExcluded_ExclusiveNoMatchDoesNotExclude(t *testing.T) {
	rules := []config_domain.MatchingRule{
		{Field: "name", Regex: "dev-.*", Inclusive: false},
	}
	// "prod-db" does not match → not excluded
	assert.False(t, IsExcluded(testConn{Name: "prod-db", Type: "mysql"}, rules))
}

// ── inclusive rules (Inclusive: true) ────────────────────────────────────────

func TestIsExcluded_InclusiveMatchDoesNotExclude(t *testing.T) {
	rules := []config_domain.MatchingRule{
		{Field: "name", Regex: "prod-.*", Inclusive: true},
	}
	// "prod-db" matches → flipped → NOT excluded
	assert.False(t, IsExcluded(testConn{Name: "prod-db", Type: "mysql"}, rules))
}

func TestIsExcluded_InclusiveNoMatchExcludes(t *testing.T) {
	rules := []config_domain.MatchingRule{
		{Field: "name", Regex: "prod-.*", Inclusive: true},
	}
	// "dev-db" does not match → flipped → IS excluded
	assert.True(t, IsExcluded(testConn{Name: "dev-db", Type: "mysql"}, rules))
}

// ── edge cases ────────────────────────────────────────────────────────────────

func TestIsExcluded_InvalidRegexExcludes(t *testing.T) {
	rules := []config_domain.MatchingRule{
		{Field: "name", Regex: "[bad", Inclusive: false},
	}
	// IsValid returns (true, nil) for invalid regex → excluded
	assert.True(t, IsExcluded(testConn{Name: "anything", Type: "mysql"}, rules))
}

func TestIsExcluded_MissingFieldIsSkipped(t *testing.T) {
	rules := []config_domain.MatchingRule{
		{Field: "nonexistent_field", Regex: ".*", Inclusive: false},
	}
	// Field absent in JSON → MissingFieldErr → rule skipped → not excluded
	assert.False(t, IsExcluded(testConn{Name: "anything", Type: "mysql"}, rules))
}

func TestIsExcluded_EmptyRegexSkipsRule(t *testing.T) {
	rules := []config_domain.MatchingRule{
		{Field: "name", Regex: "", Inclusive: false},
	}
	// Empty regex → MissingFieldErr → rule skipped → not excluded
	assert.False(t, IsExcluded(testConn{Name: "anything"}, rules))
}

func TestIsExcluded_UnmarshalFailureExcludes(t *testing.T) {
	rules := []config_domain.MatchingRule{
		{Field: "name", Regex: ".*"},
	}
	// Channels cannot be JSON-marshalled → conservative true
	assert.True(t, IsExcluded(make(chan int), rules))
}

// ── contact-point shape (numeric orgId field) ─────────────────────────────────

func TestIsExcluded_ContactPoint_NameExclusiveMatch(t *testing.T) {
	rules := []config_domain.MatchingRule{
		{Field: "name", Regex: "discord"},
	}
	assert.True(t, IsExcluded(testContactPoint{Name: "discord-alerts", OrgID: 1}, rules))
	assert.False(t, IsExcluded(testContactPoint{Name: "pagerduty", OrgID: 1}, rules))
}

func TestIsExcluded_ContactPoint_OrgIDInclusiveFilter(t *testing.T) {
	rules := []config_domain.MatchingRule{
		{Field: "orgId", Regex: "^1$", Inclusive: true},
	}
	// orgId=1 matches inclusive → flipped → NOT excluded (kept)
	assert.False(t, IsExcluded(testContactPoint{Name: "discord-alerts", OrgID: 1}, rules))
	// orgId=2 does not match → flipped → IS excluded
	assert.True(t, IsExcluded(testContactPoint{Name: "other", OrgID: 2}, rules))
}

// ── multiple rules ────────────────────────────────────────────────────────────

func TestIsExcluded_FirstMatchingRuleWins(t *testing.T) {
	rules := []config_domain.MatchingRule{
		{Field: "name", Regex: "discord"}, // exclusive: matches → excluded; loop stops here
		{Field: "name", Regex: "alerts"},  // also matches, but first rule already returned
	}
	assert.True(t, IsExcluded(testContactPoint{Name: "discord-alerts", OrgID: 1}, rules))
}

func TestIsExcluded_NoRuleMatches_NotExcluded(t *testing.T) {
	rules := []config_domain.MatchingRule{
		{Field: "name", Regex: "slack"},
		{Field: "name", Regex: "teams"},
	}
	assert.False(t, IsExcluded(testContactPoint{Name: "pagerduty", OrgID: 1}, rules))
}

// ── array field paths (gjson receivers.#.type style) ─────────────────────────

func TestIsExcluded_Array_ExclusiveMatch_Excludes(t *testing.T) {
	cp := testContactPointFull{
		Name:      "discord",
		OrgID:     1,
		Receivers: []testReceiver{{Type: "discord", UID: "abc"}},
	}
	rules := []config_domain.MatchingRule{
		{Field: "receivers.#.type", Regex: "discord", Inclusive: false},
	}
	// "discord" element found → matched → excluded
	assert.True(t, IsExcluded(cp, rules))
}

func TestIsExcluded_Array_ExclusiveNoMatch_NotExcluded(t *testing.T) {
	cp := testContactPointFull{
		Name:      "discord",
		OrgID:     1,
		Receivers: []testReceiver{{Type: "discord", UID: "abc"}},
	}
	rules := []config_domain.MatchingRule{
		{Field: "receivers.#.type", Regex: "slack", Inclusive: false},
	}
	// No element matches "slack" → not excluded
	assert.False(t, IsExcluded(cp, rules))
}

func TestIsExcluded_Array_InclusiveMatch_NotExcluded(t *testing.T) {
	cp := testContactPointFull{
		Name:      "discord",
		OrgID:     1,
		Receivers: []testReceiver{{Type: "discord", UID: "abc"}},
	}
	rules := []config_domain.MatchingRule{
		{Field: "receivers.#.type", Regex: "discord", Inclusive: true},
	}
	// "discord" matches the inclusive rule → flipped → NOT excluded (kept)
	assert.False(t, IsExcluded(cp, rules))
}

func TestIsExcluded_Array_InclusiveNoMatch_Excludes(t *testing.T) {
	cp := testContactPointFull{
		Name:      "discord",
		OrgID:     1,
		Receivers: []testReceiver{{Type: "discord", UID: "abc"}},
	}
	rules := []config_domain.MatchingRule{
		{Field: "receivers.#.type", Regex: "slack", Inclusive: true},
	}
	// No element matches "slack" → inclusive no-match → IS excluded
	assert.True(t, IsExcluded(cp, rules))
}

func TestIsExcluded_Array_MultipleElements_FirstMatchWins(t *testing.T) {
	cp := testContactPointFull{
		Name:  "multi",
		OrgID: 1,
		Receivers: []testReceiver{
			{Type: "email", UID: "x"},
			{Type: "slack", UID: "y"},
		},
	}
	rules := []config_domain.MatchingRule{
		{Field: "receivers.#.type", Regex: "slack", Inclusive: false},
	}
	// "slack" found in array → excluded
	assert.True(t, IsExcluded(cp, rules))
}

func TestIsExcluded_Array_Empty_NotExcluded(t *testing.T) {
	cp := testContactPointFull{Name: "empty", OrgID: 1, Receivers: []testReceiver{}}
	rules := []config_domain.MatchingRule{
		{Field: "receivers.#.type", Regex: ".*", Inclusive: false},
	}
	// Empty array → no elements to match → not excluded
	assert.False(t, IsExcluded(cp, rules))
}

func TestIsExcluded_Array_Empty_Inclusive_Excludes(t *testing.T) {
	cp := testContactPointFull{Name: "empty", OrgID: 1, Receivers: []testReceiver{}}
	rules := []config_domain.MatchingRule{
		{Field: "receivers.#.type", Regex: ".*", Inclusive: true},
	}
	// Empty array → no match → inclusive no-match → IS excluded
	assert.True(t, IsExcluded(cp, rules))
}
