package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/daylamtayari/cierge/server/internal/config"
	"github.com/daylamtayari/cierge/server/internal/util"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var defaultArgonParams = &util.Argon2Params{
	Memory:      64 * 1024,
	Iterations:  3,
	Parallelism: 4,
	SaltLength:  16,
	KeyLength:   32,
}

var (
	AccessTokenCookieName  = "access_token"
	RefreshTokenCookieName = "refresh_token"

	ErrAccountLocked       = errors.New("account is temporarily locked")
	ErrFailFetchUser       = errors.New("failed to fetch user")
	ErrFailRecordFailLogin = errors.New("failed to record a failed login")
	ErrInvalidCredentials  = errors.New("invalid credential")
)

type PasswordValidationError string

func (e PasswordValidationError) Error() string { return string(e) }

const (
	ErrPasswordTooShort       PasswordValidationError = "password must be at least 8 characters"
	ErrPasswordTooLong        PasswordValidationError = "password must be at most 128 characters"
	ErrPasswordMissingLetter  PasswordValidationError = "password must contain at least one letter"
	ErrPasswordMissingDigit   PasswordValidationError = "password must contain at least one digit"
	ErrPasswordMissingSpecial PasswordValidationError = "password must contain at least one special character"
)

type AuthCookie struct {
	Name   string
	Value  string
	MaxAge time.Duration
}

type Auth struct {
	userService  *User
	tokenService *Token
	authConfig   *config.Auth
	argonParams  *util.Argon2Params
}

func NewAuth(userService *User, tokenService *Token, authConfig *config.Auth) *Auth {
	return &Auth{
		userService:  userService,
		tokenService: tokenService,
		authConfig:   authConfig,
		argonParams:  defaultArgonParams,
	}
}

// Generates a cryptographically random password satisfying the complexity requirements:
// 16 characters, with at least one letter, one digit, and one special character
func generateRandomPassword() (string, error) {
	const (
		letters  = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
		digits   = "0123456789"
		specials = "!@#$%^&*-_=+?"
		allChars = letters + digits + specials
		length   = 16
	)

	password := make([]byte, length)

	// Guarantee one character from each required class
	for i, charset := range []string{letters, digits, specials} {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		password[i] = charset[idx.Int64()]
	}

	// Fill remaining positions with random characters
	for i := 3; i < length; i++ {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(allChars))))
		if err != nil {
			return "", err
		}
		password[i] = allChars[idx.Int64()]
	}

	// Shuffle to avoid predictable positions for required characters
	for i := length - 1; i > 0; i-- {
		j, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return "", err
		}
		password[i], password[j.Int64()] = password[j.Int64()], password[i]
	}

	return string(password), nil
}

// Authenticates a user with email and password and returns a slice of AuthCookie
func (s *Auth) Login(ctx context.Context, email string, password string) ([]AuthCookie, error) {
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

// Changes a user's password to the provided new password
// NOTE: Assumes that the old password has been verified prior to calling this
func (s *Auth) ChangePassword(ctx context.Context, newPassword string, userId uuid.UUID) error {
	if err := s.validatePassword(newPassword); err != nil {
		return err
	}
	return s.userService.UpdatePassword(ctx, userId, s.HashPassword(newPassword))
}

// Hashes a given password and returns the argon2id hash
func (s *Auth) HashPassword(password string) string {
	return util.HashSaltString(password, s.argonParams)
}

// Validates that a password meets the complexity requirements:
// 8–128 characters, at least one letter, one digit, and one special character
func (s *Auth) validatePassword(password string) error {
	length := utf8.RuneCountInString(password)
	if length < 8 {
		return ErrPasswordTooShort
	}
	if length > 128 {
		return ErrPasswordTooLong
	}

	var hasLetter, hasDigit, hasSpecial bool
	for _, r := range password {
		switch {
		case unicode.IsLetter(r):
			hasLetter = true
		case unicode.IsDigit(r):
			hasDigit = true
		default:
			hasSpecial = true
		}
	}

	if !hasLetter {
		return ErrPasswordMissingLetter
	}
	if !hasDigit {
		return ErrPasswordMissingDigit
	}
	if !hasSpecial {
		return ErrPasswordMissingSpecial
	}
	return nil
}

// Performs a logout by validating the tokens and if valid, revoking them
func (s *Auth) Logout(ctx context.Context, accessToken string, refreshToken string) error {
	if accessToken != "" {
		accessTokenClaims, err := s.tokenService.ValidateAuthToken(ctx, accessToken)
		if err != nil {
			return err
		}
		err = s.tokenService.RevokeToken(ctx, accessTokenClaims.ID, "logout")
		if err != nil {
			return err
		}
	}

	if refreshToken != "" {
		refreshTokenClaims, err := s.tokenService.ValidateRefreshToken(ctx, refreshToken)
		if err != nil {
			return err
		}
		err = s.tokenService.RevokeToken(ctx, refreshTokenClaims.ID, "logout")
		if err != nil {
			return err
		}
	}

	return nil
}

// Refresh validates a refresh token, revokes it, and issues a new token set
func (s *Auth) Refresh(ctx context.Context, refreshToken string) ([]AuthCookie, error) {
	claims, err := s.tokenService.ValidateRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, err
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	user, err := s.userService.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user.IsAccountLocked() {
		return nil, ErrAccountLocked
	}

	tokenSet, err := s.generateTokenSet(ctx, userID)
	if err != nil {
		return nil, err
	}

	if err := s.tokenService.RevokeToken(ctx, claims.ID, "refresh"); err != nil {
		return nil, err
	}

	return tokenSet, nil
}

// Generate a set of AuthCookie's containing tokens
func (s *Auth) generateTokenSet(ctx context.Context, userID uuid.UUID) ([]AuthCookie, error) {
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
