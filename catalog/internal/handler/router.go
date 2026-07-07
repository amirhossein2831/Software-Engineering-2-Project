package handler

import (
	"github.com/gofiber/fiber/v3"

	"catalog/internal/service"
)

func NewRouter(app *fiber.App, svc *service.CatalogService) {
	app.Get("/healthz", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	h := NewCatalogHandler(svc)

	app.Get("/events", h.ListEvents)
	app.Get("/events/:id", h.GetEvent)
	app.Post("/events", h.CreateEvent)
	app.Post("/events/:id/publish", h.PublishEvent)
	app.Post("/events/:id/pricing", h.SetPricing)

	app.Post("/venues", h.CreateVenue)
	app.Get("/venues/:id", h.GetVenue)
	app.Post("/venues/:id/sectors", h.AddSector)
}
