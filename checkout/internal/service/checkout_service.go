package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"checkout/internal/clients"
	"checkout/internal/events"
	"checkout/internal/model"
	"checkout/internal/payment"
	"checkout/internal/repository"
)

const eventsTopic = "checkout.events"

var (
	ErrHoldInvalid = errors.New("hold is not valid for checkout")
	ErrNotFound    = repository.ErrNotFound
)

type CheckoutService struct {
	repo        *repository.OrderRepo
	reservation *clients.ReservationClient
	catalog     *clients.CatalogClient
	gateway     payment.Gateway
	pub         *events.Publisher
}

func NewCheckoutService(repo *repository.OrderRepo, reservation *clients.ReservationClient, catalog *clients.CatalogClient, gateway payment.Gateway, pub *events.Publisher) *CheckoutService {
	return &CheckoutService{repo: repo, reservation: reservation, catalog: catalog, gateway: gateway, pub: pub}
}

type OrderEvent struct {
	Type    string      `json:"type"`
	OrderID uuid.UUID   `json:"order_id"`
	UserID  uuid.UUID   `json:"user_id"`
	EventID uuid.UUID   `json:"event_id"`
	SeatIDs []uuid.UUID `json:"seat_ids"`
	Amount  int64       `json:"amount"`
}

func (s *CheckoutService) Checkout(ctx context.Context, userID, holdID uuid.UUID, idempotencyKey string) (*model.Order, error) {
	if idempotencyKey != "" {
		existing, err := s.repo.GetByIdempotencyKey(ctx, idempotencyKey)
		if err == nil {
			return existing, nil
		}
		if !errors.Is(err, repository.ErrNotFound) {
			return nil, err
		}
	}

	hold, err := s.reservation.GetHold(ctx, holdID)
	if errors.Is(err, clients.ErrHoldNotFound) {
		return nil, ErrHoldInvalid
	}
	if err != nil {
		return nil, err
	}
	if hold.UserID != userID || hold.Status != "locked" || time.Now().After(hold.ExpiresAt) {
		return nil, ErrHoldInvalid
	}

	quote, err := s.catalog.Price(ctx, hold.EventID, hold.SeatIDs)
	if err != nil {
		return nil, err
	}

	order := &model.Order{
		ID:             uuid.New(),
		UserID:         userID,
		EventID:        hold.EventID,
		HoldID:         holdID,
		SeatIDs:        hold.SeatIDs,
		Amount:         quote.Amount,
		Currency:       quote.Currency,
		Status:         model.OrderPending,
		IdempotencyKey: idempotencyKey,
	}
	if err := s.repo.Create(ctx, order); err != nil {
		if idempotencyKey != "" && errors.Is(err, gorm.ErrDuplicatedKey) {
			return s.repo.GetByIdempotencyKey(ctx, idempotencyKey)
		}
		return nil, err
	}
	_ = s.repo.Log(ctx, order.ID, "order", "created")
	s.publish(ctx, order, "order.created")

	result, err := s.gateway.Charge(ctx, payment.Charge{OrderID: order.ID, Amount: order.Amount, Currency: order.Currency})
	if err != nil {
		return s.failOrder(ctx, order, "mock", ""), nil
	}
	_ = s.repo.SavePayment(ctx, &model.Payment{
		ID:          uuid.New(),
		OrderID:     order.ID,
		Provider:    result.Provider,
		ProviderRef: result.ProviderRef,
		Status:      model.PaymentSucceeded,
		Amount:      order.Amount,
	})
	_ = s.repo.Log(ctx, order.ID, "charge", "succeeded")

	if err := s.reservation.Commit(ctx, holdID); err != nil {
		return s.compensate(ctx, order, result), nil
	}

	_ = s.repo.UpdateStatus(ctx, order.ID, model.OrderPaid)
	_ = s.repo.Log(ctx, order.ID, "commit", "succeeded")
	order.Status = model.OrderPaid
	s.publish(ctx, order, "payment.succeeded")
	return order, nil
}

func (s *CheckoutService) failOrder(ctx context.Context, order *model.Order, provider, providerRef string) *model.Order {
	_ = s.repo.SavePayment(ctx, &model.Payment{
		ID:          uuid.New(),
		OrderID:     order.ID,
		Provider:    provider,
		ProviderRef: providerRef,
		Status:      model.PaymentFailed,
		Amount:      order.Amount,
	})
	_ = s.reservation.Release(ctx, order.HoldID)
	_ = s.repo.UpdateStatus(ctx, order.ID, model.OrderFailed)
	_ = s.repo.Log(ctx, order.ID, "charge", "failed")
	order.Status = model.OrderFailed
	s.publish(ctx, order, "payment.failed")
	return order
}

func (s *CheckoutService) compensate(ctx context.Context, order *model.Order, result *payment.Result) *model.Order {
	_ = s.gateway.Refund(ctx, result.ProviderRef)
	_ = s.repo.SavePayment(ctx, &model.Payment{
		ID:          uuid.New(),
		OrderID:     order.ID,
		Provider:    result.Provider,
		ProviderRef: result.ProviderRef,
		Status:      model.PaymentRefunded,
		Amount:      order.Amount,
	})
	_ = s.reservation.Release(ctx, order.HoldID)
	_ = s.repo.UpdateStatus(ctx, order.ID, model.OrderCompensated)
	_ = s.repo.Log(ctx, order.ID, "commit", "failed")
	order.Status = model.OrderCompensated
	s.publish(ctx, order, "payment.failed")
	return order
}

type OrderDetail struct {
	Order    *model.Order    `json:"order"`
	Payments []model.Payment `json:"payments"`
}

func (s *CheckoutService) GetOrder(ctx context.Context, id uuid.UUID) (*OrderDetail, error) {
	order, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	payments, err := s.repo.Payments(ctx, id)
	if err != nil {
		return nil, err
	}
	return &OrderDetail{Order: order, Payments: payments}, nil
}

func (s *CheckoutService) ListOrders(ctx context.Context, userID uuid.UUID) ([]model.Order, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *CheckoutService) publish(ctx context.Context, order *model.Order, eventType string) {
	_ = s.pub.Publish(ctx, eventsTopic, order.ID.String(), OrderEvent{
		Type:    eventType,
		OrderID: order.ID,
		UserID:  order.UserID,
		EventID: order.EventID,
		SeatIDs: order.SeatIDs,
		Amount:  order.Amount,
	})
}
