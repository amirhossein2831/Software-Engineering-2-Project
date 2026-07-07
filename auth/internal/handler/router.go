package handler

import (
	"github.com/gofiber/fiber/v3"

	"auth/internal/service"
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

	a := app.Group("/admin")
	a.Get("/users", h.ListUsers)
	a.Post("/users/:id/role", h.SetRole)
}
