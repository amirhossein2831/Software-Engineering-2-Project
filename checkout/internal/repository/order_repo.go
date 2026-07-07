package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"checkout/internal/model"
)

var ErrNotFound = errors.New("order not found")

type OrderRepo struct {
	db *gorm.DB
}

func NewOrderRepo(db *gorm.DB) *OrderRepo {
	return &OrderRepo{db: db}
}

func (r *OrderRepo) Create(ctx context.Context, o *model.Order) error {
	return r.db.WithContext(ctx).Create(o).Error
}

func (r *OrderRepo) Get(ctx context.Context, id uuid.UUID) (*model.Order, error) {
	var o model.Order
	if err := r.db.WithContext(ctx).First(&o, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &o, nil
}

func (r *OrderRepo) GetByIdempotencyKey(ctx context.Context, key string) (*model.Order, error) {
	var o model.Order
	if err := r.db.WithContext(ctx).First(&o, "idempotency_key = ?", key).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &o, nil
}

func (r *OrderRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]model.Order, error) {
	var orders []model.Order
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&orders).Error
	return orders, err
}

func (r *OrderRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status model.OrderStatus) error {
	return r.db.WithContext(ctx).
		Model(&model.Order{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *OrderRepo) SavePayment(ctx context.Context, p *model.Payment) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *OrderRepo) Log(ctx context.Context, orderID uuid.UUID, step, state string) error {
	return r.db.WithContext(ctx).Create(&model.SagaLog{
		ID:      uuid.New(),
		OrderID: orderID,
		Step:    step,
		State:   state,
	}).Error
}

func (r *OrderRepo) Payments(ctx context.Context, orderID uuid.UUID) ([]model.Payment, error) {
	var payments []model.Payment
	err := r.db.WithContext(ctx).
		Where("order_id = ?", orderID).
		Order("created_at ASC").
		Find(&payments).Error
	return payments, err
}
