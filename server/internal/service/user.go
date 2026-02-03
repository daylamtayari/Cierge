package service

import (
	"context"
	"errors"
	"net/mail"
	"strings"
	"time"

	"github.com/daylamtayari/cierge/server/internal/model"
	"github.com/daylamtayari/cierge/server/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrUserDNE      = errors.New("user does not exist")
	ErrInvalidEmail = errors.New("invalid email address")
)

type UserService struct {
	userRepo *repository.User
}

func NewUserService(userRepo *repository.User) *UserService {
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

// Retrieves all users
func (s *UserService) GetUsers(ctx context.Context) ([]model.User, error) {
	return s.userRepo.List(ctx)
}

// Retrieves user count
func (s *UserService) GetUserCount(ctx context.Context) (int, error) {
	users, err := s.GetUsers(ctx)
	if err != nil {
		return 0, err
	}
	return len(users), nil
}

// Records a successful login
func (s *UserService) RecordSuccessfulLogin(ctx context.Context, userId uuid.UUID) error {
	return s.userRepo.RecordSuccessfulLogin(ctx, userId)
}

// Records a failed login
func (s *UserService) RecordFailedLogin(ctx context.Context, userID uuid.UUID, lockUntil *time.Time) error {
	return s.userRepo.RecordFailedLogin(ctx, userID, lockUntil)
}

// Create a user from their email (that is then validated), password hash, and isAdmin boolean value
// Returns a user object pointer and an error which is nil if successful
func (s *UserService) Create(ctx context.Context, email string, hashedPassword string, isAdmin bool) (*model.User, error) {
	// Email validation
	if len(email) > 254 {
		return nil, ErrInvalidEmail
	}
	parsedEmail, err := mail.ParseAddress(strings.ToLower(email))
	if err != nil {
		return nil, ErrInvalidEmail
	}
	emailParts := strings.Split(parsedEmail.Address, "@")
	if len(emailParts) != 2 {
		return nil, ErrInvalidEmail
	}
	if len(emailParts[0]) > 64 || len(emailParts[1]) > 255 {
		return nil, ErrInvalidEmail
	}

	user := model.User{
		Email:        parsedEmail.Address,
		PasswordHash: &hashedPassword,
		IsAdmin:      isAdmin,
	}

	err = s.userRepo.Create(ctx, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
