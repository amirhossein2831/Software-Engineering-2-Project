package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"

	"auth/internal/repository"
	"auth/internal/service"
)

type AuthHandler struct {
	svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

type credentialsRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (h *AuthHandler) Register(c fiber.Ctx) error {
	var req credentialsRequest
	if err := c.Bind().Body(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	if req.Email == "" || req.Password == "" {
		return fiber.NewError(fiber.StatusBadRequest, "email and password are required")
	}
	u, err := h.svc.Register(c.Context(), req.Email, req.Password)
	if errors.Is(err, repository.ErrEmailTaken) {
		return fiber.NewError(fiber.StatusConflict, "email already registered")
	}
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not register")
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":    u.ID,
		"email": u.Email,
		"role":  u.Role,
	})
}

func (h *AuthHandler) Login(c fiber.Ctx) error {
	var req credentialsRequest
	if err := c.Bind().Body(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	access, refresh, err := h.svc.Login(c.Context(), req.Email, req.Password)
	if errors.Is(err, service.ErrInvalidCredentials) {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid credentials")
	}
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "login failed")
	}
	return c.JSON(tokenResponse{AccessToken: access, RefreshToken: refresh})
}

func (h *AuthHandler) Refresh(c fiber.Ctx) error {
	var req refreshRequest
	if err := c.Bind().Body(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	access, refresh, err := h.svc.Refresh(c.Context(), req.RefreshToken)
	if errors.Is(err, service.ErrInvalidRefresh) {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid refresh token")
	}
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "refresh failed")
	}
	return c.JSON(tokenResponse{AccessToken: access, RefreshToken: refresh})
}

func (h *AuthHandler) Logout(c fiber.Ctx) error {
	var req refreshRequest
	if err := c.Bind().Body(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	if err := h.svc.Logout(c.Context(), req.RefreshToken); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "logout failed")
	}
	return c.SendStatus(fiber.StatusNoContent)
}
