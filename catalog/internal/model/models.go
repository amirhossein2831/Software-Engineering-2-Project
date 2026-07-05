package model

import (
	"time"

	"github.com/google/uuid"
)

type EventStatus string

const (
	StatusDraft     EventStatus = "draft"
	StatusPublished EventStatus = "published"
)

type Venue struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name      string    `gorm:"not null" json:"name"`
	Address   string    `json:"address"`
	CreatedBy uuid.UUID `gorm:"type:uuid;not null" json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	Sectors   []Sector  `gorm:"foreignKey:VenueID" json:"sectors,omitempty"`
}

func (Venue) TableName() string { return "venues" }

type Sector struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	VenueID  uuid.UUID `gorm:"type:uuid;index;not null" json:"venue_id"`
	Name     string    `gorm:"not null" json:"name"`
	RowCount int       `gorm:"not null" json:"row_count"`
	ColCount int       `gorm:"not null" json:"col_count"`
}

func (Sector) TableName() string { return "sectors" }

type Event struct {
	ID          uuid.UUID   `gorm:"type:uuid;primaryKey" json:"id"`
	Title       string      `gorm:"not null;index" json:"title"`
	Description string      `json:"description"`
	Genre       string      `gorm:"index" json:"genre"`
	Location    string      `gorm:"index" json:"location"`
	VenueID     uuid.UUID   `gorm:"type:uuid;index;not null" json:"venue_id"`
	OrganizerID uuid.UUID   `gorm:"type:uuid;index;not null" json:"organizer_id"`
	StartsAt    time.Time   `gorm:"index" json:"starts_at"`
	Status      EventStatus `gorm:"type:varchar(16);not null;default:draft;index" json:"status"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

func (Event) TableName() string { return "events" }

type Pricing struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	EventID  uuid.UUID `gorm:"type:uuid;index;not null" json:"event_id"`
	SectorID uuid.UUID `gorm:"type:uuid;index;not null" json:"sector_id"`
	Amount   int64     `gorm:"not null" json:"amount"`
	Currency string    `gorm:"type:varchar(3);not null;default:USD" json:"currency"`
}

func (Pricing) TableName() string { return "pricings" }

type Seat struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	EventID  uuid.UUID `gorm:"type:uuid;index;not null" json:"event_id"`
	SectorID uuid.UUID `gorm:"type:uuid;index;not null" json:"sector_id"`
	RowLabel string    `gorm:"not null" json:"row_label"`
	Number   int       `gorm:"not null" json:"number"`
}

func (Seat) TableName() string { return "seats" }
