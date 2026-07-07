package handler

import (
	"github.com/gofiber/fiber/v3"

	"checkout/internal/service"
)

func NewRouter(app *fiber.App, svc *service.CheckoutService) {
	app.Get("/healthz", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	h := NewCheckoutHandler(svc)

	app.Post("/checkout", h.Checkout)
	app.Get("/orders", h.ListOrders)
	app.Get("/orders/:id", h.GetOrder)
}
