package repository

import (
	"context"
	"errors"
	"fmt"

	"my-go-app/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrNotFound = errors.New("not found")

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find by email: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find by id: %w", err)
	}
	return &user, nil
}
