package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"my-go-app/internal/auth"
	"my-go-app/internal/model"
	"my-go-app/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailTaken         = errors.New("email already registered")
	ErrValidation         = errors.New("validation error")
)

type AuthService struct {
	users  *repository.UserRepository
	tokens *auth.TokenManager
}

func NewAuthService(users *repository.UserRepository, tokens *auth.TokenManager) *AuthService {
	return &AuthService{users: users, tokens: tokens}
}

type AuthResult struct {
	User      *model.User
	Token     string
	ExpiresAt time.Time
}

// RegisterInput holds all fields required to create an account.
type RegisterInput struct {
	Name         string
	Email        string
	Password     string
	Gender       string
	MotherName   string
	FatherName   string
	MobileNumber string
	ChildAge     int
	ChildClass   string
}

func (s *AuthService) Register(ctx context.Context, in RegisterInput) (*AuthResult, error) {
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))

	hash, err := auth.HashPassword(in.Password)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Name:         in.Name,
		Email:        in.Email,
		Gender:       in.Gender,
		MotherName:   in.MotherName,
		FatherName:   in.FatherName,
		MobileNumber: in.MobileNumber,
		ChildAge:     in.ChildAge,
		ChildClass:   in.ChildClass,
		Password:     hash,
	}

	if err := s.users.Create(ctx, user); err != nil {
		if isUniqueViolation(err) {
			return nil, ErrEmailTaken
		}
		return nil, err
	}

	token, expiresAt, err := s.tokens.Issue(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	return &AuthResult{User: user, Token: token, ExpiresAt: expiresAt}, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*AuthResult, error) {
	email = strings.TrimSpace(strings.ToLower(email))

	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if !auth.CheckPassword(user.Password, password) {
		return nil, ErrInvalidCredentials
	}

	token, expiresAt, err := s.tokens.Issue(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	return &AuthResult{User: user, Token: token, ExpiresAt: expiresAt}, nil
}

func (s *AuthService) Me(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	return user, nil
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	// GORM wraps postgres unique_violation (23505).
	msg := strings.ToLower(err.Error())
	return errors.Is(err, gorm.ErrDuplicatedKey) ||
		strings.Contains(msg, "duplicate key") ||
		strings.Contains(msg, "unique constraint")
}
