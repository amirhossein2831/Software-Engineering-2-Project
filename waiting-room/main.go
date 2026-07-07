package main

import (
	"time"

	"github.com/gofiber/fiber/v3"

	"waitingroom/internal/config"
	"waitingroom/internal/handler"
	"waitingroom/internal/logger"
	"waitingroom/internal/queue"
	"waitingroom/internal/service"
	"waitingroom/internal/token"
)

func main() {
	log := logger.New("waiting-room")

	rate := config.GetInt("ADMIT_RATE_PER_SEC", 5)
	q := queue.New(config.Get("REDIS_ADDR", "localhost:6379"), float64(rate))

	signer := token.NewSigner(config.Get("ADMISSION_SECRET", "dev-admission-secret"))
	admissionTTL := time.Duration(config.GetInt("ADMISSION_TTL_SECONDS", 300)) * time.Second

	svc := service.NewWaitingService(q, signer, admissionTTL)

	app := fiber.New()
	handler.NewRouter(app, svc)

	addr := ":" + config.Get("WAITING_ROOM_PORT", "8088")
	log.Info("waiting-room listening", "addr", addr)
	if err := app.Listen(addr); err != nil {
		log.Error("server stopped", "err", err)
		panic(err)
	}
}
