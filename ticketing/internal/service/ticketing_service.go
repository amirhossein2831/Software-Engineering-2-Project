package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"ticketing/internal/events"
	"ticketing/internal/model"
	"ticketing/internal/qr"
	"ticketing/internal/repository"
)

const ticketsTopic = "ticketing.events"

type TicketingService struct {
	repo   *repository.TicketRepo
	signer *qr.Signer
	pub    *events.Publisher
}

func NewTicketingService(repo *repository.TicketRepo, signer *qr.Signer, pub *events.Publisher) *TicketingService {
	return &TicketingService{repo: repo, signer: signer, pub: pub}
}

type paymentSucceeded struct {
	Type    string      `json:"type"`
	OrderID uuid.UUID   `json:"order_id"`
	UserID  uuid.UUID   `json:"user_id"`
	EventID uuid.UUID   `json:"event_id"`
	SeatIDs []uuid.UUID `json:"seat_ids"`
}

type TicketEvent struct {
	Type     string    `json:"type"`
	TicketID uuid.UUID `json:"ticket_id"`
	OrderID  uuid.UUID `json:"order_id"`
	EventID  uuid.UUID `json:"event_id"`
	SeatID   uuid.UUID `json:"seat_id"`
	UserID   uuid.UUID `json:"user_id"`
	QRHash   string    `json:"qr_hash"`
}

func (s *TicketingService) Handle(ctx context.Context, value []byte) error {
	var ev paymentSucceeded
	if err := json.Unmarshal(value, &ev); err != nil {
		return nil
	}
	if ev.Type != "payment.succeeded" || len(ev.SeatIDs) == 0 {
		return nil
	}
	return s.issue(ctx, ev)
}

func (s *TicketingService) issue(ctx context.Context, ev paymentSucceeded) error {
	tickets := make([]model.Ticket, 0, len(ev.SeatIDs))
	now := time.Now()
	for _, seatID := range ev.SeatIDs {
		tickets = append(tickets, model.Ticket{
			ID:       uuid.New(),
			OrderID:  ev.OrderID,
			EventID:  ev.EventID,
			SeatID:   seatID,
			UserID:   ev.UserID,
			QRHash:   s.signer.Hash(ev.OrderID, ev.EventID, seatID),
			IssuedAt: now,
		})
	}
	n, err := s.repo.CreateIgnoreConflict(ctx, tickets)
	if err != nil {
		return err
	}
	if n == 0 {
		return nil
	}
	for _, t := range tickets {
		_ = s.pub.Publish(ctx, ticketsTopic, t.OrderID.String(), TicketEvent{
			Type:     "ticket.issued",
			TicketID: t.ID,
			OrderID:  t.OrderID,
			EventID:  t.EventID,
			SeatID:   t.SeatID,
			UserID:   t.UserID,
			QRHash:   t.QRHash,
		})
	}
	return nil
}

func (s *TicketingService) GetTicket(ctx context.Context, id uuid.UUID) (*model.Ticket, error) {
	return s.repo.Get(ctx, id)
}

func (s *TicketingService) ListByOrder(ctx context.Context, orderID uuid.UUID) ([]model.Ticket, error) {
	return s.repo.ListByOrder(ctx, orderID)
}
