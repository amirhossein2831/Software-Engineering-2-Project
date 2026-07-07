package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"auth/internal/jwtauth"
	"auth/internal/model"
	"auth/internal/security"
)

const authTopic = "auth.events"

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidRefresh     = errors.New("invalid refresh token")
	ErrInvalidRole        = errors.New("invalid role")
)

type UserStore interface {
	Create(ctx context.Context, u *model.User) error
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	ListUsers(ctx context.Context) ([]model.User, error)
	UpdateRole(ctx context.Context, id uuid.UUID, role model.Role) error
	SaveRefresh(ctx context.Context, t *model.RefreshToken) error
	FindRefreshByHash(ctx context.Context, hash string) (*model.RefreshToken, error)
	RevokeRefresh(ctx context.Context, id uuid.UUID) error
}

type EventPublisher interface {
	Publish(ctx context.Context, topic, key string, payload any) error
}

type UserRegistered struct {
	Type   string    `json:"type"`
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
}

type AuthService struct {
	store UserStore
	jwt   *jwtauth.Manager
	pub   EventPublisher
}

func NewAuthService(store UserStore, jwt *jwtauth.Manager, pub EventPublisher) *AuthService {
	return &AuthService{store: store, jwt: jwt, pub: pub}
}

func (s *AuthService) Register(ctx context.Context, email, password string) (*model.User, error) {
	hash, err := security.HashPassword(password)
	if err != nil {
		return nil, err
	}
	u := &model.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: hash,
		Role:         model.RoleBuyer,
	}
	if err := s.store.Create(ctx, u); err != nil {
		return nil, err
	}
	_ = s.pub.Publish(ctx, authTopic, u.ID.String(), UserRegistered{
		Type:   "user.registered",
		UserID: u.ID,
		Email:  u.Email,
	})
	return u, nil
}

func (s *AuthService) ListUsers(ctx context.Context) ([]model.User, error) {
	return s.store.ListUsers(ctx)
}

func (s *AuthService) SetRole(ctx context.Context, id uuid.UUID, role model.Role) (*model.User, error) {
	if !role.Valid() {
		return nil, ErrInvalidRole
	}
	if err := s.store.UpdateRole(ctx, id, role); err != nil {
		return nil, err
	}
	return s.store.FindByID(ctx, id)
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, string, error) {
	u, err := s.store.FindByEmail(ctx, email)
	if err != nil {
		return "", "", ErrInvalidCredentials
	}
	if !security.CheckPassword(u.PasswordHash, password) {
		return "", "", ErrInvalidCredentials
	}
	return s.issueTokens(ctx, u)
}

func (s *AuthService) Refresh(ctx context.Context, rawRefresh string) (string, string, error) {
	hash := jwtauth.HashToken(rawRefresh)
	tok, err := s.store.FindRefreshByHash(ctx, hash)
	if err != nil {
		return "", "", ErrInvalidRefresh
	}
	if !tok.Active(time.Now()) {
		return "", "", ErrInvalidRefresh
	}
	if err := s.store.RevokeRefresh(ctx, tok.ID); err != nil {
		return "", "", err
	}
	u, err := s.store.FindByID(ctx, tok.UserID)
	if err != nil {
		return "", "", ErrInvalidRefresh
	}
	return s.issueTokens(ctx, u)
}

func (s *AuthService) Logout(ctx context.Context, rawRefresh string) error {
	hash := jwtauth.HashToken(rawRefresh)
	tok, err := s.store.FindRefreshByHash(ctx, hash)
	if err != nil {
		return nil
	}
	return s.store.RevokeRefresh(ctx, tok.ID)
}

func (s *AuthService) issueTokens(ctx context.Context, u *model.User) (string, string, error) {
	access, err := s.jwt.IssueAccess(u.ID.String(), string(u.Role))
	if err != nil {
		return "", "", err
	}
	raw, hash := s.jwt.NewRefreshToken()
	rt := &model.RefreshToken{
		ID:        uuid.New(),
		UserID:    u.ID,
		TokenHash: hash,
		ExpiresAt: time.Now().Add(s.jwt.RefreshTTL()),
	}
	if err := s.store.SaveRefresh(ctx, rt); err != nil {
		return "", "", err
	}
	return access, raw, nil
}
