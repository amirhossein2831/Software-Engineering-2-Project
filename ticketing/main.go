package main

import (
	"context"

	"github.com/gofiber/fiber/v3"

	"ticketing/internal/config"
	"ticketing/internal/database"
	"ticketing/internal/events"
	"ticketing/internal/handler"
	"ticketing/internal/logger"
	"ticketing/internal/model"
	"ticketing/internal/qr"
	"ticketing/internal/repository"
	"ticketing/internal/service"
)

func main() {
	log := logger.New("ticketing")

	db := database.MustOpen(config.MustGet("DATABASE_URL"))
	if err := db.AutoMigrate(&model.Ticket{}); err != nil {
		log.Error("migration failed", "err", err)
		panic(err)
	}

	signer := qr.NewSigner(config.Get("TICKET_SIGNING_KEY", "dev-signing-key"))
	pub := events.NewPublisher(config.Get("KAFKA_BROKERS", ""))
	defer pub.Close()

	repo := repository.NewTicketRepo(db)
	svc := service.NewTicketingService(repo, signer, pub)

	consumer := events.NewConsumer(
		config.Get("KAFKA_BROKERS", ""),
		config.Get("CHECKOUT_TOPIC", "checkout.events"),
		config.Get("CONSUMER_GROUP", "ticketing"),
	)
	go func() {
		if err := consumer.Run(context.Background(), func(ctx context.Context, key, value []byte) error {
			return svc.Handle(ctx, value)
		}); err != nil {
			log.Error("consumer stopped", "err", err)
		}
	}()

	app := fiber.New()
	handler.NewRouter(app, svc)

	addr := ":" + config.Get("TICKETING_PORT", "8085")
	log.Info("ticketing listening", "addr", addr)
	if err := app.Listen(addr); err != nil {
		log.Error("server stopped", "err", err)
		panic(err)
	}
}
