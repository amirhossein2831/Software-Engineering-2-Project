package main

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3"

	"auth/internal/config"
	"auth/internal/database"
	"auth/internal/events"
	"auth/internal/handler"
	"auth/internal/jwtauth"
	"auth/internal/logger"
	"auth/internal/model"
	"auth/internal/repository"
	"auth/internal/seed"
	"auth/internal/service"
)

func main() {
	log := logger.New("auth")

	db := database.MustOpen(config.MustGet("DATABASE_URL"))
	if err := db.AutoMigrate(&model.User{}, &model.RefreshToken{}); err != nil {
		log.Error("migration failed", "err", err)
		panic(err)
	}

	accessTTL := parseDuration(config.Get("ACCESS_TOKEN_TTL", "15m"), 15*time.Minute)
	refreshTTL := parseDuration(config.Get("REFRESH_TOKEN_TTL", "720h"), 720*time.Hour)

	jwtMgr := jwtauth.NewManager(config.MustGet("JWT_SECRET"), accessTTL, refreshTTL)
	pub := events.NewPublisher(config.Get("KAFKA_BROKERS", ""))
	defer pub.Close()

	repo := repository.NewUserRepo(db)

	if res, err := seed.EnsureAdmins(context.Background(), repo, seed.Admins); err != nil {
		log.Error("admin seed failed", "err", err)
	} else if res.Created+res.Promoted > 0 {
		log.Info("admin seed applied", "created", res.Created, "promoted", res.Promoted)
	}

	svc := service.NewAuthService(repo, jwtMgr, pub)

	app := fiber.New()
	handler.NewRouter(app, svc)

	addr := ":" + config.Get("AUTH_PORT", "8081")
	log.Info("auth listening", "addr", addr)
	if err := app.Listen(addr); err != nil {
		log.Error("server stopped", "err", err)
		panic(err)
	}
}

func parseDuration(v string, fallback time.Duration) time.Duration {
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}
