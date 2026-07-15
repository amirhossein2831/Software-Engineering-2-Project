package handler

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"catalog/internal/repository"
	"catalog/internal/service"
)

type CatalogHandler struct {
	svc *service.CatalogService
}

func NewCatalogHandler(svc *service.CatalogService) *CatalogHandler {
	return &CatalogHandler{svc: svc}
}

func userID(c fiber.Ctx) (uuid.UUID, error) {
	raw := c.Get("X-User-Id")
	if raw == "" {
		return uuid.Nil, errors.New("missing user")
	}
	return uuid.Parse(raw)
}

func parseID(c fiber.Ctx, name string) (uuid.UUID, error) {
	return uuid.Parse(c.Params(name))
}

type createVenueRequest struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

func (h *CatalogHandler) CreateVenue(c fiber.Ctx) error {
	uid, err := userID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "missing user")
	}
	var req createVenueRequest
	if err := c.Bind().Body(&req); err != nil || req.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "name is required")
	}
	v, err := h.svc.CreateVenue(c.Context(), req.Name, req.Address, uid)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not create venue")
	}
	return c.Status(fiber.StatusCreated).JSON(v)
}

func (h *CatalogHandler) GetVenue(c fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	v, err := h.svc.GetVenue(c.Context(), id)
	if errors.Is(err, repository.ErrNotFound) {
		return fiber.NewError(fiber.StatusNotFound, "venue not found")
	}
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "error")
	}
	return c.JSON(v)
}

func (h *CatalogHandler) ListVenues(c fiber.Ctx) error {
	venues, err := h.svc.ListVenues(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "error")
	}
	return c.JSON(fiber.Map{"venues": venues})
}

func (h *CatalogHandler) UpdateVenue(c fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	var req createVenueRequest
	if err := c.Bind().Body(&req); err != nil || req.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "name is required")
	}
	v, err := h.svc.UpdateVenue(c.Context(), id, req.Name, req.Address)
	if errors.Is(err, repository.ErrNotFound) {
		return fiber.NewError(fiber.StatusNotFound, "venue not found")
	}
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not update venue")
	}
	return c.JSON(v)
}

func (h *CatalogHandler) DeleteVenue(c fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	err = h.svc.DeleteVenue(c.Context(), id)
	if errors.Is(err, service.ErrVenueInUse) {
		return fiber.NewError(fiber.StatusConflict, "venue has events; delete or reassign them first")
	}
	if errors.Is(err, repository.ErrNotFound) {
		return fiber.NewError(fiber.StatusNotFound, "venue not found")
	}
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not delete venue")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *CatalogHandler) DeleteSector(c fiber.Ctx) error {
	venueID, err := parseID(c, "id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	sectorID, err := parseID(c, "sectorId")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid sector id")
	}
	err = h.svc.DeleteSector(c.Context(), venueID, sectorID)
	if errors.Is(err, repository.ErrNotFound) {
		return fiber.NewError(fiber.StatusNotFound, "sector not found")
	}
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not delete sector")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

type createSectorRequest struct {
	Name     string `json:"name"`
	RowCount int    `json:"row_count"`
	ColCount int    `json:"col_count"`
}

func (h *CatalogHandler) AddSector(c fiber.Ctx) error {
	venueID, err := parseID(c, "id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	var req createSectorRequest
	if err := c.Bind().Body(&req); err != nil || req.Name == "" || req.RowCount <= 0 || req.ColCount <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "name, row_count, col_count are required")
	}
	sec, err := h.svc.AddSector(c.Context(), venueID, req.Name, req.RowCount, req.ColCount)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not add sector")
	}
	return c.Status(fiber.StatusCreated).JSON(sec)
}

type createEventRequest struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Genre       string    `json:"genre"`
	Location    string    `json:"location"`
	VenueID     string    `json:"venue_id"`
	StartsAt    time.Time `json:"starts_at"`
}

func (h *CatalogHandler) CreateEvent(c fiber.Ctx) error {
	uid, err := userID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "missing user")
	}
	var req createEventRequest
	if err := c.Bind().Body(&req); err != nil || req.Title == "" || req.VenueID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "title and venue_id are required")
	}
	venueID, err := uuid.Parse(req.VenueID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid venue_id")
	}
	e, err := h.svc.CreateEvent(c.Context(), service.CreateEventInput{
		Title:       req.Title,
		Description: req.Description,
		Genre:       req.Genre,
		Location:    req.Location,
		VenueID:     venueID,
		OrganizerID: uid,
		StartsAt:    req.StartsAt,
	})
	if errors.Is(err, repository.ErrNotFound) {
		return fiber.NewError(fiber.StatusBadRequest, "venue not found")
	}
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not create event")
	}
	return c.Status(fiber.StatusCreated).JSON(e)
}

func (h *CatalogHandler) PublishEvent(c fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	if err := h.svc.PublishEvent(c.Context(), id); errors.Is(err, repository.ErrNotFound) {
		return fiber.NewError(fiber.StatusNotFound, "event not found")
	} else if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not publish")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

type updateEventRequest struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Genre       string    `json:"genre"`
	Location    string    `json:"location"`
	StartsAt    time.Time `json:"starts_at"`
}

func (h *CatalogHandler) UpdateEvent(c fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	var req updateEventRequest
	if err := c.Bind().Body(&req); err != nil || req.Title == "" {
		return fiber.NewError(fiber.StatusBadRequest, "title is required")
	}
	e, err := h.svc.UpdateEvent(c.Context(), id, service.UpdateEventInput{
		Title:       req.Title,
		Description: req.Description,
		Genre:       req.Genre,
		Location:    req.Location,
		StartsAt:    req.StartsAt,
	})
	if errors.Is(err, repository.ErrNotFound) {
		return fiber.NewError(fiber.StatusNotFound, "event not found")
	}
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not update event")
	}
	return c.JSON(e)
}

func (h *CatalogHandler) DeleteEvent(c fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	err = h.svc.DeleteEvent(c.Context(), id)
	if errors.Is(err, service.ErrEventNotDraft) {
		return fiber.NewError(fiber.StatusConflict, "only draft events can be deleted")
	}
	if errors.Is(err, repository.ErrNotFound) {
		return fiber.NewError(fiber.StatusNotFound, "event not found")
	}
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not delete event")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

type pricingRequest struct {
	SectorID string `json:"sector_id"`
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

func (h *CatalogHandler) SetPricing(c fiber.Ctx) error {
	eventID, err := parseID(c, "id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	var req pricingRequest
	if err := c.Bind().Body(&req); err != nil || req.SectorID == "" || req.Amount < 0 {
		return fiber.NewError(fiber.StatusBadRequest, "sector_id and amount are required")
	}
	sectorID, err := uuid.Parse(req.SectorID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid sector_id")
	}
	p, err := h.svc.SetPricing(c.Context(), eventID, sectorID, req.Amount, req.Currency)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not set pricing")
	}
	return c.Status(fiber.StatusCreated).JSON(p)
}

func (h *CatalogHandler) ListEvents(c fiber.Ctx) error {
	events, err := h.svc.ListEvents(c.Context(), repository.EventFilter{
		Genre:         c.Query("genre"),
		Location:      c.Query("location"),
		OnlyPublished: c.Query("include_drafts") != "true",
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "error")
	}
	return c.JSON(fiber.Map{"events": events})
}

func (h *CatalogHandler) GetEvent(c fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	detail, err := h.svc.GetEventDetail(c.Context(), id)
	if errors.Is(err, repository.ErrNotFound) {
		return fiber.NewError(fiber.StatusNotFound, "event not found")
	}
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "error")
	}
	return c.JSON(detail)
}
