package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/daylamtayari/cierge/internal/config"
	"github.com/daylamtayari/cierge/internal/model"
	"github.com/daylamtayari/cierge/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type TokenType string

const (
	BearerToken TokenType = "Bearer"
	ApiToken    TokenType = "api"
)

var (
	ErrApiKeyCheckFail       = errors.New("failed to check API keys")
	ErrExpiredToken          = errors.New("expired token")
	ErrInvalidHeaderFormat   = errors.New("invalid authorization header format")
	ErrInvalidIssuer         = errors.New("invalid issuer")
	ErrInvalidSigningMethod  = errors.New("invalid signing method")
	ErrInvalidToken          = errors.New("invalid token")
	ErrInvalidTokenSignature = errors.New("invalid token signature")
	ErrInvalidTokenType      = errors.New("invalid token type")
	ErrRevocationCheckFail   = errors.New("revocation check failed")
	ErrRevokedToken          = errors.New("revoked token")
	ErrUnknownApiKey         = errors.New("unknown API key")
)

type TokenRevocationError struct {
	Err error
	model.Revocation
}

func (e *TokenRevocationError) Error() string {
	return fmt.Sprintf("%w revocation_id: %s user_id: %s revoked_at: %v", e.Err, e.ID, e.UserID, e.RevokedAt)
}

func (e *TokenRevocationError) Unwrap() error {
	return e.Err
}

type AccessTokenClaims struct {
	jwt.RegisteredClaims
}

type RefreshTokenClaims struct {
	jwt.RegisteredClaims
}

type TokenService struct {
	userRepo           *repository.UserRepository
	revocationRepo     *repository.RevocationRepository
	jwtSecret          string
	jwtIssuer          string
	accessTokenExpiry  time.Duration
	refreshTokenExpiry time.Duration
}

func NewTokenService(userRepo *repository.UserRepository, authConfig config.AuthConfig) *TokenService {
	return &TokenService{
		userRepo:           userRepo,
		jwtSecret:          authConfig.JWTSecret,
		jwtIssuer:          authConfig.JWTIssuer,
		accessTokenExpiry:  authConfig.AccessTokenExpiry.Duration(),
		refreshTokenExpiry: authConfig.RefreshTokenExpiry.Duration(),
	}
}

// Identifies the token type and extracts the token from a given auth header
func (s *TokenService) ExtractToken(ctx context.Context, authHeader string) (TokenType, string, error) {
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 {
		return TokenType("invalid"), "", ErrInvalidHeaderFormat
	}

	switch parts[0] {
	case string(BearerToken):
		if strings.Count(parts[1], ".") == 2 {
			return BearerToken, parts[1], nil
		}
		return BearerToken, "", ErrInvalidToken
	case string(ApiToken):
		if len(parts[1]) == 30 {
			return ApiToken, parts[1], nil
		}
		return ApiToken, "", ErrInvalidToken
	default:
		return TokenType("invalid"), "", ErrInvalidTokenType
	}
}

// Validates an API key token and returns the corresponding user
func (s *TokenService) ValidateApiToken(ctx context.Context, apiToken string) (*model.User, error) {
	// Perform a light validation check that the token is only alphanumerics
	// Length should already be equal to 30 as checked by the ExtractToken method
	for _, r := range apiToken {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return nil, ErrInvalidToken
		}
	}
	user, err := s.userRepo.GetByApiKey(ctx, apiToken)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("%w: %w", ErrApiKeyCheckFail, err)
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUnknownApiKey
	}
	return user, nil
}

// Validates a bearer token and returns the corresponding access token claims
func (s *TokenService) ValidateBearerToken(ctx context.Context, bearerToken string) (*AccessTokenClaims, error) {
	token, err := jwt.ParseWithClaims(bearerToken, &AccessTokenClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%w: %v", ErrInvalidSigningMethod, token.Header["alg"])
		} else if token.Method.Alg() != jwt.SigningMethodHS512.Alg() {
			return nil, fmt.Errorf("%w: %v", ErrInvalidSigningMethod, token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		switch {
		case errors.Is(err, jwt.ErrTokenExpired):
			return nil, ErrExpiredToken
		case errors.Is(err, jwt.ErrTokenSignatureInvalid):
			return nil, ErrInvalidTokenSignature
		default:
			return nil, fmt.Errorf("%w: %w", ErrInvalidToken, err)
		}
	}

	claims, ok := token.Claims.(*AccessTokenClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	if claims.Issuer != s.jwtIssuer {
		return nil, fmt.Errorf("%w: %v", ErrInvalidIssuer, s.jwtIssuer)
	}

	// Check if the token has been revoked
	revocation, err := s.revocationRepo.GetByJTI(ctx, claims.ID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("%w: %w", ErrRevocationCheckFail, err)
	} else if revocation != nil {
		revokedTokenErr := TokenRevocationError{
			Err:        err,
			Revocation: *revocation,
		}
		return nil, &revokedTokenErr
	}

	return claims, nil
}
