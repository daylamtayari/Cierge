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
	userService *service.User
}

func NewUser(userService *service.User) *User {
	return &User{
		userService: userService,
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
