package handler

import (
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
}

func NewUser(userService *service.User, tokenService *service.Token) *User {
	return &User{
		userService:  userService,
		tokenService: tokenService,
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
	c.Set("message", "returned user")
}

// Generates a new API key for a user
// NOTE: This will replace the existing API key and as such
// will invalidate any past API keys. Highly recommend fetching
// the user and checking if they have an active API key first and
// if so, getting explicit confirmation.
func (h *User) APIKey(c *gin.Context) {
	errorCol := appctx.ErrorCollector(c.Request.Context())

	contextUser, ok := c.Get("user")
	if !ok {
		errorCol.Add(nil, zerolog.ErrorLevel, false, nil, "user object not found in gin context when expected")
		util.RespondInternalServerError(c)
		return
	}
	user := contextUser.(*model.User)

	apiKey, err := h.tokenService.GenerateAPIKey(c.Request.Context(), user.ID)
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
