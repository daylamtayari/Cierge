package handler

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/daylamtayari/cierge/api"
	appctx "github.com/daylamtayari/cierge/server/internal/context"
	"github.com/daylamtayari/cierge/server/internal/model"
	"github.com/daylamtayari/cierge/server/internal/service"
	"github.com/daylamtayari/cierge/server/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type Job struct {
	jobService        *service.Job
	restaurantService *service.Restaurant
	dropConfigService *service.DropConfig
}

func NewJob(jobService *service.Job, restaurantService *service.Restaurant, dropConfigService *service.DropConfig) *Job {
	return &Job{
		jobService:        jobService,
		restaurantService: restaurantService,
		dropConfigService: dropConfigService,
	}
}

// GET /api/job/list - Lists out all of a user's job
func (h *Job) List(c *gin.Context) {
	errorCol := appctx.ErrorCollector(c.Request.Context())

	upcomingQuery := strings.ToLower(c.DefaultQuery("upcoming", "false"))
	upcomingOnly, _ := strconv.ParseBool(upcomingQuery) // Ignore error as false is default behaviour and is what is returned if it also returns an error

	jobs, err := h.jobService.GetByUser(c.Request.Context(), appctx.UserID(c.Request.Context()))
	if err != nil && !errors.Is(err, service.ErrJobDNE) {
		errorCol.Add(err, zerolog.ErrorLevel, false, nil, "failed to fetch jobs for user")
		util.RespondInternalServerError(c)
		return
	}

	apiJobs := make([]*api.Job, 0)
	for _, job := range jobs {
		if !upcomingOnly {
			apiJobs = append(apiJobs, job.ToAPI())
		} else if job.Status == model.JobStatusCreated || job.Status == model.JobStatusScheduled {
			apiJobs = append(apiJobs, job.ToAPI())
		}
	}

	c.JSON(200, apiJobs)
	c.Set("message", "retrieved own jobs")
}

// POST /api/job - Create a new job
func (h *Job) Create(c *gin.Context) {
	errorCol := appctx.ErrorCollector(c.Request.Context())
	logger := appctx.Logger(c.Request.Context())

	var jobCreationReq api.JobCreationRequest
	if err := c.ShouldBindBodyWithJSON(&jobCreationReq); err != nil {
		errorCol.Add(err, zerolog.InfoLevel, true, nil, "job creation request has improper format")
		util.RespondBadRequest(c, "Invalid job creation request")
		return
	}

	logger.UpdateContext(func(c zerolog.Context) zerolog.Context {
		return c.
			Str("restaurant_id", jobCreationReq.RestaurantID.String()).
			Str("reservation_date", jobCreationReq.ReservationDate).
			Int16("party_size", jobCreationReq.PartySize).
			Strs("preferred_times", jobCreationReq.PreferredTimes).
			Str("drop_config_id", jobCreationReq.DropConfigID.String())
	})

	// Validation of request
	restaurant, err := h.restaurantService.GetByID(c.Request.Context(), jobCreationReq.RestaurantID)
	if err != nil && errors.Is(err, service.ErrRestaurantDNE) {
		errorCol.Add(err, zerolog.InfoLevel, true, nil, "no restaurant exists with specified ID")
		util.RespondBadRequest(c, "Invalid restaurant ID")
		return
	} else if err != nil {
		errorCol.Add(err, zerolog.ErrorLevel, false, nil, "failed to retrieve restaurant")
		util.RespondInternalServerError(c)
		return
	}
	reservationDate, err := time.Parse("2006-01-02", jobCreationReq.ReservationDate)
	if err != nil {
		errorCol.Add(err, zerolog.InfoLevel, true, nil, "invalid reservation date")
		util.RespondBadRequest(c, "Invalid reservation date")
	}
	if time.Now().After(reservationDate) {
		errorCol.Add(nil, zerolog.InfoLevel, true, nil, "reservation date is in the past")
		util.RespondBadRequest(c, "Reservation date is in the past")
		return
	}
	for _, preferredTime := range jobCreationReq.PreferredTimes {
		_, err := time.Parse("15:04", preferredTime)
		if err != nil {
			errorCol.Add(err, zerolog.InfoLevel, true, nil, "invalid preferred time")
			util.RespondBadRequest(c, "Invalid preferred time")
			return
		}
	}
	dropConfig, err := h.dropConfigService.GetByID(c.Request.Context(), jobCreationReq.DropConfigID)
	if err != nil && errors.Is(err, service.ErrDropConfigDNE) {
		errorCol.Add(err, zerolog.InfoLevel, true, nil, "no drop config exists with specified ID")
		util.RespondBadRequest(c, "Invalid drop configuration ID")
		return
	}
	if h.dropConfigService.IsScheduledAtPast(dropConfig, reservationDate, restaurant.Timezone.Location) {
		errorCol.Add(nil, zerolog.InfoLevel, true, nil, "job execution date is in the past")
		util.RespondBadRequest(c, "Job cannot be scheduled in the past")
		return
	}

	// Creation of job
	job, err := h.jobService.Create(c.Request.Context(), &jobCreationReq, restaurant, dropConfig)
	if err != nil {
		errorCol.Add(err, zerolog.ErrorLevel, false, nil, "failed to create job")
		util.RespondInternalServerError(c)
		return
	}
	err = h.dropConfigService.IncrementConfidence(c.Request.Context(), dropConfig.ID, restaurant.ID)
	if err != nil {
		errorCol.Add(err, zerolog.ErrorLevel, false, nil, "failed to increment drop config confidence")
	}
	err = h.jobService.Schedule(c.Request.Context(), job, restaurant)
	if err != nil && errors.Is(err, service.ErrTokenDNE) {
		errorCol.Add(err, zerolog.InfoLevel, false, nil, "no token configured")
		util.RespondConflict(c, "Platform token not configured for platform")
		return
	} else if err != nil {
		errorCol.Add(err, zerolog.ErrorLevel, false, nil, "failed to schedule job")
		util.RespondInternalServerError(c)
		return
	}
	err = h.jobService.UpdateStatus(c.Request.Context(), model.JobStatusScheduled, job.ID)
	if err != nil {
		errorCol.Add(err, zerolog.ErrorLevel, false, nil, "failed to mark job as scheduled")
		// Don't return an error as the job was successfully scheduled
	}

	c.JSON(200, job.ToAPI())
	c.Set("message", "created and scheduled job")
}

