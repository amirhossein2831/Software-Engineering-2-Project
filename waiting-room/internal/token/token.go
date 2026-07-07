package token

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

type Signer struct {
	key []byte
}

func NewSigner(secret string) *Signer {
	return &Signer{key: []byte(secret)}
}

func (s *Signer) Issue(eventID, userID string, ttl time.Duration) string {
	claims := Claims{EventID: eventID, UserID: userID, Exp: time.Now().Add(ttl).Unix()}
	payload, _ := json.Marshal(claims)
	body := base64.RawURLEncoding.EncodeToString(payload)
	return body + "." + s.sign(body)
}

func (s *Signer) Verify(tok string) (*Claims, error) {
	body, sig, ok := strings.Cut(tok, ".")
	if !ok {
		return nil, ErrMalformed
	}
	if !hmac.Equal([]byte(sig), []byte(s.sign(body))) {
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

func (s *Signer) sign(body string) string {
	mac := hmac.New(sha256.New, s.key)
	mac.Write([]byte(body))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
