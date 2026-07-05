package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"ticketing/internal/iam/auth"
	"ticketing/internal/iam/model"
	"ticketing/pkg/jwtauth"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidRefresh     = errors.New("invalid refresh token")
)

type UserStore interface {
	Create(ctx context.Context, u *model.User) error
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	SaveRefresh(ctx context.Context, t *model.RefreshToken) error
	FindRefreshByHash(ctx context.Context, hash string) (*model.RefreshToken, error)
	RevokeRefresh(ctx context.Context, id uuid.UUID) error
}

type AuthService struct {
	store UserStore
	jwt   *jwtauth.Manager
}

func NewAuthService(store UserStore, jwt *jwtauth.Manager) *AuthService {
	return &AuthService{store: store, jwt: jwt}
}

func (s *AuthService) Register(ctx context.Context, email, password string) (*model.User, error) {
	hash, err := auth.HashPassword(password)
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
	return u, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, string, error) {
	u, err := s.store.FindByEmail(ctx, email)
	if err != nil {
		return "", "", ErrInvalidCredentials
	}
	if !auth.CheckPassword(u.PasswordHash, password) {
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
