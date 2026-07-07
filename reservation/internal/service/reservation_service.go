package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"reservation/internal/events"
	"reservation/internal/model"
	"reservation/internal/redislock"
	"reservation/internal/repository"
)

const eventsTopic = "reservation.events"

type ReservationService struct {
	repo   *repository.ReservationRepo
	locker *redislock.Locker
	pub    *events.Publisher
}

func NewReservationService(repo *repository.ReservationRepo, locker *redislock.Locker, pub *events.Publisher) *ReservationService {
	return &ReservationService{repo: repo, locker: locker, pub: pub}
}

type SeatEvent struct {
	Type    string      `json:"type"`
	HoldID  uuid.UUID   `json:"hold_id"`
	EventID uuid.UUID   `json:"event_id"`
	UserID  uuid.UUID   `json:"user_id,omitempty"`
	SeatIDs []uuid.UUID `json:"seat_ids"`
}

func seatKeys(eventID uuid.UUID, seatIDs []uuid.UUID) []string {
	keys := make([]string, 0, len(seatIDs))
	for _, sid := range seatIDs {
		keys = append(keys, redislock.SeatKey(eventID, sid))
	}
	return keys
}

func (s *ReservationService) Hold(ctx context.Context, eventID, userID uuid.UUID, seatIDs []uuid.UUID) (*model.Reservation, error) {
	res := &model.Reservation{
		ID:        uuid.New(),
		EventID:   eventID,
		UserID:    userID,
		Status:    model.HoldLocked,
		ExpiresAt: time.Now().Add(s.locker.TTL()),
	}
	keys := seatKeys(eventID, seatIDs)
	ok, err := s.locker.Acquire(ctx, keys, res.ID.String())
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, repository.ErrConflict
	}
	if err := s.repo.Lock(ctx, res, seatIDs); err != nil {
		_ = s.locker.Release(ctx, keys)
		return nil, err
	}
	s.publish(ctx, SeatEvent{Type: "seat.locked", HoldID: res.ID, EventID: eventID, UserID: userID, SeatIDs: seatIDs})
	return res, nil
}

func (s *ReservationService) Commit(ctx context.Context, holdID uuid.UUID) (*model.Reservation, error) {
	seatIDs, err := s.repo.SeatIDsByHold(ctx, holdID)
	if err != nil {
		return nil, err
	}
	res, err := s.repo.Commit(ctx, holdID)
	if err != nil {
		return nil, err
	}
	_ = s.locker.Release(ctx, seatKeys(res.EventID, seatIDs))
	s.publish(ctx, SeatEvent{Type: "seat.booked", HoldID: res.ID, EventID: res.EventID, UserID: res.UserID, SeatIDs: seatIDs})
	return res, nil
}

func (s *ReservationService) Release(ctx context.Context, holdID uuid.UUID) (*model.Reservation, error) {
	seatIDs, err := s.repo.SeatIDsByHold(ctx, holdID)
	if err != nil {
		return nil, err
	}
	res, err := s.repo.Release(ctx, holdID)
	if err != nil {
		return nil, err
	}
	_ = s.locker.Release(ctx, seatKeys(res.EventID, seatIDs))
	s.publish(ctx, SeatEvent{Type: "seat.released", HoldID: res.ID, EventID: res.EventID, UserID: res.UserID, SeatIDs: seatIDs})
	return res, nil
}

type HoldDetail struct {
	Reservation *model.Reservation `json:"reservation"`
	SeatIDs     []uuid.UUID        `json:"seat_ids"`
}

func (s *ReservationService) Get(ctx context.Context, holdID uuid.UUID) (*HoldDetail, error) {
	res, err := s.repo.Get(ctx, holdID)
	if err != nil {
		return nil, err
	}
	seatIDs, err := s.repo.SeatIDsByHold(ctx, holdID)
	if err != nil {
		return nil, err
	}
	return &HoldDetail{Reservation: res, SeatIDs: seatIDs}, nil
}

func (s *ReservationService) SeatStates(ctx context.Context, eventID uuid.UUID) ([]model.SeatState, error) {
	return s.repo.SeatStates(ctx, eventID)
}

func (s *ReservationService) SweepExpired(ctx context.Context) (int, error) {
	holds, err := s.repo.ExpiredHolds(ctx, time.Now())
	if err != nil {
		return 0, err
	}
	for _, h := range holds {
		if _, err := s.Release(ctx, h.ID); err != nil {
			return 0, err
		}
	}
	return len(holds), nil
}

func (s *ReservationService) publish(ctx context.Context, ev SeatEvent) {
	_ = s.pub.Publish(ctx, eventsTopic, ev.EventID.String(), ev)
}
