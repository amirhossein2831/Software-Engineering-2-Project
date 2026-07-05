package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"auth/internal/model"
)

var (
	ErrEmailTaken = errors.New("email already registered")
	ErrNotFound   = errors.New("not found")
)

type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, u *model.User) error {
	err := r.db.WithContext(ctx).Create(u).Error
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return ErrEmailTaken
	}
	return err
}

func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var u model.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var u model.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) SaveRefresh(ctx context.Context, t *model.RefreshToken) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *UserRepo) FindRefreshByHash(ctx context.Context, hash string) (*model.RefreshToken, error) {
	var t model.RefreshToken
	err := r.db.WithContext(ctx).Where("token_hash = ?", hash).First(&t).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *UserRepo) RevokeRefresh(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&model.RefreshToken{}).
		Where("id = ?", id).
		Update("revoked_at", gorm.Expr("NOW()")).Error
}
