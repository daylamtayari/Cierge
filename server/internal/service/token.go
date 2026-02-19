package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"
	"unicode"

	"github.com/daylamtayari/cierge/server/internal/config"
	appctx "github.com/daylamtayari/cierge/server/internal/context"
	"github.com/daylamtayari/cierge/server/internal/model"
	"github.com/daylamtayari/cierge/server/internal/repository"
	tokenstore "github.com/daylamtayari/cierge/server/internal/token_store"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type TokenType string

const (
	BearerToken   TokenType = "Bearer"
	ApiToken      TokenType = "api"
	CallbackToken TokenType = "callback"
)

var (
	ErrApiKeyCheckFail       = errors.New("failed to check API keys")
	ErrApiKeyGenerationFail  = errors.New("failed to generate API key")
	ErrExpiredToken          = errors.New("expired token")
	ErrInvalidHeaderFormat   = errors.New("invalid authorization header format")
	ErrInvalidIssuer         = errors.New("invalid issuer")
	ErrInvalidSigningMethod  = errors.New("invalid signing method")
	ErrInvalidToken          = errors.New("invalid token")
	ErrInvalidTokenSignature = errors.New("invalid token signature")
	ErrInvalidTokenType      = errors.New("invalid token type")
	ErrRevocationCheckFail   = errors.New("revocation check failed")
	ErrRevokedToken          = errors.New("revoked token")
	ErrRevokedTokenNoAt      = errors.New("no revoked_at value in revoked token")
	ErrRevokedTokenNoBy      = errors.New("no revoked_by value in revoked token")
	ErrSignatureFail         = errors.New("failed to sign token")
	ErrTokenStoreFail        = errors.New("failed to store token in the token store")
	ErrUnknownApiKey         = errors.New("unknown API key")
)

type TokenRevocationError struct {
	Err       error
	JTI       string
	UserID    uuid.UUID
	RevokedBy string
	RevokedAt time.Time
}

func (e *TokenRevocationError) Error() string {
	return fmt.Sprintf("%v jti: %s user_id: %v revoked_at: %v revoked_by: %s", e.Err, e.JTI, e.UserID, e.RevokedAt, e.RevokedBy)
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

type Token struct {
	userService        *User
	jobRepo            *repository.Job
	tokenStore         *tokenstore.Store
	jwtSecret          string
	jwtIssuer          string
	accessTokenExpiry  time.Duration
	refreshTokenExpiry time.Duration
}

func NewToken(userService *User, jobRepo *repository.Job, authConfig config.Auth, tokenStore *tokenstore.Store) *Token {
	return &Token{
		userService:        userService,
		jobRepo:            jobRepo,
		tokenStore:         tokenStore,
		jwtSecret:          authConfig.JWTSecret,
		jwtIssuer:          authConfig.JWTIssuer,
		accessTokenExpiry:  authConfig.AccessTokenExpiry.Duration(),
		refreshTokenExpiry: authConfig.RefreshTokenExpiry.Duration(),
	}
}

// Identifies the token type and extracts the token from a given auth header
func (s *Token) ExtractToken(ctx context.Context, authHeader string) (TokenType, string, error) {
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
		return TokenType(parts[0]), "", ErrInvalidTokenType
	}
}

// Creates a SHA-256 hash of a secret
func (s *Token) hashSecret(apiKey string) string {
	hash := sha256.Sum256([]byte(apiKey))
	return hex.EncodeToString(hash[:])
}

// Validates an API key token and returns the corresponding user
func (s *Token) ValidateApiToken(ctx context.Context, apiToken string) (*model.User, error) {
	// Perform a light validation check that the token is only alphanumerics
	// Length should already be equal to 30 as checked by the ExtractToken method
	for _, r := range apiToken {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return nil, ErrInvalidToken
		}
	}

	hashedToken := s.hashSecret(apiToken)

	user, err := s.userService.GetByApiKey(ctx, hashedToken)
	if err != nil && errors.Is(err, ErrUserDNE) {
		return nil, ErrUnknownApiKey
	} else if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrApiKeyCheckFail, err)
	}
	return user, nil
}

// Validates whether a callback secret matches the specified hash
func (s *Token) ValidateCallbackSecret(ctx context.Context, callbackSecretHash string, callbackSecret string) bool {
	return subtle.ConstantTimeCompare([]byte(callbackSecretHash), []byte(s.hashSecret(callbackSecret))) == 1
}

// Validates a bearer token wrapping the ValidateJWTToken method
func (s *Token) ValidateBearerToken(ctx context.Context, bearerToken string) (*AccessTokenClaims, error) {
	claims, err := s.validateJWTToken(ctx, bearerToken)
	if err != nil {
		return nil, err
	}
	return &AccessTokenClaims{RegisteredClaims: *claims}, err
}

// Validates a refresh token wrapping the ValidateJWTToken method
func (s *Token) ValidateRefreshToken(ctx context.Context, refreshToken string) (*RefreshTokenClaims, error) {
	claims, err := s.validateJWTToken(ctx, refreshToken)
	if err != nil {
		return nil, err
	}
	return &RefreshTokenClaims{RegisteredClaims: *claims}, err
}

