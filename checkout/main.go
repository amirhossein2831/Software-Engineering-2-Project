package main

import (
	"time"

	"github.com/gofiber/fiber/v3"

	"checkout/internal/clients"
	"checkout/internal/config"
	"checkout/internal/database"
	"checkout/internal/events"
	"checkout/internal/handler"
	"checkout/internal/logger"
	"checkout/internal/model"
	"checkout/internal/payment"
	"checkout/internal/repository"
	"checkout/internal/service"
)

func main() {
	log := logger.New("checkout")

	db := database.MustOpen(config.MustGet("DATABASE_URL"))
	if err := db.AutoMigrate(&model.Order{}, &model.Payment{}, &model.SagaLog{}); err != nil {
		log.Error("migration failed", "err", err)
		panic(err)
	}

	timeout := time.Duration(config.GetInt("UPSTREAM_TIMEOUT_SECONDS", 5)) * time.Second
	reservation := clients.NewReservationClient(config.Get("RESERVATION_URL", "http://localhost:8083"), timeout)
	catalog := clients.NewCatalogClient(config.Get("CATALOG_URL", "http://localhost:8082"), timeout)

	gateway := payment.NewMockGateway(
		config.GetFloat("PAYMENT_FAIL_RATE", 0),
		time.Duration(config.GetInt("PAYMENT_LATENCY_MS", 0))*time.Millisecond,
	)
	pub := events.NewPublisher(config.Get("KAFKA_BROKERS", ""))
	defer pub.Close()

	repo := repository.NewOrderRepo(db)
	svc := service.NewCheckoutService(repo, reservation, catalog, gateway, pub)

	app := fiber.New()
	handler.NewRouter(app, svc)

	addr := ":" + config.Get("CHECKOUT_PORT", "8084")
	log.Info("checkout listening", "addr", addr)
	if err := app.Listen(addr); err != nil {
		log.Error("server stopped", "err", err)
		panic(err)
	}
}
