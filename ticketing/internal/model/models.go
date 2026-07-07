package model

import (
	"time"

	"github.com/google/uuid"
)

type Ticket struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	OrderID  uuid.UUID `gorm:"type:uuid;index;not null" json:"order_id"`
	EventID  uuid.UUID `gorm:"type:uuid;index;not null" json:"event_id"`
	SeatID   uuid.UUID `gorm:"type:uuid;not null" json:"seat_id"`
	UserID   uuid.UUID `gorm:"type:uuid;index;not null" json:"user_id"`
	QRHash   string    `gorm:"type:varchar(64);uniqueIndex;not null" json:"qr_hash"`
	IssuedAt time.Time `json:"issued_at"`
}

func (Ticket) TableName() string { return "tickets" }
