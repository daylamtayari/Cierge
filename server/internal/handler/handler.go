package handler

import (
	"github.com/daylamtayari/cierge/server/internal/config"
	"github.com/daylamtayari/cierge/server/internal/service"
)

type Handlers struct {
	Auth        *Auth
	JobCallback *JobCallback
}

func New(services *service.Services, cfg *config.Config) *Handlers {
	return &Handlers{
		Auth:        NewAuth(services.Auth, cfg.IsDevelopment()),
		JobCallback: NewJobCallback(services.Job, services.Reservation),
	}
}
