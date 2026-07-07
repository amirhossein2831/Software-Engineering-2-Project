package qr

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

	"github.com/google/uuid"
)

type Signer struct {
	key []byte
}

func NewSigner(secret string) *Signer {
	return &Signer{key: []byte(secret)}
}

func (s *Signer) Hash(parts ...uuid.UUID) string {
	h := hmac.New(sha256.New, s.key)
	for _, p := range parts {
		h.Write(p[:])
	}
	return hex.EncodeToString(h.Sum(nil))
}
