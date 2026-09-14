package config_domain

import "strings"

// stubLookupSvc is a minimal outbound.LookupService for tests in this package.
type stubLookupSvc struct{}

func (s stubLookupSvc) IsLookupRef(raw string) bool { return strings.HasPrefix(raw, "lookup:") }
func (s stubLookupSvc) Prefix() string              { return "lookup:" }
