package handler

import (
	"errors"

	appctx "github.com/daylamtayari/cierge/server/internal/context"
	"github.com/daylamtayari/cierge/server/internal/model"
	"github.com/daylamtayari/cierge/server/internal/service"
	"github.com/daylamtayari/cierge/server/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type User struct {
	userService  *service.User
	tokenService *service.Token
	authService  *service.Auth
}

type createUserRequest struct {
	Email string `json:"email" binding:"required"`
}

type changePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

func NewUser(userService *service.User, tokenService *service.Token, authService *service.Auth) *User {
	return &User{
		userService:  userService,
		tokenService: tokenService,
		authService:  authService,
	}
}

func (h *User) Me(c *gin.Context) {
	errorCol := appctx.ErrorCollector(c.Request.Context())

	contextUser, ok := c.Get("user")
	if !ok {
		errorCol.Add(nil, zerolog.ErrorLevel, false, nil, "user object not found in gin context when expected")
		util.RespondInternalServerError(c)
		return
	}

	modelUser := contextUser.(*model.User)
	apiUser := modelUser.ToAPI()

	c.JSON(200, apiUser)
	c.Set("message", "retrieved own user")
}

// Generates a new API key for a user
// NOTE: This will replace the existing API key and as such
// will invalidate any past API keys. Highly recommend fetching
// the user and checking if they have an active API key first and
// if so, getting explicit confirmation.
func (h *User) APIKey(c *gin.Context) {
	errorCol := appctx.ErrorCollector(c.Request.Context())

	apiKey, err := h.tokenService.GenerateAPIKey(c.Request.Context(), appctx.UserID(c.Request.Context()))
	if err != nil {
		errorCol.Add(nil, zerolog.ErrorLevel, false, nil, "failed to generate API key")
		util.RespondInternalServerError(c)
		return
	}

	c.JSON(200, gin.H{
		"api_key": apiKey,
	})
	c.Set("message", "generated new API key")
}

// POST /api/user/password - Changes a user's password
func (h *User) ChangePassword(c *gin.Context) {
	errorCol := appctx.ErrorCollector(c.Request.Context())

	var changePasswordReq changePasswordRequest
	if err := c.ShouldBindBodyWithJSON(&changePasswordReq); err != nil {
		errorCol.Add(err, zerolog.InfoLevel, true, nil, "change password request has invalid format")
		util.RespondBadRequest(c, "Invalid change password request")
		return
	}

	contextUser, ok := c.Get("user")
	if !ok {
		errorCol.Add(nil, zerolog.ErrorLevel, false, nil, "user object not found in gin context when expected")
		util.RespondInternalServerError(c)
		return
	}
	user := contextUser.(*model.User)

	if user.PasswordHash == nil {
		util.RespondForbidden(c)
		return
	}

	// Verify old password
	match, err := util.SecureVerifyHash(*user.PasswordHash, changePasswordReq.OldPassword)
	if err != nil {
		errorCol.Add(err, zerolog.ErrorLevel, false, nil, "failed to verify password hash")
		util.RespondInternalServerError(c)
		return
	}
	if !match {
		errorCol.Add(nil, zerolog.InfoLevel, true, nil, "provided old password is incorrect")
		util.RespondBadRequest(c, "Incorrect old password")
		return
	}

	if err := h.authService.ChangePassword(c.Request.Context(), changePasswordReq.NewPassword, user.ID); err != nil {
		var valErr service.PasswordValidationError
		if errors.As(err, &valErr) {
			errorCol.Add(err, zerolog.InfoLevel, true, nil, "new password failed validation")
			util.RespondBadRequest(c, valErr.Error())
			return
		}
		errorCol.Add(err, zerolog.ErrorLevel, false, nil, "failed to update password")
		util.RespondInternalServerError(c)
		return
	}

	c.JSON(200, gin.H{"message": "password changed successfully"})
	c.Set("message", "changed password")
}

// PUT /api/admin/user - Creates a new user with a randomly generated password
func (h *User) Create(c *gin.Context) {
	errorCol := appctx.ErrorCollector(c.Request.Context())

	var req createUserRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		errorCol.Add(err, zerolog.InfoLevel, true, nil, "create user request has invalid format")
		util.RespondBadRequest(c, "Invalid create user request")
		return
	}

	user, password, err := h.userService.CreateWithGeneratedPassword(c.Request.Context(), req.Email, false)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidEmail):
			errorCol.Add(err, zerolog.InfoLevel, true, nil, "invalid email provided for user creation")
			util.RespondBadRequest(c, "Invalid email address")
		case errors.Is(err, service.ErrUserAlreadyExists):
			errorCol.Add(err, zerolog.InfoLevel, true, nil, "user already exists with that email")
			util.RespondConflict(c, "A user with that email already exists")
		default:
			errorCol.Add(err, zerolog.ErrorLevel, false, nil, "failed to create user")
			util.RespondInternalServerError(c)
		}
		return
	}

	c.JSON(201, gin.H{
		"user":     user.ToAPI(),
		"password": password,
	})
	c.Set("message", "created user")
}
