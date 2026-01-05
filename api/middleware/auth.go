package middleware

import (
	"errors"
	"net/http"

	appctx "github.com/daylamtayari/cierge/internal/context"
	"github.com/daylamtayari/cierge/internal/model"
	"github.com/daylamtayari/cierge/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthMiddleware struct {
	tokenService *service.TokenService
	userService  *service.UserService
}

func NewAuthMiddleware(tokenService *service.TokenService, userService *service.UserService) *AuthMiddleware {
	return &AuthMiddleware{
		tokenService: tokenService,
		userService:  userService,
	}
}

// Checks user authentication
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := appctx.Logger(c.Request.Context())

		authHeader := c.GetHeader("Authorization")

		tokenType, tokenString, err := m.tokenService.ExtractToken(c.Request.Context(), authHeader)
		if err != nil {
			if errors.Is(err, service.ErrInvalidTokenType) {
				logger.Info().Err(err).Str("input_token_type", string(tokenType)).Msg("failed authentication attempt due to an invalid token type")
			} else {
				logger.Info().Err(err).Msgf("failed authentication attempt due to an invalid %v token", tokenType)
			}
			m.respondUnauthorized(c)
			return
		}

		var user *model.User

		if tokenType == service.ApiToken {
			logger.Debug().Msg("handling API token")
			validatedUser, err := m.tokenService.ValidateApiToken(c.Request.Context(), tokenString)
			if err != nil {
				if errors.Is(err, service.ErrApiKeyCheckFail) {
					logger.Error().Err(err).Str("auth_method", string(tokenType)).Msg("failed authentication attempt due to an error checking the API key")
					m.respondInternalServerError(c)
					return
				} else {
					logger.Info().Err(err).Str("auth_method", string(tokenType)).Msg("failed authentication attempt due to an incorrect API key")
					m.respondUnauthorized(c)
					return
				}
			}
			user = validatedUser
		} else {
			logger.Debug().Msg("handling bearer token")
			claims, err := m.tokenService.ValidateBearerToken(c.Request.Context(), tokenString)

			if err != nil {
				var revocationError *service.TokenRevocationError
				switch {
				case errors.Is(err, service.ErrInvalidIssuer):
					logger.Warn().Err(err).Str("auth_method", string(tokenType)).Msg("valid bearer token was received but from an invalid issuer")
				case errors.Is(err, service.ErrInvalidTokenSignature):
					logger.Warn().Err(err).Str("auth_method", string(tokenType)).Msg("failed authentication attempt due to an invalid token signature")
				case errors.As(err, &revocationError):
					logger.Warn().Err(revocationError.Err).
						Str("auth_method", string(tokenType)).
						Str("user_id", revocationError.UserID.String()).
						Time("revoked_at", revocationError.RevokedAt).
						Str("revoked_by", revocationError.RevokedBy).
						Str("revoked_jti", revocationError.JTI).
						Msg("failed authentication attempt due to attempted usage of a revoked bearer token")
				default:
					logger.Info().Err(err).Str("auth_method", string(tokenType)).Msg("failed authentication attempt with a bearer token")
				}
				m.respondUnauthorized(c)
				return
			}

			userID, err := uuid.Parse(claims.Subject)
			if err != nil {
				logger.Error().Err(err).Str("auth_method", string(tokenType)).Str("jwt_subject", claims.Subject).Msg("failed to parse the bearer token subject into a UUID")
				// Return an internal server error if the user ID value failed to parse
				// due to either not actually being a UUID which is very problematic
				// or due to an issue with the parsing, also problematic
				m.respondInternalServerError(c)
				return
			}

			retrievedUser, err := m.userService.GetByID(c.Request.Context(), userID)
			if err != nil {
				if errors.Is(err, service.ErrUserDNE) {
					logger.Warn().Str("auth_method", string(tokenType)).Str("user_id", userID.String()).Msg("user with valid bearer token no longer exists")
					m.respondUnauthorized(c)
					return
				}
				logger.Error().Err(err).Str("auth_method", string(tokenType)).Str("user_id", userID.String()).Msg("failed to retrieve user")
				m.respondInternalServerError(c)
				return
			}

			user = retrievedUser
		}
		c.Set("user", user)
		c.Set("is_admin", user.IsAdmin)

		ctx := appctx.WithUserID(c.Request.Context(), user.ID)
		augLogger := logger.With().
			Str("user_id", user.ID.String()).
			Str("auth_method", string(tokenType)).
			Bool("is_admin", user.IsAdmin).
			Logger()
		ctx = appctx.WithLogger(ctx, &augLogger)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// Ensures that the authenticated user has administrator privileges
// NOTE: Must be chained after RequireAuth()
func (m *AuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := appctx.Logger(c.Request.Context())

		isAdmin := c.GetBool("is_admin")
		if !isAdmin {
			logger.Warn().Msg("user attempted to access an administrative endpoint")
			m.respondForbidden(c)
			return
		}

		// Add an admin_route field to the log entry for requests
		// made to a route restricted to admins
		augLogger := logger.With().
			Bool("admin_route", true).
			Logger()
		ctx := appctx.WithLogger(c.Request.Context(), &augLogger)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// Return an Unauthorized error response
func (m *AuthMiddleware) respondUnauthorized(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"error":      "Unauthorized",
		"request_id": appctx.RequestID(c.Request.Context()),
	})
}

// Return a Forbidden error response
func (m *AuthMiddleware) respondForbidden(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
		"error":      "Forbidden",
		"request_id": appctx.RequestID(c.Request.Context()),
	})
}

// Return an Internal Server Error response
func (m *AuthMiddleware) respondInternalServerError(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
		"error":      "Internal server error",
		"request_id": appctx.RequestID(c.Request.Context()),
	})
}
