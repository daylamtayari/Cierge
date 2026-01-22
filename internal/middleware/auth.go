package middleware

import (
	"errors"
	"fmt"
	"net/http"

	appctx "github.com/daylamtayari/cierge/internal/context"
	"github.com/daylamtayari/cierge/internal/model"
	"github.com/daylamtayari/cierge/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
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
		errorCol := appctx.ErrorCollector(c.Request.Context())

		var tokenType service.TokenType
		var tokenString string

		cookieAccessToken, err := c.Cookie(service.AccessTokenCookieName)
		if err == nil && cookieAccessToken != "" {
			tokenType = service.BearerToken
			tokenString = cookieAccessToken
		} else {
			// Fall back to Auth header - for API usage
			authHeader := c.GetHeader("Authorization")
			tokenType, tokenString, err = m.tokenService.ExtractToken(c.Request.Context(), authHeader)
			if err != nil {
				if errors.Is(err, service.ErrInvalidTokenType) {
					errorCol.Add(err, zerolog.InfoLevel, true, map[string]any{"input_token_type": string(tokenType)}, "failed authentication attempt due to an invalid token type")
				} else {
					errorCol.Add(err, zerolog.InfoLevel, true, nil, fmt.Sprintf("failed authentication attempt due to an invalid %v token", tokenType))
				}
				respondUnauthorized(c)
				return
			}
		}

		var user *model.User

		logger.UpdateContext(func(c zerolog.Context) zerolog.Context {
			return c.Str("auth_method", string(tokenType))
		})

		if tokenType == service.ApiToken {
			validatedUser, err := m.tokenService.ValidateApiToken(c.Request.Context(), tokenString)
			if err != nil {
				if errors.Is(err, service.ErrApiKeyCheckFail) {
					errorCol.Add(err, zerolog.ErrorLevel, false, nil, "failed authentication attempt due to an error checking the API key")
					respondInternalServerError(c)
					return
				} else {
					errorCol.Add(err, zerolog.InfoLevel, true, nil, "failed authentication attempt due to an incorrect API key")
					respondUnauthorized(c)
					return
				}
			}
			user = validatedUser
		} else {
			claims, err := m.tokenService.ValidateBearerToken(c.Request.Context(), tokenString)

			if err != nil {
				var revocationError *service.TokenRevocationError
				switch {
				case errors.Is(err, service.ErrInvalidIssuer):
					errorCol.Add(err, zerolog.WarnLevel, true, nil, "failed authentication attempt due to a valid bearer token but from an invalid issuer")
				case errors.Is(err, service.ErrInvalidTokenSignature):
					errorCol.Add(err, zerolog.WarnLevel, true, nil, "failed authentication attempt due to an invalid token signature")
				case errors.As(err, &revocationError):
					errorCol.Add(err, zerolog.WarnLevel, true, map[string]any{
						"user_id":     revocationError.UserID.String(),
						"revoked_at":  revocationError.RevokedAt,
						"revoked_by":  revocationError.RevokedBy,
						"revoked_jti": revocationError.JTI,
					}, "failed authentication attempt due to attempted usage of a revoked bearer token")
				default:
					errorCol.Add(err, zerolog.InfoLevel, true, nil, "failed authentication attempt with a bearer token")
				}
				respondUnauthorized(c)
				return
			}

			userID, err := uuid.Parse(claims.Subject)
			if err != nil {
				errorCol.Add(err, zerolog.ErrorLevel, false, map[string]any{"jwt_subject": claims.Subject}, "failed to parse the bearer token subject into a UUID")
				// Return an internal server error if the user ID value failed to parse
				// due to either not actually being a UUID which is very problematic
				// or due to an issue with the parsing, also problematic
				respondInternalServerError(c)
				return
			}

			retrievedUser, err := m.userService.GetByID(c.Request.Context(), userID)
			if err != nil {
				if errors.Is(err, service.ErrUserDNE) {
					errorCol.Add(err, zerolog.WarnLevel, true, map[string]any{"user_id": userID.String()}, "user with valid bearer token no longer exists")
					respondUnauthorized(c)
					return
				}
				errorCol.Add(err, zerolog.ErrorLevel, false, map[string]any{"user_id": userID.String()}, "failed to retrieve user during auth flow")
				respondInternalServerError(c)
				return
			}

			user = retrievedUser
		}

		c.Set("user", user)
		c.Set("is_admin", user.IsAdmin)
		ctx := appctx.WithUserID(c.Request.Context(), user.ID)

		logger.UpdateContext(func(c zerolog.Context) zerolog.Context {
			return c.
				Str("user_id", user.ID.String()).
				Str("auth_method", string(tokenType)).
				Bool("is_admin", user.IsAdmin)
		})

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// Ensures that an authenticated user that is required to change their
// password on next login, is not allowed to proceed
// NOTE: Must be chained after RequireAuth()
func (m *AuthMiddleware) RequirePasswordChange() gin.HandlerFunc {
	return func(c *gin.Context) {
		errorCol := appctx.ErrorCollector(c.Request.Context())

		gUser, ok := c.Get("user")
		var user *model.User
		if !ok {
			errorCol.Add(nil, zerolog.ErrorLevel, false, nil, "user object not found in gin context when expected")
			respondForbidden(c)
			return
		} else if user, ok = gUser.(*model.User); !ok {
			errorCol.Add(nil, zerolog.ErrorLevel, false, nil, "user value from gin context failed to cast to a User type")
			respondForbidden(c)
			return
		}

		if user.NeedsPasswordChange() {
			errorCol.Add(nil, zerolog.InfoLevel, true, nil, "user requires password change before proceeding")
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":      "Forbidden",
				"request_id": appctx.RequestID(c.Request.Context()),
				"message":    "Password change is required",
			})
			return
		}

		c.Next()
	}
}

// Ensures that the authenticated user has administrator privileges
// NOTE: Must be chained after RequireAuth()
func (m *AuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := appctx.Logger(c.Request.Context())
		errorCol := appctx.ErrorCollector(c.Request.Context())

		isAdmin := c.GetBool("is_admin")
		if !isAdmin {
			errorCol.Add(nil, zerolog.WarnLevel, true, nil, "user attempted to access an administrative endpoint")
			respondForbidden(c)
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
