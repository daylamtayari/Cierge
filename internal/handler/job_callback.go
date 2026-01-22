package handler

import (
	"github.com/daylamtayari/cierge/internal/service"
	"github.com/gin-gonic/gin"
)

type JobCallbackHandler struct {
	jobService *service.JobService
}

func NewJobCallbackHandler(jobService *service.JobService) *JobCallbackHandler {
	return &JobCallbackHandler{
		jobService: jobService,
	}
}

func (h *JobCallbackHandler) HandleJobCallback(c *gin.Context) {

}
