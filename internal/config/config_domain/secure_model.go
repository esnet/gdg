package config_domain

import "log/slog"

// SecureModel holds secure information like password and token
type SecureModel struct {
	Password string `mapstructure:"password" json:"password" yaml:"password"` // #nosec G117 this deals with encrypted values
	Token    string `mapstructure:"token"  json:"token" yaml:"token"`
}

func (sm *SecureModel) Empty() bool {
	return sm == nil || (sm.Password == "" && sm.Token == "")
}

// UpdateSecureModel updates Token and Password by applying fn; logs errors on failure.
func (sm *SecureModel) UpdateSecureModel(fn func(string) (string, error)) {
	if sm.Token != "" {
		newToken, err := fn(sm.Token)
		if err == nil {
			sm.Token = newToken
		} else {
			slog.Warn("error updating secure model, cannot decode token", "err", err)
		}
	}
	if sm.Password != "" {
		newPassword, err := fn(sm.Password)
		if err == nil {
			sm.Password = newPassword
		} else {
			slog.Warn("error updating secure model, cannot decode password", "err", err)
		}
	}
}
