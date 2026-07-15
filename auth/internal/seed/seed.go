package seed

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"auth/internal/model"
	"auth/internal/repository"
	"auth/internal/security"
)

type Admin struct {
	Email    string
	Password string
}

var Admins = []Admin{
	{Email: "admin@tickets.local", Password: "admin12345"},
}

type Result struct {
	Created  int
	Promoted int
}

type store interface {
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	Create(ctx context.Context, u *model.User) error
	UpdateRole(ctx context.Context, id uuid.UUID, role model.Role) error
}

func EnsureAdmins(ctx context.Context, s store, admins []Admin) (Result, error) {
	var res Result
	for _, a := range admins {
		if a.Email == "" || a.Password == "" {
			continue
		}

		existing, err := s.FindByEmail(ctx, a.Email)
		if err == nil {
			if existing.Role != model.RoleAdmin {
				if err := s.UpdateRole(ctx, existing.ID, model.RoleAdmin); err != nil {
					return res, err
				}
				res.Promoted++
			}
			continue
		}
		if !errors.Is(err, repository.ErrNotFound) {
			return res, err
		}

		hash, err := security.HashPassword(a.Password)
		if err != nil {
			return res, err
		}
		u := &model.User{
			ID:           uuid.New(),
			Email:        a.Email,
			PasswordHash: hash,
			Role:         model.RoleAdmin,
		}
		if err := s.Create(ctx, u); err != nil {
			return res, err
		}
		res.Created++
	}
	return res, nil
}
