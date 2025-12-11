package domain

// SecureModel holds secure information like password and token
type SecureModel struct {
	Password string `mapstructure:"password" json:"password" yaml:"password"`
	Token    string `mapstructure:"token"  json:"token" yaml:"token"`
}
