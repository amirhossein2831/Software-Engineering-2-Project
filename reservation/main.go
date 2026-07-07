package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"

	"reservation/internal/config"
	"reservation/internal/database"
	"reservation/internal/events"
	"reservation/internal/handler"
	"reservation/internal/logger"
	"reservation/internal/model"
	"reservation/internal/redislock"
	"reservation/internal/repository"
	"reservation/internal/service"
)

func main() {
	log := logger.New("reservation")

	db := database.MustOpen(config.MustGet("DATABASE_URL"))
	if err := db.AutoMigrate(&model.SeatState{}, &model.Reservation{}); err != nil {
		log.Error("migration failed", "err", err)
		panic(err)
	}

	ttl := time.Duration(config.GetInt("HOLD_TTL_SECONDS", 300)) * time.Second
	locker := redislock.New(config.Get("REDIS_ADDR", "localhost:6379"), ttl)
	pub := events.NewPublisher(config.Get("KAFKA_BROKERS", ""))
	defer pub.Close()

	repo := repository.NewReservationRepo(db)
	svc := service.NewReservationService(repo, locker, pub)

	go runSweeper(svc, log, time.Duration(config.GetInt("SWEEP_INTERVAL_SECONDS", 30))*time.Second)

	app := fiber.New()
	handler.NewRouter(app, svc)

	addr := ":" + config.Get("RESERVATION_PORT", "8083")
	log.Info("reservation listening", "addr", addr)
	if err := app.Listen(addr); err != nil {
		log.Error("server stopped", "err", err)
		panic(err)
	}
}

func runSweeper(svc *service.ReservationService, log *slog.Logger, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		n, err := svc.SweepExpired(context.Background())
		if err != nil {
			log.Error("sweep failed", "err", err)
			continue
		}
		if n > 0 {
			log.Info("released expired holds", "count", n)
		}
	}
}
