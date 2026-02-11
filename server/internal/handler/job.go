package handler

import (
	"strconv"

	"github.com/daylamtayari/cierge/api"
	appctx "github.com/daylamtayari/cierge/server/internal/context"
	"github.com/daylamtayari/cierge/server/internal/model"
	"github.com/daylamtayari/cierge/server/internal/service"
	"github.com/daylamtayari/cierge/server/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type Job struct {
	jobService *service.Job
}

func NewJob(jobService *service.Job) *Job {
	return &Job{
		jobService: jobService,
	}
}

// Lists out all of a user's job
func (h *Job) List(c *gin.Context) {
	errorCol := appctx.ErrorCollector(c.Request.Context())

	upcomingQuery := c.DefaultQuery("upcoming", "false")
	upcomingOnly, _ := strconv.ParseBool(upcomingQuery) // Ignore error as false is default behaviour and is what is returned if it also returns an error

	contextUser, ok := c.Get("user")
	if !ok {
		errorCol.Add(nil, zerolog.ErrorLevel, false, nil, "user object not found in gin context when expected")
		util.RespondInternalServerError(c)
		return
	}
	user := contextUser.(*model.User)

	apiJobs := make([]*api.Job, 0)
	for _, job := range user.Jobs {
		if !upcomingOnly {
			apiJobs = append(apiJobs, job.ToAPI())
		} else if job.Status == model.JobStatusCreated || job.Status == model.JobStatusScheduled {
			apiJobs = append(apiJobs, job.ToAPI())
		}
	}

	c.JSON(200, apiJobs)
	c.Set("message", "retrieved own jobs")
}
