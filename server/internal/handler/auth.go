package handler

import (
	"errors"
	"net/http"

	appctx "github.com/daylamtayari/cierge/server/internal/context"
	"github.com/daylamtayari/cierge/server/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type Auth struct {
	authService   *service.AuthService
	isDevelopment bool
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func NewAuth(authService *service.AuthService, isDevelopment bool) *Auth {
	return &Auth{
		authService:   authService,
		isDevelopment: isDevelopment,
	}
}

// POST /auth/login
func (h *Auth) Login(c *gin.Context) {
	errorCol := appctx.ErrorCollector(c.Request.Context())
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorCol.Add(err, zerolog.InfoLevel, true, nil, "improper login request format")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":      "Invalid request",
			"request_id": appctx.RequestID(c.Request.Context()),
		})
		return
	}

	tokenSet, err := h.authService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserDNE):
			errorCol.Add(err, zerolog.InfoLevel, true, nil, "user does not exist")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":      "Invalid credentials",
				"request_id": appctx.RequestID(c.Request.Context()),
			})
			return
		case errors.Is(err, service.ErrInvalidCredentials):
			errorCol.Add(err, zerolog.InfoLevel, true, nil, "invalid user credentials")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":      "Invalid credentials",
				"request_id": appctx.RequestID(c.Request.Context()),
			})
			return
		case errors.Is(err, service.ErrAccountLocked):
			errorCol.Add(err, zerolog.InfoLevel, true, nil, "authentication attempted for a locked account")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":      "Too many requests",
				"request_id": appctx.RequestID(c.Request.Context()),
			})
			return
		default:
			errorCol.Add(err, zerolog.ErrorLevel, false, nil, "internal server error during the login flow")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error":      "Internal server error",
				"request_id": appctx.RequestID(c.Request.Context()),
			})
			return
		}
	}

	c.SetSameSite(http.SameSiteLaxMode)
	for _, cookie := range tokenSet {
		c.SetCookie(
			cookie.Name,
			cookie.Value,
			int(cookie.MaxAge.Seconds()),
			"/",
			"",
			h.isDevelopment,
			true,
		)
	}

	c.JSON(200, gin.H{"message": "login successful"})
	c.Set("message", "successful login")
}
