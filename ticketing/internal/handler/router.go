package handler

import (
	"github.com/gofiber/fiber/v3"

	"ticketing/internal/service"
)

func NewRouter(app *fiber.App, svc *service.TicketingService) {
	app.Get("/healthz", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	h := NewTicketHandler(svc)

	app.Get("/tickets", h.ListTickets)
	app.Get("/tickets/:id", h.GetTicket)
}
