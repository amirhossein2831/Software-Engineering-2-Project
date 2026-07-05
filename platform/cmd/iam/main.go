package main

import (
	"time"

	"github.com/gofiber/fiber/v3"

	"ticketing/internal/iam/handler"
	"ticketing/internal/iam/model"
	"ticketing/internal/iam/repository"
	"ticketing/internal/iam/service"
	"ticketing/pkg/config"
	"ticketing/pkg/database"
	"ticketing/pkg/jwtauth"
	"ticketing/pkg/logger"
)

func main() {
	log := logger.New("iam")

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

	addr := ":" + config.Get("IAM_PORT", "8081")
	log.Info("iam listening", "addr", addr)
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