// Validates a JWT token and returns the corresponding access token claims
func (s *Token) validateJWTToken(ctx context.Context, jwtToken string) (*jwt.RegisteredClaims, error) {
	token, err := jwt.ParseWithClaims(jwtToken, &jwt.RegisteredClaims{}, func(token *jwt.Token) (any, error) {
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

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	if claims.Issuer != s.jwtIssuer {
		return nil, fmt.Errorf("%w: %v", ErrInvalidIssuer, s.jwtIssuer)
	}

	// Check if the token has been revoked
	revocation, err := s.tokenStore.GetToken(ctx, claims.ID)
	if err != nil && !errors.Is(err, tokenstore.ErrTokenNotFound) {
		return nil, fmt.Errorf("%w: %w", ErrRevocationCheckFail, err)
	} else if errors.Is(err, tokenstore.ErrTokenNotFound) {
		return nil, err
	} else if revocation.Revoked {
		if revocation.RevokedAt == nil {
			return nil, ErrRevokedTokenNoAt
		}
		if revocation.RevokedBy == nil {
			return nil, ErrRevokedTokenNoBy
		}

		revokedTokenErr := TokenRevocationError{
			Err:       err,
			JTI:       claims.ID,
			UserID:    revocation.UserID,
			RevokedBy: *revocation.RevokedBy,
			RevokedAt: *revocation.RevokedAt,
		}
		return nil, &revokedTokenErr
	}

	return claims, nil
}

// Generates an access token for a given user ID and returns the token and an optional error
func (s *Token) GenerateAccessToken(ctx context.Context, userID uuid.UUID) (string, error) {
	return s.generateJWTToken(ctx, userID, s.accessTokenExpiry)
}

// Generates a refresh token for a given user ID and returns the token and an optional error
func (s *Token) GenerateRefreshToken(ctx context.Context, userID uuid.UUID) (string, error) {
	return s.generateJWTToken(ctx, userID, s.refreshTokenExpiry)
}

// Generates a JWT for a given user ID and with a given expiry and returns the token and an optional error
// Also stores the JTI in the token store
func (s *Token) generateJWTToken(ctx context.Context, userID uuid.UUID, expiry time.Duration) (string, error) {
	jti := uuid.New().String()
	now := time.Now().UTC()

	claims := jwt.RegisteredClaims{
		ID:        jti,
		Subject:   userID.String(),
		Issuer:    s.jwtIssuer,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", ErrSignatureFail
	}

	err = s.tokenStore.StoreToken(ctx, jti, userID, expiry)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrTokenStoreFail, err)
	}

	// Adding logging of the JTI
	appctx.Logger(ctx).UpdateContext(func(c zerolog.Context) zerolog.Context {
		return c.Str("jti", jti)
	})

	return tokenString, nil
}

// Securely generate a random alphanumeric string
func (s *Token) generateSecretKey() (string, error) {
	const (
		secretKeyLength = 30
		charset         = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	)

	charsetLen := big.NewInt(int64(len(charset)))
	// Generate random alphanumeric string
	key := make([]byte, secretKeyLength)
	for j := range key {
		num, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return "", fmt.Errorf("%w: %w", ErrApiKeyGenerationFail, err)
		}
		key[j] = charset[num.Int64()]
	}

	return string(key), nil
}

// GenerateAPIKey creates or replaces a user's API key
// Returns the plaintext API key and an error that is nil if successful
func (s *Token) GenerateAPIKey(ctx context.Context, userID uuid.UUID) (string, error) {
	const maxRetries = 10
	// Generate a unique API key with collision checking
	var apiKey string
	var hashedKey string
	var err error
	for i := range maxRetries {
		apiKey, err = s.generateSecretKey()
		if err != nil {
			return "", err
		}
		hashedKey = s.hashSecret(apiKey)
		// Check if the key already exists
		exists, err := s.userService.ExistsByApiKey(ctx, hashedKey)
		if err != nil {
			return "", fmt.Errorf("%w: failed to check key uniqueness: %w", ErrApiKeyGenerationFail, err)
		}
		if !exists {
			break
		}

		// If all retries are exhausted
		if i == maxRetries-1 {
			return "", fmt.Errorf("%w: failed to generate unique key after %d attempts", ErrApiKeyGenerationFail, maxRetries)
		}
	}

	err = s.userService.userRepo.UpdateAPIKey(ctx, userID, hashedKey)
	if err != nil {
		return "", fmt.Errorf("%w: failed to update user API key: %w", ErrApiKeyGenerationFail, err)
	}

	return apiKey, nil
}

// Generate a callback secret and sets it for the specified job
// Returns the plaintext secret and an error that is nil if successful
func (s *Token) GenerateCallbackSecret(ctx context.Context, jobID uuid.UUID) (string, error) {
	secret, err := s.generateSecretKey()
	if err != nil {
		return "", err
	}

	err = s.jobRepo.SetCallbackSecretHash(ctx, s.hashSecret(secret), jobID)
	if err != nil {
		return "", err
	}
	return secret, nil
}

// Revokes a token for a given JTI
func (s *Token) RevokeToken(ctx context.Context, jti string, revokedBy string) error {
	data, err := s.tokenStore.GetToken(ctx, jti)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	data.Revoked = true
	data.RevokedBy = &revokedBy
	data.RevokedAt = &now

	return s.tokenStore.UpdateToken(ctx, jti, data)
}
