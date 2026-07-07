package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"ticketing/internal/repository"
	"ticketing/internal/service"
)

type TicketHandler struct {
	svc *service.TicketingService
}

func NewTicketHandler(svc *service.TicketingService) *TicketHandler {
	return &TicketHandler{svc: svc}
}

func (h *TicketHandler) GetTicket(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	ticket, err := h.svc.GetTicket(c.Context(), id)
	if errors.Is(err, repository.ErrNotFound) {
		return fiber.NewError(fiber.StatusNotFound, "ticket not found")
	}
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "error")
	}
	return c.JSON(ticket)
}

func (h *TicketHandler) ListTickets(c fiber.Ctx) error {
	raw := c.Query("order")
	if raw == "" {
		return fiber.NewError(fiber.StatusBadRequest, "order query param is required")
	}
	orderID, err := uuid.Parse(raw)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid order id")
	}
	tickets, err := h.svc.ListByOrder(c.Context(), orderID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "error")
	}
	return c.JSON(fiber.Map{"tickets": tickets})
}
