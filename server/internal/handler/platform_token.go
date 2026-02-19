package handler

import (
	"errors"
	"strings"

	"github.com/daylamtayari/cierge/api"
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

// GET /api/user/token - Returns a user's platform tokens
// for either the specified platforms or all platforms
// if unspecified
func (h *PlatformToken) Get(c *gin.Context) {
	errorCol := appctx.ErrorCollector(c.Request.Context())

	tokens := make([]*model.PlatformToken, 0)
	platform := strings.ToLower(c.Query("platform"))

	userID := appctx.UserID(c.Request.Context())

	var err error
	switch platform {
	case "":
		tokens, err = h.ptService.GetByUser(c.Request.Context(), userID)

	case "resy", "opentable":
		var token *model.PlatformToken
		token, err = h.ptService.GetByUserAndPlatform(c.Request.Context(), userID, platform)
		if token != nil {
			tokens = append(tokens, token)
		}

	default:
		errorCol.Add(nil, zerolog.InfoLevel, true, map[string]any{"platform": platform}, "unsupported platform specified")
		util.RespondBadRequest(c, "unsupported platform specified")
		return
	}
	if err != nil && !errors.Is(err, service.ErrTokenDNE) {
		errorCol.Add(err, zerolog.ErrorLevel, false, map[string]any{"platform": platform}, "failed to retrieve platform tokens for user")
		util.RespondInternalServerError(c)
		return
	}

	apiTokens := make([]api.PlatformToken, 0)
	for _, token := range tokens {
		apiTokens = append(apiTokens, *token.ToAPI())
	}

	c.JSON(200, apiTokens)
	c.Set("message", "retrieved platform tokens for user")
}

// POST /api/user/token - Creates a new platform token
func (h *PlatformToken) Create(c *gin.Context) {
	errorCol := appctx.ErrorCollector(c.Request.Context())

	platform := strings.ToLower(c.Query("platform"))
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

	newToken, err := h.ptService.Create(c.Request.Context(), appctx.UserID(c.Request.Context()), platform, token)
	if err != nil {
		errorCol.Add(err, zerolog.ErrorLevel, false, map[string]any{"platform": platform}, "error creating platform token")
		util.RespondInternalServerError(c)
		return
	}

	c.JSON(200, newToken)
	c.Set("message", "created new platform token for "+platform)
}
