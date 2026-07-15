package service

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/google/uuid"

	"catalog/internal/model"
	"catalog/internal/repository"
)

var (
	ErrVenueInUse    = errors.New("venue has events")
	ErrEventNotDraft = errors.New("event is not a draft")
)

type CatalogService struct {
	repo *repository.CatalogRepo
}

func NewCatalogService(repo *repository.CatalogRepo) *CatalogService {
	return &CatalogService{repo: repo}
}

func (s *CatalogService) CreateVenue(ctx context.Context, name, address string, createdBy uuid.UUID) (*model.Venue, error) {
	v := &model.Venue{ID: uuid.New(), Name: name, Address: address, CreatedBy: createdBy}
	if err := s.repo.CreateVenue(ctx, v); err != nil {
		return nil, err
	}
	return v, nil
}

func (s *CatalogService) GetVenue(ctx context.Context, id uuid.UUID) (*model.Venue, error) {
	return s.repo.GetVenue(ctx, id)
}

func (s *CatalogService) ListVenues(ctx context.Context) ([]model.Venue, error) {
	return s.repo.ListVenues(ctx)
}

func (s *CatalogService) UpdateVenue(ctx context.Context, id uuid.UUID, name, address string) (*model.Venue, error) {
	if err := s.repo.UpdateVenue(ctx, id, name, address); err != nil {
		return nil, err
	}
	return s.repo.GetVenue(ctx, id)
}

func (s *CatalogService) DeleteVenue(ctx context.Context, id uuid.UUID) error {
	n, err := s.repo.CountEventsByVenue(ctx, id)
	if err != nil {
		return err
	}
	if n > 0 {
		return ErrVenueInUse
	}
	return s.repo.DeleteVenue(ctx, id)
}

func (s *CatalogService) DeleteSector(ctx context.Context, venueID, sectorID uuid.UUID) error {
	return s.repo.DeleteSector(ctx, venueID, sectorID)
}

func (s *CatalogService) AddSector(ctx context.Context, venueID uuid.UUID, name string, rows, cols int) (*model.Sector, error) {
	sec := &model.Sector{ID: uuid.New(), VenueID: venueID, Name: name, RowCount: rows, ColCount: cols}
	if err := s.repo.CreateSector(ctx, sec); err != nil {
		return nil, err
	}
	return sec, nil
}

type CreateEventInput struct {
	Title       string
	Description string
	Genre       string
	Location    string
	VenueID     uuid.UUID
	OrganizerID uuid.UUID
	StartsAt    time.Time
}

func (s *CatalogService) CreateEvent(ctx context.Context, in CreateEventInput) (*model.Event, error) {
	venue, err := s.repo.GetVenue(ctx, in.VenueID)
	if err != nil {
		return nil, err
	}
	e := &model.Event{
		ID:          uuid.New(),
		Title:       in.Title,
		Description: in.Description,
		Genre:       in.Genre,
		Location:    in.Location,
		VenueID:     in.VenueID,
		OrganizerID: in.OrganizerID,
		StartsAt:    in.StartsAt,
		Status:      model.StatusDraft,
	}
	if err := s.repo.CreateEvent(ctx, e); err != nil {
		return nil, err
	}
	seats := buildSeats(e.ID, venue.Sectors)
	if err := s.repo.CreateSeats(ctx, seats); err != nil {
		return nil, err
	}
	return e, nil
}

func (s *CatalogService) PublishEvent(ctx context.Context, id uuid.UUID) error {
	return s.repo.UpdateEventStatus(ctx, id, model.StatusPublished)
}

type UpdateEventInput struct {
	Title       string
	Description string
	Genre       string
	Location    string
	StartsAt    time.Time
}

func (s *CatalogService) UpdateEvent(ctx context.Context, id uuid.UUID, in UpdateEventInput) (*model.Event, error) {
	fields := map[string]any{
		"title":       in.Title,
		"description": in.Description,
		"genre":       in.Genre,
		"location":    in.Location,
		"starts_at":   in.StartsAt,
	}
	if err := s.repo.UpdateEvent(ctx, id, fields); err != nil {
		return nil, err
	}
	return s.repo.GetEvent(ctx, id)
}

func (s *CatalogService) DeleteEvent(ctx context.Context, id uuid.UUID) error {
	e, err := s.repo.GetEvent(ctx, id)
	if err != nil {
		return err
	}
	if e.Status != model.StatusDraft {
		return ErrEventNotDraft
	}
	return s.repo.DeleteEvent(ctx, id)
}

func (s *CatalogService) SetPricing(ctx context.Context, eventID, sectorID uuid.UUID, amount int64, currency string) (*model.Pricing, error) {
	if currency == "" {
		currency = "USD"
	}
	p := &model.Pricing{ID: uuid.New(), EventID: eventID, SectorID: sectorID, Amount: amount, Currency: currency}
	if err := s.repo.CreatePricing(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *CatalogService) ListEvents(ctx context.Context, f repository.EventFilter) ([]model.Event, error) {
	return s.repo.ListEvents(ctx, f)
}

type EventDetail struct {
	Event   *model.Event    `json:"event"`
	Pricing []model.Pricing `json:"pricing"`
	Seats   []model.Seat    `json:"seats"`
}

func (s *CatalogService) GetEventDetail(ctx context.Context, id uuid.UUID) (*EventDetail, error) {
	e, err := s.repo.GetEvent(ctx, id)
	if err != nil {
		return nil, err
	}
	pricing, err := s.repo.ListPricing(ctx, id)
	if err != nil {
		return nil, err
	}
	seats, err := s.repo.ListSeats(ctx, id)
	if err != nil {
		return nil, err
	}
	return &EventDetail{Event: e, Pricing: pricing, Seats: seats}, nil
}

func buildSeats(eventID uuid.UUID, sectors []model.Sector) []model.Seat {
	var seats []model.Seat
	for _, sec := range sectors {
		for row := 1; row <= sec.RowCount; row++ {
			for col := 1; col <= sec.ColCount; col++ {
				seats = append(seats, model.Seat{
					ID:       uuid.New(),
					EventID:  eventID,
					SectorID: sec.ID,
					RowLabel: strconv.Itoa(row),
					Number:   col,
				})
			}
		}
	}
	return seats
}
