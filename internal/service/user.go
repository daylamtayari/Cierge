package service

import (
	"context"
	"errors"
	"time"

	"github.com/daylamtayari/cierge/internal/model"
	"github.com/daylamtayari/cierge/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrUserDNE = errors.New("user does not exist")
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// Get user from a given UUID
func (s *UserService) GetByID(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserDNE
	} else if err != nil {
		return nil, err
	}
	return user, nil
}

// Get a user from a given UUID
func (s *UserService) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserDNE
	} else if err != nil {
		return nil, err
	}
	return user, nil
}

// Gets a user from a given API key
func (s *UserService) GetByApiKey(ctx context.Context, apiKey string) (*model.User, error) {
	user, err := s.userRepo.GetByApiKey(ctx, apiKey)
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserDNE
	} else if err != nil {
		return nil, err
	}
	return user, nil
}

// Records a successful login
func (s *UserService) RecordSuccessfulLogin(ctx context.Context, userId uuid.UUID) error {
	return s.userRepo.RecordSuccessfulLogin(ctx, userId)
}

func (s *UserService) RecordFailedLogin(ctx context.Context, userID uuid.UUID, lockUntil *time.Time) error {
	return s.userRepo.RecordFailedLogin(ctx, userID, lockUntil)
}
