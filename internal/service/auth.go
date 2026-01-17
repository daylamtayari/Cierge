package service

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/daylamtayari/cierge/internal/config"
	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"
	"gorm.io/gorm"
)

var (
	AccessTokenCookieName  = "access_token"
	RefreshTokenCookieName = "refresh_token"

	ErrAccountLocked         = errors.New("account is temporarily locked")
	ErrFailDecodeHash        = errors.New("failed to decode hash")
	ErrFailDecodeSalt        = errors.New("failed to decode salt")
	ErrFailFetchUser         = errors.New("failed to fetch user")
	ErrFailParseAttribute    = errors.New("failed to parse hash attribute")
	ErrFailRecordFailLogin   = errors.New("failed to record a failed login")
	ErrIncompatibleAlgorithm = errors.New("incompatible hashing algorithm")
	ErrIncompatibleVersion   = errors.New("incompatible argon2 version")
	ErrInvalidCredentials    = errors.New("invalid credential")
	ErrInvalidHashFormat     = errors.New("invalid password hash format")
)

type AuthCookie struct {
	Name   string
	Value  string
	MaxAge time.Duration
}

type argon2Params struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

type AuthService struct {
	userService  *UserService
	tokenService *TokenService
	authConfig   *config.AuthConfig
	argonParams  *argon2Params
}

func NewAuthService(userService *UserService, tokenService *TokenService, authConfig *config.AuthConfig) *AuthService {
	return &AuthService{
		userService:  userService,
		tokenService: tokenService,
		authConfig:   authConfig,
		argonParams: &argon2Params{
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

	match, err := s.verifyPassword(*user.PasswordHash, password)
	if err != nil {
		return nil, err
	}
	if !match {
		if user.FailedLoginAttempts+1 >= s.authConfig.RateLimitRequests && time.Now().UTC().Sub(*user.FirstFailedLogin) >= s.authConfig.RateLimitWindow.Duration() {
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
	return tokenSet, nil
}

// Checks if the provided password matches the stored encoded hash
func (s *AuthService) verifyPassword(encodedHash string, password string) (bool, error) {
	params, salt, hash, err := s.decodeHash(encodedHash)
	if err != nil {
		return false, err
	}

	compareHash := argon2.IDKey(
		[]byte(password),
		salt,
		params.Iterations,
		params.Memory,
		params.Parallelism,
		params.KeyLength,
	)

	if subtle.ConstantTimeCompare(hash, compareHash) == 1 {
		return true, nil
	}
	return false, nil
}

// Parses the encoded hash string and retrieves the params, salt, and hash
// Allows for the extraction of the parameters from the hash string
func (s *AuthService) decodeHash(encodedHash string) (*argon2Params, []byte, []byte, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return nil, nil, nil, ErrInvalidHashFormat
	}
	if parts[1] != "argon2id" {
		return nil, nil, nil, ErrIncompatibleAlgorithm
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return nil, nil, nil, ErrFailParseAttribute
	}
	if version != argon2.Version {
		return nil, nil, nil, ErrIncompatibleVersion
	}

	params := &argon2Params{}
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &params.Memory, &params.Iterations, &params.Parallelism); err != nil {
		return nil, nil, nil, ErrFailParseAttribute
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%w: %w", ErrFailDecodeSalt, err)
	}
	params.SaltLength = uint32(len(salt))
	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%w: %w", ErrFailDecodeHash, err)
	}
	params.KeyLength = uint32(len(hash))

	return params, salt, hash, nil
}

// Hashes a given password and returns the argon2id hash
func (s *AuthService) HashPassword(password string) string {
	salt := make([]byte, s.argonParams.SaltLength)
	rand.Read(salt) // nolint:errcheck
	// crypto/rand.Read "never" returns an error and if it fails, it crashes the program per the documentation

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		s.argonParams.Iterations,
		s.argonParams.Memory,
		s.argonParams.Parallelism,
		s.argonParams.KeyLength,
	)

	encodedSalt := base64.RawStdEncoding.EncodeToString(salt)
	encodedHash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		s.argonParams.Memory,
		s.argonParams.Iterations,
		s.argonParams.Parallelism,
		encodedSalt,
		encodedHash,
	)
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
