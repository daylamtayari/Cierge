package service

import (
	"context"
	"errors"
	"net/mail"
	"strings"
	"time"

	"github.com/daylamtayari/cierge/server/internal/model"
	"github.com/daylamtayari/cierge/server/internal/repository"
	"github.com/daylamtayari/cierge/server/internal/util"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrUserDNE           = errors.New("user does not exist")
	ErrUserAlreadyExists = errors.New("user with that email already exists")
	ErrInvalidEmail      = errors.New("invalid email address")
)

type User struct {
	userRepo *repository.User
}

func NewUser(userRepo *repository.User) *User {
	return &User{
		userRepo: userRepo,
	}
}

// Get user from a given UUID
func (s *User) GetByID(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserDNE
	} else if err != nil {
		return nil, err
	}
	return user, nil
}

// Get a user from a given UUID
func (s *User) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserDNE
	} else if err != nil {
		return nil, err
	}
	return user, nil
}

// Gets a user from a given API key
func (s *User) GetByApiKey(ctx context.Context, apiKey string) (*model.User, error) {
	user, err := s.userRepo.GetByApiKey(ctx, apiKey)
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserDNE
	} else if err != nil {
		return nil, err
	}
	return user, nil
}

// Retrieves all users
func (s *User) GetUsers(ctx context.Context) ([]model.User, error) {
	return s.userRepo.List(ctx)
}

// Retrieves user count
func (s *User) GetUserCount(ctx context.Context) (int, error) {
	users, err := s.GetUsers(ctx)
	if err != nil {
		return 0, err
	}
	return len(users), nil
}

// Checks if an API key string exists
func (s *User) ExistsByApiKey(ctx context.Context, apiKey string) (bool, error) {
	return s.userRepo.ExistsByApiKey(ctx, apiKey)
}

// Records a successful login
func (s *User) RecordSuccessfulLogin(ctx context.Context, userId uuid.UUID) error {
	return s.userRepo.RecordSuccessfulLogin(ctx, userId)
}

// Records a failed login
func (s *User) RecordFailedLogin(ctx context.Context, userID uuid.UUID, lockUntil *time.Time) error {
	return s.userRepo.RecordFailedLogin(ctx, userID, lockUntil)
}

// Updates the password hash for a user
func (s *User) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	return s.userRepo.UpdatePassword(ctx, userID, passwordHash)
}

// Creates or updates a user's API key
func (s *User) UpdateAPIKey(ctx context.Context, id uuid.UUID, apiKey string) error {
	return s.userRepo.UpdateAPIKey(ctx, id, apiKey)
}

// Create a user from their email (that is then validated), password hash, and isAdmin boolean value
// Returns a user object pointer and an error which is nil if successful
func (s *User) Create(ctx context.Context, email string, hashedPassword string, isAdmin bool) (*model.User, error) {
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
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrUserAlreadyExists
		}
		return nil, err
	}
	return &user, nil
}

// Creates a new user with a randomly generated password that satisfies complexity requirements
// Returns the created user and the plaintext password
func (s *User) CreateWithGeneratedPassword(ctx context.Context, email string, isAdmin bool) (*model.User, string, error) {
	password, err := generateRandomPassword()
	if err != nil {
		return nil, "", err
	}

	hashedPassword := util.HashSaltString(password, defaultArgonParams)

	user, err := s.Create(ctx, email, hashedPassword, isAdmin)
	if err != nil {
		return nil, "", err
	}

	return user, password, nil
}
