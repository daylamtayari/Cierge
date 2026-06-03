package handler

import (
	"errors"
	"net/http"

	"github.com/daylamtayari/cierge/resy"
	appctx "github.com/daylamtayari/cierge/server/internal/context"
	"github.com/daylamtayari/cierge/server/internal/service"
	"github.com/daylamtayari/cierge/server/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type Proxy struct {
	proxyResyService *service.ProxyResy
	ptService        *service.PlatformToken
}

func NewProxy(proxyResyService *service.ProxyResy, ptService *service.PlatformToken) *Proxy {
	return &Proxy{
		proxyResyService: proxyResyService,
		ptService:        ptService,
	}
}

func (h *Proxy) ResyAuth(c *gin.Context) {
	errorCol := appctx.ErrorCollector(c.Request.Context())

	var resyAuthReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindBodyWithJSON(&resyAuthReq); err != nil {
		errorCol.Add(err, zerolog.InfoLevel, true, nil, "resy proxy login request has improper format")
		util.RespondBadRequest(c, "Invalid Resy login request")
		return
	}

	tokens, err := h.proxyResyService.Auth(c.Request.Context(), resyAuthReq.Email, resyAuthReq.Password)
	if errors.Is(err, resy.ErrUnauthorized) {
		// Only reason the email is logged on failed logins due to invalid credentials is to monitor abuse
		errorCol.Add(err, zerolog.InfoLevel, true, map[string]any{"email": resyAuthReq.Email}, "failed to login to resy due to incorrect resy credentials")
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error":      "Forbidden",
			"message":    "Incorrect Resy credentials",
			"request_id": appctx.RequestID(c.Request.Context()),
		})
		return
	} else if err != nil {
		errorCol.Add(err, zerolog.ErrorLevel, false, nil, "resy auth request failed")
		util.RespondFailedDep(c, "Failed to perform authentication to Resy")
		return
	}

	newToken, err := h.ptService.Create(c.Request.Context(), appctx.UserID(c.Request.Context()), "resy", tokens)
	if err != nil {
		errorCol.Add(err, zerolog.ErrorLevel, false, map[string]any{"platform": "resy"}, "error creating platform token")
		util.RespondInternalServerError(c)
		return
	}

	c.JSON(200, newToken.ToAPI())
	c.Set("message", "created new platform token for resy")
}

func (h *Proxy) ResyRestaurant(c *gin.Context) {
	errorCol := appctx.ErrorCollector(c.Request.Context())

	var resyResReq struct {
		Query string `json:"query"`
	}
	if err := c.ShouldBindBodyWithJSON(&resyResReq); err != nil {
		errorCol.Add(err, zerolog.InfoLevel, true, nil, "restaurant query has improper format")
		util.RespondBadRequest(c, "Invalid query format")
		return
	}

	venues, err := h.proxyResyService.Restaurant(c.Request.Context(), resyResReq.Query)
	if err != nil {
		errorCol.Add(err, zerolog.ErrorLevel, false, map[string]any{"query": resyResReq.Query}, "failed to proxy restaurant search")
		util.RespondInternalServerError(c)
		return
	}

	c.JSON(200, venues)
	c.Set("message", "proxied search for resy restaurants")
}
