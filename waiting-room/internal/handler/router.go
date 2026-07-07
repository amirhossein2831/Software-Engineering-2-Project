package handler

import (
	"github.com/gofiber/fiber/v3"

	"waitingroom/internal/service"
)

func NewRouter(app *fiber.App, svc *service.WaitingService) {
	app.Get("/healthz", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	h := NewWaitingHandler(svc)

	app.Post("/queue/:eventID/join", h.Join)
	app.Get("/queue/:eventID/status", h.Status)
	app.Put("/queue/:eventID/rate", h.SetRate)
}
