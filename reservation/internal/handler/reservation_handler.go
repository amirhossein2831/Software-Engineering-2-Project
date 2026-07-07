package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"reservation/internal/repository"
	"reservation/internal/service"
)

type ReservationHandler struct {
	svc *service.ReservationService
}

func NewReservationHandler(svc *service.ReservationService) *ReservationHandler {
	return &ReservationHandler{svc: svc}
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

type holdRequest struct {
	EventID string   `json:"event_id"`
	SeatIDs []string `json:"seat_ids"`
}

func (h *ReservationHandler) Hold(c fiber.Ctx) error {
	uid, err := userID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "missing user")
	}
	var req holdRequest
	if err := c.Bind().Body(&req); err != nil || req.EventID == "" || len(req.SeatIDs) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "event_id and seat_ids are required")
	}
	eventID, err := uuid.Parse(req.EventID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid event_id")
	}
	seatIDs := make([]uuid.UUID, 0, len(req.SeatIDs))
	for _, raw := range req.SeatIDs {
		sid, err := uuid.Parse(raw)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid seat_id")
		}
		seatIDs = append(seatIDs, sid)
	}
	res, err := h.svc.Hold(c.Context(), eventID, uid, seatIDs)
	if errors.Is(err, repository.ErrConflict) {
		return fiber.NewError(fiber.StatusConflict, "one or more seats are not available")
	}
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not hold seats")
	}
	return c.Status(fiber.StatusCreated).JSON(res)
}

func (h *ReservationHandler) Commit(c fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	res, err := h.svc.Commit(c.Context(), id)
	if errors.Is(err, repository.ErrNotFound) {
		return fiber.NewError(fiber.StatusNotFound, "hold not found")
	}
	if errors.Is(err, repository.ErrExpired) {
		return fiber.NewError(fiber.StatusConflict, "hold is not active")
	}
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not commit hold")
	}
	return c.JSON(res)
}

func (h *ReservationHandler) Release(c fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	res, err := h.svc.Release(c.Context(), id)
	if errors.Is(err, repository.ErrNotFound) {
		return fiber.NewError(fiber.StatusNotFound, "hold not found")
	}
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not release hold")
	}
	return c.JSON(res)
}

func (h *ReservationHandler) GetHold(c fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	detail, err := h.svc.Get(c.Context(), id)
	if errors.Is(err, repository.ErrNotFound) {
		return fiber.NewError(fiber.StatusNotFound, "hold not found")
	}
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "error")
	}
	return c.JSON(detail)
}

func (h *ReservationHandler) SeatStates(c fiber.Ctx) error {
	eventID, err := parseID(c, "eventID")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid event id")
	}
	states, err := h.svc.SeatStates(c.Context(), eventID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "error")
	}
	return c.JSON(fiber.Map{"seats": states})
}
