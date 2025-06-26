package domain

type SecureModel struct {
	Password string `mapstructure:"password" json:"password"`
	Token    string `mapstructure:"token"  json:"token"`
}
