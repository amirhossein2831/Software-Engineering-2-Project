package main

import (
	"github.com/gofiber/fiber/v3"

	"catalog/internal/config"
	"catalog/internal/database"
	"catalog/internal/handler"
	"catalog/internal/logger"
	"catalog/internal/model"
	"catalog/internal/repository"
	"catalog/internal/service"
)

func main() {
	log := logger.New("catalog")

	db := database.MustOpen(config.MustGet("DATABASE_URL"))
	if err := db.AutoMigrate(&model.Venue{}, &model.Sector{}, &model.Event{}, &model.Pricing{}, &model.Seat{}); err != nil {
		log.Error("migration failed", "err", err)
		panic(err)
	}

	repo := repository.NewCatalogRepo(db)
	svc := service.NewCatalogService(repo)

	app := fiber.New()
	handler.NewRouter(app, svc)

	addr := ":" + config.Get("CATALOG_PORT", "8082")
	log.Info("catalog listening", "addr", addr)
	if err := app.Listen(addr); err != nil {
		log.Error("server stopped", "err", err)
		panic(err)
	}
}
