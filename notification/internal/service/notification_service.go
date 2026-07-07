package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"notification/internal/model"
	"notification/internal/repository"
)

type NotificationService struct {
	repo *repository.NotificationRepo
}

func NewNotificationService(repo *repository.NotificationRepo) *NotificationService {
	return &NotificationService{repo: repo}
}

type envelope struct {
	Type     string    `json:"type"`
	OrderID  uuid.UUID `json:"order_id"`
	TicketID uuid.UUID `json:"ticket_id"`
	UserID   uuid.UUID `json:"user_id"`
	Email    string    `json:"email"`
	Phone    string    `json:"phone"`
}

func (s *NotificationService) Handle(ctx context.Context, value []byte) error {
	var ev envelope
	if err := json.Unmarshal(value, &ev); err != nil {
		return nil
	}
	switch ev.Type {
	case "user.registered":
		return s.repo.UpsertRecipient(ctx, &model.Recipient{
			UserID:    ev.UserID,
			Email:     ev.Email,
			Phone:     ev.Phone,
			UpdatedAt: time.Now(),
		})
	case "payment.succeeded":
		return s.enqueue(ctx, ev, "payment_succeeded", "payment.succeeded:"+ev.OrderID.String(), value)
	case "payment.failed":
		return s.enqueue(ctx, ev, "payment_failed", "payment.failed:"+ev.OrderID.String(), value)
	case "ticket.issued":
		return s.enqueue(ctx, ev, "ticket_issued", "ticket.issued:"+ev.TicketID.String(), value)
	default:
		return nil
	}
}

func (s *NotificationService) enqueue(ctx context.Context, ev envelope, template, eventKey string, payload []byte) error {
	_, err := s.repo.Enqueue(ctx, &model.NotificationOutbox{
		ID:       uuid.New(),
		EventKey: eventKey,
		Channel:  model.ChannelEmail,
		To:       s.resolve(ctx, ev.UserID),
		Template: template,
		Payload:  string(payload),
		Status:   model.StatusPending,
	})
	return err
}

func (s *NotificationService) resolve(ctx context.Context, userID uuid.UUID) string {
	rec, err := s.repo.GetRecipient(ctx, userID)
	if err == nil && rec.Email != "" {
		return rec.Email
	}
	return "user:" + userID.String()
}
