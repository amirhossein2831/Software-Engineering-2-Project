package main

import (
	"time"

	"github.com/gofiber/fiber/v3"

	"auth/internal/config"
	"auth/internal/database"
	"auth/internal/handler"
	"auth/internal/jwtauth"
	"auth/internal/logger"
	"auth/internal/model"
	"auth/internal/repository"
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
	repo := repository.NewUserRepo(db)
	svc := service.NewAuthService(repo, jwtMgr)

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
