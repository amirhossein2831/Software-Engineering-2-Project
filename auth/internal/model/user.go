package model

import (
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleBuyer     Role = "buyer"
	RoleOrganizer Role = "organizer"
	RoleAdmin     Role = "admin"
)

func (r Role) Valid() bool {
	switch r {
	case RoleBuyer, RoleOrganizer, RoleAdmin:
		return true
	default:
		return false
	}
}

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	Email        string    `gorm:"uniqueIndex;not null"`
	PasswordHash string    `gorm:"not null"`
	Role         Role      `gorm:"type:varchar(16);not null;default:buyer"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (User) TableName() string { return "users" }
