package clients

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

var (
	ErrHoldNotFound = errors.New("hold not found")
	ErrUpstream     = errors.New("upstream request failed")
)

type Hold struct {
	ID        uuid.UUID
	EventID   uuid.UUID
	UserID    uuid.UUID
	Status    string
	ExpiresAt time.Time
	SeatIDs   []uuid.UUID
}

type ReservationClient struct {
	baseURL string
	http    *http.Client
}

func NewReservationClient(baseURL string, timeout time.Duration) *ReservationClient {
	return &ReservationClient{baseURL: baseURL, http: &http.Client{Timeout: timeout}}
}

type holdDetailResponse struct {
	Reservation struct {
		ID        uuid.UUID `json:"id"`
		EventID   uuid.UUID `json:"event_id"`
		UserID    uuid.UUID `json:"user_id"`
		Status    string    `json:"status"`
		ExpiresAt time.Time `json:"expires_at"`
	} `json:"reservation"`
	SeatIDs []uuid.UUID `json:"seat_ids"`
}

func (c *ReservationClient) GetHold(ctx context.Context, holdID uuid.UUID) (*Hold, error) {
	var resp holdDetailResponse
	status, err := doJSON(ctx, c.http, http.MethodGet, c.baseURL+"/holds/"+holdID.String(), nil, &resp)
	if err != nil {
		return nil, err
	}
	if status == http.StatusNotFound {
		return nil, ErrHoldNotFound
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("%w: reservation GetHold status %d", ErrUpstream, status)
	}
	return &Hold{
		ID:        resp.Reservation.ID,
		EventID:   resp.Reservation.EventID,
		UserID:    resp.Reservation.UserID,
		Status:    resp.Reservation.Status,
		ExpiresAt: resp.Reservation.ExpiresAt,
		SeatIDs:   resp.SeatIDs,
	}, nil
}

func (c *ReservationClient) Commit(ctx context.Context, holdID uuid.UUID) error {
	status, err := doJSON(ctx, c.http, http.MethodPost, c.baseURL+"/holds/"+holdID.String()+"/commit", nil, nil)
	if err != nil {
		return err
	}
	if status != http.StatusOK {
		return fmt.Errorf("%w: reservation Commit status %d", ErrUpstream, status)
	}
	return nil
}

func (c *ReservationClient) Release(ctx context.Context, holdID uuid.UUID) error {
	status, err := doJSON(ctx, c.http, http.MethodPost, c.baseURL+"/holds/"+holdID.String()+"/release", nil, nil)
	if err != nil {
		return err
	}
	if status != http.StatusOK {
		return fmt.Errorf("%w: reservation Release status %d", ErrUpstream, status)
	}
	return nil
}
