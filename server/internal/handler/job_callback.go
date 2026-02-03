package handler

import (
	"net/http"

	"github.com/daylamtayari/cierge/reservation"
	appctx "github.com/daylamtayari/cierge/server/internal/context"
	"github.com/daylamtayari/cierge/server/internal/model"
	"github.com/daylamtayari/cierge/server/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type JobCallback struct {
	jobService         *service.Job
	reservationService *service.Reservation
}

func NewJobCallback(jobService *service.Job, reservationService *service.Reservation) *JobCallback {
	return &JobCallback{
		jobService:         jobService,
		reservationService: reservationService,
	}
}

// Handles a callback request from a job output and updates the
// job value, creates a reservation, and send a notification
func (h *JobCallback) HandleJobCallback(c *gin.Context) {
	errorCol := appctx.ErrorCollector(c.Request.Context())

	var callbackReq reservation.Output
	if err := c.ShouldBindJSON(&callbackReq); err != nil {
		errorCol.Add(err, zerolog.WarnLevel, true, nil, "job callback request has improper format")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":      "Invalid request",
			"request_id": appctx.RequestID(c.Request.Context()),
		})
		return
	}

	contextJob, ok := c.Get("job")
	if !ok {
		errorCol.Add(nil, zerolog.ErrorLevel, false, nil, "job object not found in context")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error":      "Internal server error",
			"request_id": appctx.RequestID(c.Request.Context()),
		})
		return
	}
	job, ok := contextJob.(*model.Job)
	if !ok {
		errorCol.Add(nil, zerolog.ErrorLevel, false, map[string]any{"job": contextJob}, "job object in context is not a pointer to a Job type")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error":      "Internal server error",
			"request_id": appctx.RequestID(c.Request.Context()),
		})
		return
	}

	if callbackReq.JobId != job.ID {
		errorCol.Add(nil, zerolog.ErrorLevel, false, map[string]any{"callback_job_id": callbackReq.JobId, "retrieved_job_id": job.ID}, "job ID in the callback is different than the retrieved job")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error":      "Internal server error",
			"request_id": appctx.RequestID(c.Request.Context()),
		})
		return
	}

	updatedJob, err := h.jobService.UpdateFromCallback(c.Request.Context(), job, callbackReq)
	if err != nil {
		errorCol.Add(err, zerolog.ErrorLevel, false, map[string]any{"job": updatedJob}, "failed to update job from callback")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error":      "Internal server error",
			"request_id": appctx.RequestID(c.Request.Context()),
		})
		return
	}

	_, err = h.reservationService.CreateFromJob(c.Request.Context(), updatedJob)
	if err != nil {
		errorCol.Add(err, zerolog.ErrorLevel, false, map[string]any{"job": updatedJob}, "failed to create reservation from job")
	}

	// TODO: Send notification

	c.JSON(http.StatusOK, gin.H{
		"message": "Callback request accepted successfully",
	})
	c.Set("message", "received and handled job callback")
}
