package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"

	"waitingroom/internal/service"
)

type WaitingHandler struct {
	svc *service.WaitingService
}

func NewWaitingHandler(svc *service.WaitingService) *WaitingHandler {
	return &WaitingHandler{svc: svc}
}

func userID(c fiber.Ctx) (string, error) {
	raw := c.Get("X-User-Id")
	if raw == "" {
		return "", fiber.NewError(fiber.StatusUnauthorized, "missing user")
	}
	return raw, nil
}

func (h *WaitingHandler) Join(c fiber.Ctx) error {
	eventID := c.Params("eventID")
	if eventID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "event id is required")
	}
	uid, err := userID(c)
	if err != nil {
		return err
	}
	res, err := h.svc.Join(c.Context(), eventID, uid)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not join queue")
	}
	return c.Status(fiber.StatusCreated).JSON(res)
}

func (h *WaitingHandler) Status(c fiber.Ctx) error {
	eventID := c.Params("eventID")
	if eventID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "event id is required")
	}
	uid, err := userID(c)
	if err != nil {
		return err
	}
	res, err := h.svc.Status(c.Context(), eventID, uid)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not read status")
	}
	return c.JSON(res)
}

type rateRequest struct {
	Rate float64 `json:"rate"`
}

func (h *WaitingHandler) SetRate(c fiber.Ctx) error {
	eventID := c.Params("eventID")
	if eventID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "event id is required")
	}
	var req rateRequest
	if err := c.Bind().Body(&req); err != nil {
		if raw := c.Query("rate"); raw != "" {
			if v, perr := strconv.ParseFloat(raw, 64); perr == nil {
				req.Rate = v
			}
		}
	}
	if req.Rate <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "rate must be positive")
	}
	if err := h.svc.SetRate(c.Context(), eventID, req.Rate); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not set rate")
	}
	return c.SendStatus(fiber.StatusNoContent)
}
