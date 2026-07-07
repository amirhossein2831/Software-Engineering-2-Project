package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"checkout/internal/repository"
	"checkout/internal/service"
)

type CheckoutHandler struct {
	svc *service.CheckoutService
}

func NewCheckoutHandler(svc *service.CheckoutService) *CheckoutHandler {
	return &CheckoutHandler{svc: svc}
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

type checkoutRequest struct {
	HoldID string `json:"hold_id"`
}

func (h *CheckoutHandler) Checkout(c fiber.Ctx) error {
	uid, err := userID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "missing user")
	}
	var req checkoutRequest
	if err := c.Bind().Body(&req); err != nil || req.HoldID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "hold_id is required")
	}
	holdID, err := uuid.Parse(req.HoldID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid hold_id")
	}
	order, err := h.svc.Checkout(c.Context(), uid, holdID, c.Get("Idempotency-Key"))
	if errors.Is(err, service.ErrHoldInvalid) {
		return fiber.NewError(fiber.StatusConflict, "hold is not valid for checkout")
	}
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not process checkout")
	}
	return c.Status(fiber.StatusCreated).JSON(order)
}

func (h *CheckoutHandler) GetOrder(c fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	detail, err := h.svc.GetOrder(c.Context(), id)
	if errors.Is(err, repository.ErrNotFound) {
		return fiber.NewError(fiber.StatusNotFound, "order not found")
	}
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "error")
	}
	return c.JSON(detail)
}

func (h *CheckoutHandler) ListOrders(c fiber.Ctx) error {
	uid, err := userID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "missing user")
	}
	orders, err := h.svc.ListOrders(c.Context(), uid)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "error")
	}
	return c.JSON(fiber.Map{"orders": orders})
}
