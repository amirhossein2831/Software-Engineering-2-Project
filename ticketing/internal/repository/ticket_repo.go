package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"ticketing/internal/model"
)

var ErrNotFound = errors.New("ticket not found")

type TicketRepo struct {
	db *gorm.DB
}

func NewTicketRepo(db *gorm.DB) *TicketRepo {
	return &TicketRepo{db: db}
}

func (r *TicketRepo) CreateIgnoreConflict(ctx context.Context, tickets []model.Ticket) (int, error) {
	if len(tickets) == 0 {
		return 0, nil
	}
	res := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&tickets)
	return int(res.RowsAffected), res.Error
}

func (r *TicketRepo) Get(ctx context.Context, id uuid.UUID) (*model.Ticket, error) {
	var t model.Ticket
	if err := r.db.WithContext(ctx).First(&t, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &t, nil
}

func (r *TicketRepo) ListByOrder(ctx context.Context, orderID uuid.UUID) ([]model.Ticket, error) {
	var tickets []model.Ticket
	err := r.db.WithContext(ctx).
		Where("order_id = ?", orderID).
		Order("issued_at ASC").
		Find(&tickets).Error
	return tickets, err
}
