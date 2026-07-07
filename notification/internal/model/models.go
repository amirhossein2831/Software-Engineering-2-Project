package model

import (
	"time"

	"github.com/google/uuid"
)

type Channel string

const (
	ChannelEmail Channel = "email"
	ChannelSMS   Channel = "sms"
)

type Status string

const (
	StatusPending Status = "pending"
	StatusSent    Status = "sent"
	StatusFailed  Status = "failed"
)

type NotificationOutbox struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	EventKey  string    `gorm:"type:varchar(160);uniqueIndex;not null" json:"event_key"`
	Channel   Channel   `gorm:"type:varchar(16);not null" json:"channel"`
	To        string    `gorm:"type:varchar(320);not null" json:"to"`
	Template  string    `gorm:"type:varchar(48);not null" json:"template"`
	Payload   string    `gorm:"type:text" json:"payload"`
	Status    Status    `gorm:"type:varchar(16);not null;index" json:"status"`
	Attempts  int       `gorm:"not null;default:0" json:"attempts"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (NotificationOutbox) TableName() string { return "notification_outbox" }

type Recipient struct {
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"user_id"`
	Email     string    `gorm:"type:varchar(320)" json:"email"`
	Phone     string    `gorm:"type:varchar(32)" json:"phone"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Recipient) TableName() string { return "recipients" }
