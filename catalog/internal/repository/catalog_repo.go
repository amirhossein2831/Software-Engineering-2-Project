package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"catalog/internal/model"
)

var ErrNotFound = errors.New("not found")

type EventFilter struct {
	Genre    string
	Location string
	OnlyPublished bool
}

type CatalogRepo struct {
	db *gorm.DB
}

func NewCatalogRepo(db *gorm.DB) *CatalogRepo {
	return &CatalogRepo{db: db}
}

func (r *CatalogRepo) CreateVenue(ctx context.Context, v *model.Venue) error {
	return r.db.WithContext(ctx).Create(v).Error
}

func (r *CatalogRepo) GetVenue(ctx context.Context, id uuid.UUID) (*model.Venue, error) {
	var v model.Venue
	err := r.db.WithContext(ctx).Preload("Sectors").First(&v, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *CatalogRepo) CreateSector(ctx context.Context, s *model.Sector) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *CatalogRepo) CreateEvent(ctx context.Context, e *model.Event) error {
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *CatalogRepo) GetEvent(ctx context.Context, id uuid.UUID) (*model.Event, error) {
	var e model.Event
	err := r.db.WithContext(ctx).First(&e, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *CatalogRepo) ListEvents(ctx context.Context, f EventFilter) ([]model.Event, error) {
	q := r.db.WithContext(ctx).Model(&model.Event{})
	if f.OnlyPublished {
		q = q.Where("status = ?", model.StatusPublished)
	}
	if f.Genre != "" {
		q = q.Where("genre = ?", f.Genre)
	}
	if f.Location != "" {
		q = q.Where("location = ?", f.Location)
	}
	var events []model.Event
	if err := q.Order("starts_at asc").Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

func (r *CatalogRepo) UpdateEventStatus(ctx context.Context, id uuid.UUID, status model.EventStatus) error {
	res := r.db.WithContext(ctx).Model(&model.Event{}).Where("id = ?", id).Update("status", status)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *CatalogRepo) CreatePricing(ctx context.Context, p *model.Pricing) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *CatalogRepo) ListPricing(ctx context.Context, eventID uuid.UUID) ([]model.Pricing, error) {
	var prices []model.Pricing
	if err := r.db.WithContext(ctx).Where("event_id = ?", eventID).Find(&prices).Error; err != nil {
		return nil, err
	}
	return prices, nil
}

func (r *CatalogRepo) CreateSeats(ctx context.Context, seats []model.Seat) error {
	if len(seats) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&seats).Error
}

func (r *CatalogRepo) ListSeats(ctx context.Context, eventID uuid.UUID) ([]model.Seat, error) {
	var seats []model.Seat
	if err := r.db.WithContext(ctx).Where("event_id = ?", eventID).Find(&seats).Error; err != nil {
		return nil, err
	}
	return seats, nil
}
