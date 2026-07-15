package clients

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type CatalogClient struct {
	baseURL string
	http    *http.Client
}

func NewCatalogClient(baseURL string, timeout time.Duration) *CatalogClient {
	return &CatalogClient{baseURL: baseURL, http: &http.Client{Timeout: timeout}}
}

type eventDetailResponse struct {
	Pricing []struct {
		SectorID uuid.UUID `json:"sector_id"`
		Amount   int64     `json:"amount"`
		Currency string    `json:"currency"`
	} `json:"pricing"`
	Seats []struct {
		ID       uuid.UUID `json:"id"`
		SectorID uuid.UUID `json:"sector_id"`
	} `json:"seats"`
}

type Quote struct {
	Amount   int64
	Currency string
}

func (c *CatalogClient) Price(ctx context.Context, eventID uuid.UUID, seatIDs []uuid.UUID) (*Quote, error) {
	var resp eventDetailResponse
	status, err := doJSON(ctx, c.http, http.MethodGet, c.baseURL+"/events/"+eventID.String(), nil, &resp)
	if err != nil {
		return nil, err
	}
	if status == http.StatusNotFound {
		return nil, fmt.Errorf("%w: event not found", ErrUpstream)
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("%w: catalog GetEvent status %d", ErrUpstream, status)
	}

	priceBySector := make(map[uuid.UUID]int64, len(resp.Pricing))
	currency := "USD"
	for _, p := range resp.Pricing {
		priceBySector[p.SectorID] = p.Amount
		if p.Currency != "" {
			currency = p.Currency
		}
	}
	sectorBySeat := make(map[uuid.UUID]uuid.UUID, len(resp.Seats))
	for _, s := range resp.Seats {
		sectorBySeat[s.ID] = s.SectorID
	}

	var total int64
	for _, sid := range seatIDs {
		sectorID, ok := sectorBySeat[sid]
		if !ok {
			return nil, fmt.Errorf("%w: seat %s not in event", ErrUpstream, sid)
		}
		total += priceBySector[sectorID]
	}
	return &Quote{Amount: total, Currency: currency}, nil
}
