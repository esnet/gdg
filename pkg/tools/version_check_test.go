package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type moo struct {
	s string
}

func (v moo) GetServerInfo() map[string]any {
	return map[string]any{"Version": v.s}
}

func TestGrafanaRange(t *testing.T) {
	m := moo{s: "10.5.4"}
	assert.False(t, InRange([]VersionRange{{MinVersion: "v10.2.1", MaxVersion: "v10.2.2"}}, m))
	assert.True(t, InRange([]VersionRange{{MinVersion: "v10.5.1", MaxVersion: "v10.6.0"}}, m))
	// Inclusive tests
	assert.True(t, InRange([]VersionRange{{MinVersion: "v10.5.4", MaxVersion: "v10.6.0"}}, m))
	assert.True(t, InRange([]VersionRange{{MinVersion: "v10.5.1", MaxVersion: "v10.5.4"}}, m))
	assert.False(t, InRange([]VersionRange{{MinVersion: "v10.2.1", MaxVersion: "v10.2.2"}}, m))
	m.s = "10.2.0"
	assert.False(t, InRange([]VersionRange{{MinVersion: "v10.2.1", MaxVersion: "v10.2.2"}}, m))
	m.s = "10.2.1"
	assert.True(t, InRange([]VersionRange{
		{MinVersion: "v10.2.1", MaxVersion: "v10.2.2"},
		{MinVersion: "v10.1.0", MaxVersion: "v10.5.2"},
	}, m))
}

func TestGrafanaMinVersion(t *testing.T) {
	m := moo{s: "10.5.4"}
	assert.True(t, ValidateMinimumVersion("v10.3.2", m))
	assert.False(t, ValidateMinimumVersion("v10.7.2", m))
}
