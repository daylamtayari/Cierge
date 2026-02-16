package handler

import (
	"github.com/daylamtayari/cierge/resy"
	appctx "github.com/daylamtayari/cierge/server/internal/context"
	"github.com/daylamtayari/cierge/server/internal/model"
	"github.com/daylamtayari/cierge/server/internal/service"
	"github.com/daylamtayari/cierge/server/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type PlatformToken struct {
	ptService *service.PlatformToken
}

func NewPlatformToken(platformTokenService *service.PlatformToken) *PlatformToken {
	return &PlatformToken{
		ptService: platformTokenService,
	}
}

// POST /api/user/token - Creates a new platform token
func (h *PlatformToken) Create(c *gin.Context) {
	errorCol := appctx.ErrorCollector(c.Request.Context())

	platform := c.Query("platform")
	if platform == "" {
		errorCol.Add(nil, zerolog.InfoLevel, true, nil, "platform not specified in query parameters")
		util.RespondBadRequest(c, "platform not specified")
		return
	}

	var token any
	switch platform {
	case "resy":
		var resyToken resy.Tokens
		if err := c.ShouldBindBodyWithJSON(&resyToken); err != nil {
			errorCol.Add(nil, zerolog.InfoLevel, true, nil, "incorrect token format for resy token")
			util.RespondBadRequest(c, "incorrect token format")
			return
		}
		if resyToken.ApiKey == "" {
			resyToken.ApiKey = resy.DefaultApiKey
		}
		token = resyToken

	case "opentable":
		// TODO: Complete opentable section
	default:
		errorCol.Add(nil, zerolog.InfoLevel, true, map[string]any{"platform": platform}, "unsupported platform specified")
		util.RespondBadRequest(c, "unsupported platform specified")
		return
	}

	contextUser, ok := c.Get("user")
	if !ok {
		errorCol.Add(nil, zerolog.ErrorLevel, false, nil, "user object not found in gin context when expected")
		util.RespondInternalServerError(c)
		return
	}
	user := contextUser.(*model.User)

	err := h.ptService.Create(c.Request.Context(), user.ID, platform, token)
	if err != nil {
		errorCol.Add(err, zerolog.ErrorLevel, false, map[string]any{"platform": platform}, "error creating platform token")
		util.RespondInternalServerError(c)
		return
	}
}
