package model

import (
	"time"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	OrderPending     OrderStatus = "pending"
	OrderPaid        OrderStatus = "paid"
	OrderFailed      OrderStatus = "failed"
	OrderCompensated OrderStatus = "compensated"
)

type Order struct {
	ID             uuid.UUID   `gorm:"type:uuid;primaryKey" json:"id"`
	UserID         uuid.UUID   `gorm:"type:uuid;index;not null" json:"user_id"`
	EventID        uuid.UUID   `gorm:"type:uuid;index;not null" json:"event_id"`
	HoldID         uuid.UUID   `gorm:"type:uuid;index;not null" json:"hold_id"`
	SeatIDs        []uuid.UUID `gorm:"serializer:json" json:"seat_ids"`
	Amount         int64       `gorm:"not null" json:"amount"`
	Currency       string      `gorm:"type:varchar(3);not null;default:USD" json:"currency"`
	Status         OrderStatus `gorm:"type:varchar(16);not null;index" json:"status"`
	IdempotencyKey string      `gorm:"type:varchar(128);uniqueIndex" json:"idempotency_key,omitempty"`
	CreatedAt      time.Time   `json:"created_at"`
}

func (Order) TableName() string { return "orders" }

type PaymentStatus string

const (
	PaymentSucceeded PaymentStatus = "succeeded"
	PaymentFailed    PaymentStatus = "failed"
	PaymentRefunded  PaymentStatus = "refunded"
)

type Payment struct {
	ID          uuid.UUID     `gorm:"type:uuid;primaryKey" json:"id"`
	OrderID     uuid.UUID     `gorm:"type:uuid;index;not null" json:"order_id"`
	Provider    string        `gorm:"type:varchar(32);not null" json:"provider"`
	ProviderRef string        `gorm:"type:varchar(128)" json:"provider_ref"`
	Status      PaymentStatus `gorm:"type:varchar(16);not null" json:"status"`
	Amount      int64         `gorm:"not null" json:"amount"`
	CreatedAt   time.Time     `json:"created_at"`
}

func (Payment) TableName() string { return "payments" }

type SagaLog struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	OrderID   uuid.UUID `gorm:"type:uuid;index;not null" json:"order_id"`
	Step      string    `gorm:"type:varchar(32);not null" json:"step"`
	State     string    `gorm:"type:varchar(32);not null" json:"state"`
	CreatedAt time.Time `json:"created_at"`
}

func (SagaLog) TableName() string { return "saga_logs" }
