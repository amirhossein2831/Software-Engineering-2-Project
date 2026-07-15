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
	app.Patch("/events/:id", h.UpdateEvent)
	app.Delete("/events/:id", h.DeleteEvent)
	app.Post("/events/:id/publish", h.PublishEvent)
	app.Post("/events/:id/pricing", h.SetPricing)

	app.Get("/venues", h.ListVenues)
	app.Post("/venues", h.CreateVenue)
	app.Get("/venues/:id", h.GetVenue)
	app.Patch("/venues/:id", h.UpdateVenue)
	app.Delete("/venues/:id", h.DeleteVenue)
	app.Post("/venues/:id/sectors", h.AddSector)
	app.Delete("/venues/:id/sectors/:sectorId", h.DeleteSector)
}
