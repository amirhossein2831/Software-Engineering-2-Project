package admission

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

var (
	ErrMalformed = errors.New("malformed admission token")
	ErrSignature = errors.New("invalid admission token signature")
	ErrExpired   = errors.New("admission token expired")
)

type Claims struct {
	EventID string `json:"event_id"`
	UserID  string `json:"user_id"`
	Exp     int64  `json:"exp"`
}

type Verifier struct {
	key []byte
}

func NewVerifier(secret string) *Verifier {
	return &Verifier{key: []byte(secret)}
}

func (v *Verifier) Verify(tok string) (*Claims, error) {
	body, sig, ok := strings.Cut(tok, ".")
	if !ok {
		return nil, ErrMalformed
	}
	if !hmac.Equal([]byte(sig), []byte(v.sign(body))) {
		return nil, ErrSignature
	}
	payload, err := base64.RawURLEncoding.DecodeString(body)
	if err != nil {
		return nil, ErrMalformed
	}
	var claims Claims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, ErrMalformed
	}
	if time.Now().Unix() > claims.Exp {
		return nil, ErrExpired
	}
	return &claims, nil
}

func (v *Verifier) sign(body string) string {
	mac := hmac.New(sha256.New, v.key)
	mac.Write([]byte(body))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
