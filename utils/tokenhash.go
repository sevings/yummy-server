package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"time"
)

type TokenHashConfig interface {
	ConfigString(key string) string
}

type TokenHash struct {
	cfg TokenHashConfig
}

func NewTokenHash(cfg TokenHashConfig) TokenHash {
	return TokenHash{cfg: cfg}
}

func (th TokenHash) tokenHash(token, config string) []byte {
	salt := th.cfg.ConfigString(config)
	sum := sha256.Sum256([]byte(token + salt))
	return sum[:]
}

func (th TokenHash) AppSecretHash(secret string) []byte {
	return th.tokenHash(secret, "server.app_salt")
}

func (th TokenHash) AppTokenHash(token string) []byte {
	return th.tokenHash(token, "server.app_salt")
}

func (th TokenHash) AccessTokenHash(token string) []byte {
	return th.tokenHash(token, "server.at_salt")
}

func (th TokenHash) RefreshTokenHash(token string) []byte {
	return th.tokenHash(token, "server.rt_salt")
}

func (th TokenHash) PasswordHash(password string) []byte {
	return th.tokenHash(password, "server.pass_salt")
}

func (th TokenHash) VerificationCode(email string) string {
	hash := th.tokenHash(email, "server.mail_salt")
	return hex.EncodeToString(hash)
}

func (th TokenHash) resetCode(email string, date int64) string {
	str := email + strconv.FormatInt(date, 16)
	hash := th.tokenHash(str, "server.mail_salt")
	return hex.EncodeToString(hash)
}

func (th TokenHash) ResetPasswordCode(email string) (string, int64) {
	date := time.Now().Unix()
	code := th.resetCode(email, date)
	return code, date
}

func (th TokenHash) CheckResetPasswordCode(email, code string, date int64) bool {
	now := time.Now().Unix()
	if (now - date) >= 60*60 {
		return false
	}

	if date > now {
		return false
	}

	return th.resetCode(email, date) == code
}
