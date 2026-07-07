package main

import (
	"context"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"

	"notification/internal/config"
	"notification/internal/database"
	"notification/internal/events"
	"notification/internal/logger"
	"notification/internal/model"
	"notification/internal/repository"
	"notification/internal/sender"
	"notification/internal/service"
	"notification/internal/worker"
)

func main() {
	log := logger.New("notification")

	db := database.MustOpen(config.MustGet("DATABASE_URL"))
	if err := db.AutoMigrate(&model.NotificationOutbox{}, &model.Recipient{}); err != nil {
		log.Error("migration failed", "err", err)
		panic(err)
	}

	repo := repository.NewNotificationRepo(db)
	svc := service.NewNotificationService(repo)

	registry := sender.NewRegistry(
		sender.NewEmailSender(config.Get("SMTP_ADDR", "localhost:1025"), config.Get("SMTP_FROM", "no-reply@tickets.local")),
		sender.NewSmsSender(log),
	)

	w := worker.New(
		repo, registry, log,
		config.GetInt("MAX_ATTEMPTS", 5),
		config.GetInt("WORKER_BATCH", 20),
		time.Duration(config.GetInt("WORKER_INTERVAL_SECONDS", 5))*time.Second,
	)

	topics := strings.Split(config.Get("CONSUME_TOPICS", "checkout.events,ticketing.events,auth.events"), ",")
	consumer := events.NewConsumer(config.Get("KAFKA_BROKERS", ""), topics, config.Get("CONSUMER_GROUP", "notification"))

	ctx := context.Background()
	go w.Run(ctx)
	go func() {
		if err := consumer.Run(ctx, func(ctx context.Context, key, value []byte) error {
			return svc.Handle(ctx, value)
		}); err != nil {
			log.Error("consumer stopped", "err", err)
		}
	}()

	app := fiber.New()
	app.Get("/healthz", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	addr := ":" + config.Get("NOTIFICATION_PORT", "8086")
	log.Info("notification listening", "addr", addr)
	if err := app.Listen(addr); err != nil {
		log.Error("server stopped", "err", err)
		panic(err)
	}
}
