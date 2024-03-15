package config

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"github.com/sethvargo/go-password/password"
	"log/slog"
	"math/big"
)

const (
	minPasswordLength = 8
	maxPasswordLength = 20
)

type UserSettings struct {
	RandomPassword bool `mapstructure:"random_password" yaml:"random_password"`
	MinLength      int  `mapstructure:"min_length" yaml:"min_length"`
	MaxLength      int  `mapstructure:"max_length" yaml:"max_length"`
}

func (u *UserSettings) GetPassword(username string) string {
	if !u.RandomPassword {
		return u.defaultUserPassword(username)
	} else if u.MinLength > u.MaxLength {
		slog.Warn("min length is greater than max length, falling back on default behavior")
		return u.defaultUserPassword(username)
	}

	nBig, err := rand.Int(rand.Reader, big.NewInt(int64(u.MaxLength)))
	if err != nil {
		slog.Warn("Failed to get random value")
		return u.defaultUserPassword(username)
	}
	passLength := int(nBig.Int64() + int64(u.MinLength))
	res, err := password.Generate(passLength, 1, 1, false, false)
	if err != nil {
		slog.Warn("unable to generate a proper random password, falling back on default password pattern",
			slog.String("username", username))
		return u.defaultUserPassword(username)
	}
	return res
}

func (u *UserSettings) defaultUserPassword(username string) string {
	if username == "admin" {
		return ""
	}

	username = username + ".json"
	//generate user password
	h := sha256.New()
	passwordVal := func() string {
		h.Write([]byte(username))
		hash := h.Sum(nil)
		passwordVal := fmt.Sprintf("%x", hash)
		return passwordVal
	}()

	return passwordVal
}
