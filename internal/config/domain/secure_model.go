package domain

type SecureModel struct {
	Password string `json:"password"`
	Token    string `json:"token"`
}
