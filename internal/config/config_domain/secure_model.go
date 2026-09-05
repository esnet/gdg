package config_domain

import (
	"log/slog"

	"github.com/esnet/gdg/internal/ports/outbound"
)

// SecureModel holds secure information like password and token
type SecureModel struct {
	Password string `mapstructure:"password" json:"password" yaml:"password"` // #nosec G117 this deals with encrypted values
	Token    string `mapstructure:"token"  json:"token" yaml:"token"`
}

func (sm *SecureModel) Empty() bool {
	return sm == nil || (sm.Password == "" && sm.Token == "")
}

// UpdateSecureModel updates Token and Password by applying fn (typically a
// CipherEncoder's DecodeValue); logs errors on failure.
//
// Fields holding a lookup reference (outbound.IsLookupRef — e.g.
// "lookup:gsm:projects/.../versions/1") are left untouched here: a lookup
// reference is not cipher-encoded ciphertext, so running it through a
// cipher decode function would corrupt it rather than resolve it. Those
// fields are instead resolved via ResolveLookups.
func (sm *SecureModel) UpdateSecureModel(fn func(string) (string, error)) {
	if sm.Token != "" && !outbound.IsLookupRef(sm.Token) {
		newToken, err := fn(sm.Token)
		if err == nil {
			sm.Token = newToken
		} else {
			slog.Warn("error updating secure model, cannot decode token", "err", err)
		}
	}
	if sm.Password != "" && !outbound.IsLookupRef(sm.Password) {
		newPassword, err := fn(sm.Password)
		if err == nil {
			sm.Password = newPassword
		} else {
			slog.Warn("error updating secure model, cannot decode password", "err", err)
		}
	}
}

// ResolveLookups resolves Token and Password through fn (typically a
// LookupResolver's Resolve) when — and only when — the value is a lookup
// reference (outbound.IsLookupRef). This is the mirror image of
// UpdateSecureModel: that method acts on everything except lookup
// references, this one acts on nothing but them, so a field is always
// handled by exactly one of the two, never both and never neither.
func (sm *SecureModel) ResolveLookups(fn func(string) (string, error)) {
	if sm.Token != "" && outbound.IsLookupRef(sm.Token) {
		newToken, err := fn(sm.Token)
		if err == nil {
			sm.Token = newToken
		} else {
			slog.Warn("error resolving lookup token", "err", err)
		}
	}
	if sm.Password != "" && outbound.IsLookupRef(sm.Password) {
		newPassword, err := fn(sm.Password)
		if err == nil {
			sm.Password = newPassword
		} else {
			slog.Warn("error resolving lookup password", "err", err)
		}
	}
}
