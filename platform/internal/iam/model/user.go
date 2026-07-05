// Package model holds the IAM domain entities. Per the platform convention the
// GORM structs ARE the domain models (direct domain=DB mapping); there is no
// separate persistence layer.
package model

import (
	"time"

	"github.com/google/uuid"
)

// Role is a user's access role.
type Role string

const (
	RoleBuyer     Role = "buyer"
	RoleOrganizer Role = "organizer"
	RoleAdmin     Role = "admin"
)

// Valid reports whether r is a recognised role.
func (r Role) Valid() bool {
	switch r {
	case RoleBuyer, RoleOrganizer, RoleAdmin:
		return true
	default:
		return false
	}
}

// User is a platform participant.
type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	Email        string    `gorm:"uniqueIndex;not null"`
	PasswordHash string    `gorm:"not null"`
	Role         Role      `gorm:"type:varchar(16);not null;default:buyer"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// TableName pins the table name.
func (User) TableName() string { return "users" }
