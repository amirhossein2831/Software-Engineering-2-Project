package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"reservation/internal/model"
)

var (
	ErrNotFound = errors.New("reservation not found")
	ErrConflict = errors.New("one or more seats are not available")
	ErrExpired  = errors.New("hold is not active")
)

type ReservationRepo struct {
	db *gorm.DB
}

func NewReservationRepo(db *gorm.DB) *ReservationRepo {
	return &ReservationRepo{db: db}
}

func (r *ReservationRepo) Lock(ctx context.Context, res *model.Reservation, seatIDs []uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var taken int64
		if err := tx.Model(&model.SeatState{}).
			Where("event_id = ? AND seat_id IN ? AND status <> ?", res.EventID, seatIDs, model.SeatAvailable).
			Count(&taken).Error; err != nil {
			return err
		}
		if taken > 0 {
			return ErrConflict
		}
		if err := tx.Create(res).Error; err != nil {
			return err
		}
		states := make([]model.SeatState, 0, len(seatIDs))
		for _, sid := range seatIDs {
			hold := res.ID
			states = append(states, model.SeatState{
				EventID: res.EventID,
				SeatID:  sid,
				Status:  model.SeatLocked,
				HoldID:  &hold,
			})
		}
		return tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "event_id"}, {Name: "seat_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"status", "hold_id", "updated_at"}),
		}).Create(&states).Error
	})
}

func (r *ReservationRepo) Commit(ctx context.Context, holdID uuid.UUID) (*model.Reservation, error) {
	var res model.Reservation
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&res, "id = ?", holdID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return err
		}
		if res.Status != model.HoldLocked || time.Now().After(res.ExpiresAt) {
			return ErrExpired
		}
		if err := tx.Model(&model.SeatState{}).
			Where("hold_id = ?", holdID).
			Update("status", model.SeatBooked).Error; err != nil {
			return err
		}
		res.Status = model.HoldCommitted
		return tx.Model(&res).Update("status", model.HoldCommitted).Error
	})
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (r *ReservationRepo) Release(ctx context.Context, holdID uuid.UUID) (*model.Reservation, error) {
	var res model.Reservation
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&res, "id = ?", holdID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return err
		}
		if res.Status != model.HoldLocked {
			return nil
		}
		if err := tx.Model(&model.SeatState{}).
			Where("hold_id = ?", holdID).
			Updates(map[string]any{"status": model.SeatAvailable, "hold_id": nil}).Error; err != nil {
			return err
		}
		res.Status = model.HoldReleased
		return tx.Model(&res).Update("status", model.HoldReleased).Error
	})
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (r *ReservationRepo) Get(ctx context.Context, id uuid.UUID) (*model.Reservation, error) {
	var res model.Reservation
	if err := r.db.WithContext(ctx).First(&res, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &res, nil
}

func (r *ReservationRepo) SeatIDsByHold(ctx context.Context, holdID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.db.WithContext(ctx).Model(&model.SeatState{}).
		Where("hold_id = ?", holdID).
		Pluck("seat_id", &ids).Error
	return ids, err
}

func (r *ReservationRepo) SeatStates(ctx context.Context, eventID uuid.UUID) ([]model.SeatState, error) {
	var states []model.SeatState
	err := r.db.WithContext(ctx).Where("event_id = ?", eventID).Find(&states).Error
	return states, err
}

func (r *ReservationRepo) ExpiredHolds(ctx context.Context, now time.Time) ([]model.Reservation, error) {
	var holds []model.Reservation
	err := r.db.WithContext(ctx).
		Where("status = ? AND expires_at < ?", model.HoldLocked, now).
		Find(&holds).Error
	return holds, err
}
