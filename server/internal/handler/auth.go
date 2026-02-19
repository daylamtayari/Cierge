package handler

import (
	"errors"
	"net/http"

	appctx "github.com/daylamtayari/cierge/server/internal/context"
	"github.com/daylamtayari/cierge/server/internal/service"
	tokenstore "github.com/daylamtayari/cierge/server/internal/token_store"
	"github.com/daylamtayari/cierge/server/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type Auth struct {
	authService   *service.Auth
	isDevelopment bool
}

type loginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func NewAuth(authService *service.Auth, isDevelopment bool) *Auth {
	return &Auth{
		authService:   authService,
		isDevelopment: isDevelopment,
	}
}

// POST /auth/login
func (h *Auth) Login(c *gin.Context) {
	errorCol := appctx.ErrorCollector(c.Request.Context())
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorCol.Add(err, zerolog.InfoLevel, true, nil, "improper login request format")
		util.RespondBadRequest(c, "")
		return
	}

	tokenSet, err := h.authService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserDNE):
			errorCol.Add(err, zerolog.InfoLevel, true, nil, "user does not exist")
			util.RespondUnauthorized(c)
			return
		case errors.Is(err, service.ErrInvalidCredentials):
			errorCol.Add(err, zerolog.InfoLevel, true, nil, "invalid user credentials")
			util.RespondUnauthorized(c)
			return
		case errors.Is(err, service.ErrAccountLocked):
			errorCol.Add(err, zerolog.InfoLevel, true, nil, "authentication attempted for a locked account")
			util.RespondTooMany(c)
			return
		default:
			errorCol.Add(err, zerolog.ErrorLevel, false, nil, "internal server error during the login flow")
			util.RespondInternalServerError(c)
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
			!h.isDevelopment,
			true,
		)
	}

	c.JSON(200, gin.H{"message": "login successful"})
	c.Set("message", "successful login")
}

// POST /auth/logout
func (h *Auth) Logout(c *gin.Context) {
	errorCol := appctx.ErrorCollector(c.Request.Context())

	accessToken, err := c.Cookie(service.AccessTokenCookieName)
	if err != nil && !errors.Is(err, http.ErrNoCookie) {
		errorCol.Add(err, zerolog.ErrorLevel, false, nil, "failed to retrieve access token from cookies")
		util.RespondInternalServerError(c)
		return
	}
	refreshToken, err := c.Cookie(service.RefreshTokenCookieName)
	if err != nil && !errors.Is(err, http.ErrNoCookie) {
		errorCol.Add(err, zerolog.ErrorLevel, false, nil, "failed to retrieve refresh token from cookies")
		util.RespondInternalServerError(c)
		return
	}

	err = h.authService.Logout(c.Request.Context(), accessToken, refreshToken)
	if err != nil {
		if errors.Is(err, tokenstore.ErrFailedToOpenTokenStore) {
			errorCol.Add(err, zerolog.ErrorLevel, false, nil, "failed to logout due token store error")
			util.RespondInternalServerError(c)
			return
		} else {
			errorCol.Add(err, zerolog.InfoLevel, true, nil, "failed logout due to invalid token")
			util.RespondUnauthorized(c)
			return
		}
	}

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(service.AccessTokenCookieName, "", -1, "/", "", !h.isDevelopment, true)
	c.SetCookie(service.RefreshTokenCookieName, "", -1, "/", "", !h.isDevelopment, true)

	c.JSON(200, gin.H{"message": "logout successful"})
	c.Set("message", "successful logout")
}

// POST /auth/refresh
func (h *Auth) Refresh(c *gin.Context) {
	errorCol := appctx.ErrorCollector(c.Request.Context())

	refreshToken, err := c.Cookie(service.RefreshTokenCookieName)
	if err != nil {
		errorCol.Add(err, zerolog.InfoLevel, true, nil, "no refresh token cookie")
		util.RespondUnauthorized(c)
		return
	}

	tokenSet, err := h.authService.Refresh(c.Request.Context(), refreshToken)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserDNE):
			errorCol.Add(err, zerolog.InfoLevel, true, nil, "user no longer exists during refresh")
			util.RespondUnauthorized(c)
		case errors.Is(err, service.ErrAccountLocked):
			errorCol.Add(err, zerolog.InfoLevel, true, nil, "refresh attempted for locked account")
			util.RespondTooMany(c)
		default:
			errorCol.Add(err, zerolog.InfoLevel, true, nil, "invalid or expired refresh token")
			util.RespondUnauthorized(c)
		}
		return
	}

	c.SetSameSite(http.SameSiteLaxMode)
	for _, cookie := range tokenSet {
		c.SetCookie(
			cookie.Name,
			cookie.Value,
			int(cookie.MaxAge.Seconds()),
			"/",
			"",
			!h.isDevelopment,
			true,
		)
	}

	c.JSON(200, gin.H{"message": "tokens refreshed"})
	c.Set("message", "successful token refresh")
}
