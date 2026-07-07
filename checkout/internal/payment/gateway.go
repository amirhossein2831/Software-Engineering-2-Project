package payment

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

var ErrDeclined = errors.New("payment declined")

type Charge struct {
	OrderID  uuid.UUID
	Amount   int64
	Currency string
}

type Result struct {
	Provider    string
	ProviderRef string
}

type Gateway interface {
	Charge(ctx context.Context, c Charge) (*Result, error)
	Refund(ctx context.Context, providerRef string) error
}

type MockGateway struct {
	failRate float64
	latency  time.Duration
	rng      *rand.Rand
}

func NewMockGateway(failRate float64, latency time.Duration) *MockGateway {
	return &MockGateway{
		failRate: failRate,
		latency:  latency,
		rng:      rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (g *MockGateway) Charge(ctx context.Context, c Charge) (*Result, error) {
	if g.latency > 0 {
		select {
		case <-time.After(g.latency):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	if g.rng.Float64() < g.failRate {
		return nil, ErrDeclined
	}
	return &Result{Provider: "mock", ProviderRef: uuid.NewString()}, nil
}

func (g *MockGateway) Refund(ctx context.Context, providerRef string) error {
	return nil
}
