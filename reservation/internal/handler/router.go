package handler

import (
	"github.com/gofiber/fiber/v3"

	"reservation/internal/service"
)

func NewRouter(app *fiber.App, svc *service.ReservationService) {
	app.Get("/healthz", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	h := NewReservationHandler(svc)

	app.Post("/holds", h.Hold)
	app.Get("/holds/:id", h.GetHold)
	app.Post("/holds/:id/commit", h.Commit)
	app.Post("/holds/:id/release", h.Release)
	app.Get("/seatmap/:eventID", h.SeatStates)
}
