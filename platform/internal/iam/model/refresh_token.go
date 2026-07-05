package model

import (
	"time"

	"github.com/google/uuid"
)

// RefreshToken is a persisted, hashed refresh token. The raw token is never
// stored; only its SHA-256 hash. A token is valid while RevokedAt is nil and
// ExpiresAt is in the future.
type RefreshToken struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;index;not null"`
	TokenHash string    `gorm:"uniqueIndex;not null"`
	ExpiresAt time.Time `gorm:"not null"`
	RevokedAt *time.Time
	CreatedAt time.Time
}

// TableName pins the table name.
func (RefreshToken) TableName() string { return "refresh_tokens" }

// Active reports whether the token can still be used at time now.
func (t RefreshToken) Active(now time.Time) bool {
	return t.RevokedAt == nil && now.Before(t.ExpiresAt)
}