// POST /api/job/:job/cancel
func (h *Job) Cancel(c *gin.Context) {
	errorCol := appctx.ErrorCollector(c.Request.Context())
	logger := appctx.Logger(c.Request.Context())

	jobId := c.Param("job")
	if jobId == "" {
		errorCol.Add(nil, zerolog.InfoLevel, true, nil, "no job ID specified")
		util.RespondBadRequest(c, "Job ID must be specified")
		return
	}
	jobUid, err := uuid.Parse(jobId)
	if err != nil {
		errorCol.Add(err, zerolog.InfoLevel, true, nil, "job ID specified is not a UUID")
		util.RespondBadRequest(c, "Job ID must be a valid UUID")
		return
	}

	logger.UpdateContext(func(c zerolog.Context) zerolog.Context {
		return c.Str("job_id", jobUid.String())
	})

	job, err := h.jobService.GetByID(c.Request.Context(), jobUid)
	if err != nil && errors.Is(err, service.ErrJobDNE) {
		errorCol.Add(err, zerolog.InfoLevel, true, nil, "no job exists with specified ID")
		util.RespondNotFound(c, "Job with specified ID not found")
		return
	} else if err != nil {
		errorCol.Add(err, zerolog.ErrorLevel, false, nil, "failed to retrieve job from ID")
		util.RespondInternalServerError(c)
		return
	}

	switch job.Status {
	case model.JobStatusCancelled:
		c.Status(200)
		c.Set("message", "job already cancelled")
		return
	case model.JobStatusSuccess, model.JobStatusFailed:
		errorCol.Add(nil, zerolog.InfoLevel, true, nil, "job cannot be cancelled as it has been executed")
		util.RespondConflict(c, "Job has already been executed")
		return
	}

	if time.Now().After(job.ScheduledAt) {
		errorCol.Add(nil, zerolog.InfoLevel, true, nil, "job cannot be cancelled as it's past its start time")
		util.RespondConflict(c, "Job execution has already started")
		return
	}

	err = h.jobService.Cancel(c.Request.Context(), job.ID)
	if err != nil {
		errorCol.Add(err, zerolog.ErrorLevel, false, nil, "failed to cancel job")
		util.RespondInternalServerError(c)
		return
	}
	err = h.jobService.UpdateStatus(c.Request.Context(), model.JobStatusCancelled, job.ID)
	if err != nil {
		errorCol.Add(err, zerolog.ErrorLevel, false, nil, "failed to mark job as cancelled")
	}

	c.Status(200)
	c.Set("message", "cancelled job")
}
