package handler

import (
	"errors"

	"github.com/daylamtayari/cierge/api"
	appctx "github.com/daylamtayari/cierge/server/internal/context"
	"github.com/daylamtayari/cierge/server/internal/service"
	"github.com/daylamtayari/cierge/server/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type DropConfig struct {
	dropConfigService *service.DropConfig
}

func NewDropConfig(dropConfigService *service.DropConfig) *DropConfig {
	return &DropConfig{
		dropConfigService: dropConfigService,
	}
}

type dropConfigCreateRequest struct {
	Restaurant    uuid.UUID
	DaysInAdvance int16
	DropTime      string
}

// GET /api/drop-config - Get drop configs
// Returns a slice of drop configs, ordered by confidence
// in descending order, that is empty if there are no results
func (h *DropConfig) Get(c *gin.Context) {
	logger := appctx.Logger(c.Request.Context())
	errorCol := appctx.ErrorCollector(c.Request.Context())

	restaurantId := c.Query("restaurant")
	if restaurantId == "" {
		errorCol.Add(nil, zerolog.InfoLevel, true, nil, "restaurant ID not specified in request")
		util.RespondBadRequest(c, "Restaurant ID is required")
		return
	}

	logger.UpdateContext(func(c zerolog.Context) zerolog.Context {
		return c.Str("restaurant_id", restaurantId)
	})

	restaurantUid, err := uuid.Parse(restaurantId)
	if err != nil {
		errorCol.Add(err, zerolog.InfoLevel, true, nil, "restaurant ID is not a UUID")
		util.RespondBadRequest(c, "Restaurant ID is invalid")
		return
	}

	dropConfigs, err := h.dropConfigService.GetByRestaurant(c.Request.Context(), restaurantUid)
	if err != nil {
		errorCol.Add(err, zerolog.ErrorLevel, false, nil, "failed to retrieve drop configs for restaurant")
		util.RespondInternalServerError(c)
		return
	}

	dropConfigResponse := make([]api.DropConfig, 0)
	for _, dropConfig := range dropConfigs {
		dropConfigResponse = append(dropConfigResponse, *dropConfig.ToAPI())
	}

	c.JSON(200, dropConfigResponse)
	c.Set("message", "retrieved drop configs for restaurant")
}

// POST /api/drop-config - Create drop config
// If a drop config with the same parameters exist,
// return the existing drop config
func (h *DropConfig) Create(c *gin.Context) {
	errorCol := appctx.ErrorCollector(c.Request.Context())

	var dropConfigCreateReq dropConfigCreateRequest
	if err := c.ShouldBindBodyWithJSON(&dropConfigCreateReq); err != nil {
		errorCol.Add(err, zerolog.InfoLevel, true, nil, "drop config creation request has improper format")
		util.RespondBadRequest(c, "Invalid drop configuration creation request")
		return
	}

	dropConfig, err := h.dropConfigService.Create(c.Request.Context(), dropConfigCreateReq.Restaurant, dropConfigCreateReq.DaysInAdvance, dropConfigCreateReq.DropTime)
	if err != nil && errors.Is(err, service.ErrInvalidDropTime) {
		errorCol.Add(err, zerolog.InfoLevel, true, nil, "drop time in drop config creation invalid")
		util.RespondBadRequest(c, "Drop time is invalid format")
		return
	} else if err != nil {
		errorCol.Add(err, zerolog.ErrorLevel, false, nil, "failed to create drop configuration")
		util.RespondInternalServerError(c)
		return
	}

	c.JSON(200, dropConfig.ToAPI())
	c.Set("message", "created drop config")
}
