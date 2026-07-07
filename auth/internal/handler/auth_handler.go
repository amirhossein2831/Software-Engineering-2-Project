package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"auth/internal/model"
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

func requireAdmin(c fiber.Ctx) error {
	if c.Get("X-User-Role") != string(model.RoleAdmin) {
		return fiber.NewError(fiber.StatusForbidden, "admin role required")
	}
	return nil
}

type userView struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

func toUserView(u model.User) userView {
	return userView{ID: u.ID.String(), Email: u.Email, Role: string(u.Role)}
}

func (h *AuthHandler) ListUsers(c fiber.Ctx) error {
	if err := requireAdmin(c); err != nil {
		return err
	}
	users, err := h.svc.ListUsers(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not list users")
	}
	views := make([]userView, 0, len(users))
	for _, u := range users {
		views = append(views, toUserView(u))
	}
	return c.JSON(fiber.Map{"users": views})
}

type setRoleRequest struct {
	Role string `json:"role"`
}

func (h *AuthHandler) SetRole(c fiber.Ctx) error {
	if err := requireAdmin(c); err != nil {
		return err
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	var req setRoleRequest
	if err := c.Bind().Body(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	u, err := h.svc.SetRole(c.Context(), id, model.Role(req.Role))
	if errors.Is(err, service.ErrInvalidRole) {
		return fiber.NewError(fiber.StatusBadRequest, "invalid role")
	}
	if errors.Is(err, repository.ErrNotFound) {
		return fiber.NewError(fiber.StatusNotFound, "user not found")
	}
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not update role")
	}
	return c.JSON(toUserView(*u))
}
