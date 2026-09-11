package config_domain

import (
	"strings"

	"github.com/esnet/gdg/internal/domain"
)

// stubLookupSvc is a minimal outbound.LookupService for tests in this package.
type stubLookupSvc struct{}

func (s stubLookupSvc) IsLookupRef(raw string) bool { return strings.HasPrefix(raw, "lookup:") }
func (s stubLookupSvc) Prefix() string              { return "lookup:" }

// stubCipherEncoder is a minimal outbound.CipherEncoder for tests in this package.
// DecodeValue always returns decodeResult; EncodeValue returns the input unchanged.
type stubCipherEncoder struct {
	decodeResult string
}

func (s *stubCipherEncoder) EncodeValue(b string) (string, error) { return b, nil }
func (s *stubCipherEncoder) DecodeValue(b string) (string, error) { return s.decodeResult, nil }
func (s *stubCipherEncoder) Encode(_ domain.ResourceType, b []byte) ([]byte, error) {
	return b, nil
}

func (s *stubCipherEncoder) Decode(_ domain.ResourceType, b []byte) ([]byte, error) {
	return b, nil
}

// stubLookupResolver is a minimal outbound.LookupResolver for tests in this package.
// Resolve always returns value.
type stubLookupResolver struct {
	value string
}

func (s *stubLookupResolver) Resolve(_ string) (string, error) { return s.value, nil }
