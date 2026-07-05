package handler

import (
	"github.com/gofiber/fiber/v3"

	"ticketing/internal/iam/service"
)

func NewRouter(app *fiber.App, svc *service.AuthService) {
	app.Get("/healthz", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	h := NewAuthHandler(svc)
	g := app.Group("/auth")
	g.Post("/register", h.Register)
	g.Post("/login", h.Login)
	g.Post("/refresh", h.Refresh)
	g.Post("/logout", h.Logout)
}
