package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/daylamtayari/cierge/server/internal/config"
	"github.com/daylamtayari/cierge/server/internal/util"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	AccessTokenCookieName  = "access_token"
	RefreshTokenCookieName = "refresh_token"

	ErrAccountLocked       = errors.New("account is temporarily locked")
	ErrFailFetchUser       = errors.New("failed to fetch user")
	ErrFailRecordFailLogin = errors.New("failed to record a failed login")
	ErrInvalidCredentials  = errors.New("invalid credential")
)

type AuthCookie struct {
	Name   string
	Value  string
	MaxAge time.Duration
}

type AuthService struct {
	userService  *UserService
	tokenService *TokenService
	authConfig   *config.AuthConfig
	argonParams  *util.Argon2Params
}

func NewAuthService(userService *UserService, tokenService *TokenService, authConfig *config.AuthConfig) *AuthService {
	return &AuthService{
		userService:  userService,
		tokenService: tokenService,
		authConfig:   authConfig,
		argonParams: &util.Argon2Params{
			Memory:      64 * 1024,
			Iterations:  3,
			Parallelism: 4,
			SaltLength:  16,
			KeyLength:   32,
		},
	}
}

// Authenticates a user with email and password and returns a slice of AuthCookie
func (s *AuthService) Login(ctx context.Context, email string, password string) ([]AuthCookie, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	user, err := s.userService.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserDNE
		}
		return nil, fmt.Errorf("%w: %w", ErrFailFetchUser, err)
	}

	if user.IsAccountLocked() {
		return nil, ErrAccountLocked
	}

	match, err := util.SecureVerifyHash(*user.PasswordHash, password)
	if err != nil {
		return nil, err
	}
	if !match {
		if user.FailedLoginAttempts+1 >= s.authConfig.RateLimitRequests && time.Now().UTC().Sub(*user.LastFailedLogin) <= s.authConfig.RateLimitWindow.Duration() {
			lockUntil := time.Now().UTC().Add(s.authConfig.RateLimitWindow.Duration())
			err = s.userService.RecordFailedLogin(ctx, user.ID, &lockUntil)
		} else {
			err = s.userService.RecordFailedLogin(ctx, user.ID, nil)
		}
		if err != nil {
			return nil, ErrFailRecordFailLogin
		}
		return nil, ErrInvalidCredentials
	}

	tokenSet, err := s.generateTokenSet(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	err = s.userService.RecordSuccessfulLogin(ctx, user.ID)
	return tokenSet, err
}

// Hashes a given password and returns the argon2id hash
func (s *AuthService) HashPassword(password string) string {
	return util.HashSaltString(password, s.argonParams)
}

// Generate a set of AuthCookie's containing tokens
func (s *AuthService) generateTokenSet(ctx context.Context, userID uuid.UUID) ([]AuthCookie, error) {
	authCookies := make([]AuthCookie, 0)
	// Access token
	accessToken, err := s.tokenService.GenerateAccessToken(ctx, userID)
	if err != nil {
		return nil, err
	}
	authCookies = append(authCookies, AuthCookie{
		Name:   "access_token",
		Value:  accessToken,
		MaxAge: s.authConfig.AccessTokenExpiry.Duration(),
	})

	// Refresh token
	refrehToken, err := s.tokenService.GenerateRefreshToken(ctx, userID)
	if err != nil {
		return nil, err
	}
	authCookies = append(authCookies, AuthCookie{
		Name:   "refresh_token",
		Value:  refrehToken,
		MaxAge: s.authConfig.RefreshTokenExpiry.Duration(),
	})

	return authCookies, nil
}
