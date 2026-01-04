package service

import (
	"context"
	"errors"

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
