package handler

import (
	"errors"
	"strconv"
	"strings"

	"github.com/daylamtayari/cierge/api"
	appctx "github.com/daylamtayari/cierge/server/internal/context"
	"github.com/daylamtayari/cierge/server/internal/model"
	"github.com/daylamtayari/cierge/server/internal/service"
	"github.com/daylamtayari/cierge/server/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type Reservation struct {
	reservationService *service.Reservation
}

func NewReservation(reservationService *service.Reservation) *Reservation {
	return &Reservation{
		reservationService: reservationService,
	}
}

// GET /api/reservation/list - List out all of a user's reservations
func (h *Reservation) List(c *gin.Context) {
	errorCol := appctx.ErrorCollector(c.Request.Context())

	upcomingQuery := strings.ToLower(c.DefaultQuery("upcoming", "false"))
	upcomingOnly, _ := strconv.ParseBool(upcomingQuery) // Ignore error as false is default behaviour and is what is returned if it also returns an error

	var res []*model.Reservation
	var err error

	if upcomingOnly {
		res, err = h.reservationService.GetByUserUpcoming(c.Request.Context(), appctx.UserID(c.Request.Context()))
	} else {
		res, err = h.reservationService.GetByUser(c.Request.Context(), appctx.UserID(c.Request.Context()))
	}
	if err != nil {
		errorCol.Add(err, zerolog.ErrorLevel, false, nil, "failed to fetch reservations for user")
		util.RespondInternalServerError(c)
		return
	}

	apiRes := make([]*api.Reservation, len(res))
	for _, r := range res {
		apiRes = append(apiRes, r.ToAPI())
	}

	c.JSON(200, apiRes)
	c.Set("message", "retrieved own reservations")
}

// GET /api/reservation/:reservation - Retrieve a specified reservation
func (h *Reservation) Get(c *gin.Context) {
	errorCol := appctx.ErrorCollector(c.Request.Context())

	resUid, err := uuid.Parse(c.Param("reservation"))
	if err != nil {
		errorCol.Add(err, zerolog.InfoLevel, true, nil, "invalid reservation ID")
		util.RespondBadRequest(c, "Reservation ID must be a valid UUID")
		return
	}

	res, err := h.reservationService.GetByID(c.Request.Context(), resUid)
	if err != nil && errors.Is(err, service.ErrReservationDNE) {
		util.RespondNotFound(c, "Reservation not found")
		return
	} else if err != nil {
		errorCol.Add(err, zerolog.ErrorLevel, false, nil, "failed to retrieve reservation")
		util.RespondInternalServerError(c)
		return
	}

	if res.UserID == appctx.UserID(c.Request.Context()) || c.GetBool("is_admin") {
		c.JSON(200, res.ToAPI())
		c.Set("message", "retrieved reservation")
	} else {
		util.RespondNotFound(c, "Reservation not found")
		return
	}
}
