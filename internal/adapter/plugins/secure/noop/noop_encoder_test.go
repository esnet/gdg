package noop

import (
	"testing"

	"github.com/esnet/gdg/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNoOpEncoder_EncodeValue(t *testing.T) {
	enc := NoOpEncoder{}

	tests := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"plain string", "hello"},
		{"string with special chars", "s3cr3t!@#$%"},
		{"unicode string", "café"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := enc.EncodeValue(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.input, got, "EncodeValue must return input unchanged")
		})
	}
}

func TestNoOpEncoder_DecodeValue(t *testing.T) {
	enc := NoOpEncoder{}

	tests := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"plain string", "world"},
		{"encoded-looking string", "hello%20world"},
		{"unicode string", "日本語"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := enc.DecodeValue(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.input, got, "DecodeValue must return input unchanged")
		})
	}
}

func TestNoOpEncoder_Encode(t *testing.T) {
	enc := NoOpEncoder{}

	tests := []struct {
		name         string
		resourceType domain.ResourceType
		input        []byte
	}{
		{"nil bytes", domain.DashboardResource, nil},
		{"empty bytes", domain.FolderResource, []byte{}},
		{"json payload", domain.ConnectionResource, []byte(`{"uid":"abc","title":"My Dashboard"}`)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := enc.Encode(tt.resourceType, tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.input, got, "Encode must return input bytes unchanged")
		})
	}
}

func TestNoOpEncoder_Decode(t *testing.T) {
	enc := NoOpEncoder{}

	tests := []struct {
		name         string
		resourceType domain.ResourceType
		input        []byte
	}{
		{"nil bytes", domain.DashboardResource, nil},
		{"empty bytes", domain.AlertingResource, []byte{}},
		{"arbitrary bytes", domain.TeamResource, []byte("some-encoded-data")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := enc.Decode(tt.resourceType, tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.input, got, "Decode must return input bytes unchanged")
		})
	}
}

// TestNoOpEncoder_RoundTrip verifies that Encode followed by Decode is a no-op.
func TestNoOpEncoder_RoundTrip(t *testing.T) {
	enc := NoOpEncoder{}
	payload := []byte(`{"uid":"round-trip","title":"Test"}`)

	encoded, err := enc.Encode(domain.DashboardResource, payload)
	require.NoError(t, err)

	decoded, err := enc.Decode(domain.DashboardResource, encoded)
	require.NoError(t, err)

	assert.Equal(t, payload, decoded)
}
