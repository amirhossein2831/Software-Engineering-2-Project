package model

import (
	"time"

	"github.com/google/uuid"
)

type SeatStatus string

const (
	SeatAvailable SeatStatus = "available"
	SeatLocked    SeatStatus = "locked"
	SeatBooked    SeatStatus = "booked"
)

type SeatState struct {
	EventID   uuid.UUID  `gorm:"type:uuid;primaryKey" json:"event_id"`
	SeatID    uuid.UUID  `gorm:"type:uuid;primaryKey" json:"seat_id"`
	Status    SeatStatus `gorm:"type:varchar(16);not null;default:available" json:"status"`
	HoldID    *uuid.UUID `gorm:"type:uuid;index" json:"hold_id,omitempty"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func (SeatState) TableName() string { return "seat_states" }

type ReservationStatus string

const (
	HoldLocked    ReservationStatus = "locked"
	HoldCommitted ReservationStatus = "committed"
	HoldReleased  ReservationStatus = "released"
)

type Reservation struct {
	ID        uuid.UUID         `gorm:"type:uuid;primaryKey" json:"id"`
	EventID   uuid.UUID         `gorm:"type:uuid;index;not null" json:"event_id"`
	UserID    uuid.UUID         `gorm:"type:uuid;index;not null" json:"user_id"`
	Status    ReservationStatus `gorm:"type:varchar(16);not null;index" json:"status"`
	ExpiresAt time.Time         `gorm:"index" json:"expires_at"`
	CreatedAt time.Time         `json:"created_at"`
}

func (Reservation) TableName() string { return "reservations" }
