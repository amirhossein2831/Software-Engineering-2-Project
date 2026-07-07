package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"notification/internal/model"
)

type NotificationRepo struct {
	db *gorm.DB
}

func NewNotificationRepo(db *gorm.DB) *NotificationRepo {
	return &NotificationRepo{db: db}
}

func (r *NotificationRepo) Enqueue(ctx context.Context, n *model.NotificationOutbox) (bool, error) {
	res := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "event_key"}}, DoNothing: true}).
		Create(n)
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

func (r *NotificationRepo) UpsertRecipient(ctx context.Context, rec *model.Recipient) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"email", "phone", "updated_at"}),
		}).
		Create(rec).Error
}

func (r *NotificationRepo) GetRecipient(ctx context.Context, userID uuid.UUID) (*model.Recipient, error) {
	var rec model.Recipient
	err := r.db.WithContext(ctx).First(&rec, "user_id = ?", userID).Error
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *NotificationRepo) PendingBatch(ctx context.Context, limit int) ([]model.NotificationOutbox, error) {
	var rows []model.NotificationOutbox
	err := r.db.WithContext(ctx).
		Where("status = ?", model.StatusPending).
		Order("created_at ASC").
		Limit(limit).
		Find(&rows).Error
	return rows, err
}

func (r *NotificationRepo) MarkSent(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&model.NotificationOutbox{}).
		Where("id = ?", id).
		Updates(map[string]any{"status": model.StatusSent, "updated_at": time.Now()}).Error
}

func (r *NotificationRepo) MarkAttempt(ctx context.Context, id uuid.UUID, attempts int, status model.Status) error {
	return r.db.WithContext(ctx).
		Model(&model.NotificationOutbox{}).
		Where("id = ?", id).
		Updates(map[string]any{"attempts": attempts, "status": status, "updated_at": time.Now()}).Error
}
