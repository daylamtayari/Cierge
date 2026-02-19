package handler

import (
	"errors"
	"strconv"
	"strings"

	"github.com/daylamtayari/cierge/resy"
	appctx "github.com/daylamtayari/cierge/server/internal/context"
	"github.com/daylamtayari/cierge/server/internal/service"
	"github.com/daylamtayari/cierge/server/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type Restaurant struct {
	restaurantService *service.Restaurant
}

func NewRestaurant(restaurantService *service.Restaurant) *Restaurant {
	return &Restaurant{
		restaurantService: restaurantService,
	}
}

// GET /api/restaurant - Return a restaurant object,
// fetching and creating it if the specific restaurant
// is not already stored
func (h *Restaurant) Get(c *gin.Context) {
	logger := appctx.Logger(c.Request.Context())
	errorCol := appctx.ErrorCollector(c.Request.Context())

	platform := strings.ToLower(c.Query("platform"))
	platformId := strings.ToLower(c.Query("platform-id"))

	logger.UpdateContext(func(c zerolog.Context) zerolog.Context {
		return c.
			Str("platform", platform).
			Str("platform_id", platformId)
	})

	if platform != "resy" && platform != "opentable" {
		errorCol.Add(nil, zerolog.InfoLevel, true, nil, "unsupported platform specified")
		util.RespondBadRequest(c, "unsupported platform specified")
		return
	}

	restaurant, err := h.restaurantService.GetByPlatformID(c.Request.Context(), platform, platformId)
	if err != nil && errors.Is(err, service.ErrRestaurantDNE) {
		restaurant, err = h.restaurantService.Create(c.Request.Context(), platform, platformId)
		switch {
		case errors.Is(err, strconv.ErrSyntax):
			errorCol.Add(err, zerolog.InfoLevel, true, nil, "platform ID provided could not be converted to its expected type")
			util.RespondBadRequest(c, "restaurant platform ID contains invalid values")
			return
		case errors.Is(err, resy.ErrNotFound):
			errorCol.Add(err, zerolog.InfoLevel, true, nil, "restaurant with specified ID does not exist on platform")
			util.RespondNotFound(c, "restaurant platform ID does not match any restaurant on platform")
			return
		case err != nil:
			errorCol.Add(err, zerolog.ErrorLevel, false, nil, "failed to create restaurant")
			util.RespondInternalServerError(c)
			return
		}

		c.Set("message", "created and returned restaurant")
	} else if err != nil {
		errorCol.Add(err, zerolog.ErrorLevel, false, nil, "failed to retrieve restaurant")
		util.RespondInternalServerError(c)
		return
	} else {
		c.Set("message", "retrieved restaurant")
	}

	c.JSON(200, restaurant.ToAPI())
}

// GET /api/restaurant/:id - Return a restaurant by its ID
func (h *Restaurant) GetByID(c *gin.Context) {
	errorCol := appctx.ErrorCollector(c.Request.Context())
	logger := appctx.Logger(c.Request.Context())

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errorCol.Add(err, zerolog.InfoLevel, true, nil, "invalid restaurant ID")
		util.RespondBadRequest(c, "Invalid restaurant ID")
		return
	}

	logger.UpdateContext(func(c zerolog.Context) zerolog.Context {
		return c.Str("restaurant_id", id.String())
	})

	restaurant, err := h.restaurantService.GetByID(c.Request.Context(), id)
	if err != nil && errors.Is(err, service.ErrRestaurantDNE) {
		util.RespondNotFound(c, "Restaurant not found")
		return
	} else if err != nil {
		errorCol.Add(err, zerolog.ErrorLevel, false, nil, "failed to retrieve restaurant")
		util.RespondInternalServerError(c)
		return
	}

	c.JSON(200, restaurant.ToAPI())
	c.Set("message", "retrieved restaurant")
}
