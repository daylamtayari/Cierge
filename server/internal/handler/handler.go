package handler

import (
	"github.com/daylamtayari/cierge/server/internal/config"
	"github.com/daylamtayari/cierge/server/internal/service"
)

type Handlers struct {
	Auth          *Auth
	Job           *Job
	JobCallback   *JobCallback
	User          *User
	PlatformToken *PlatformToken
}

func New(services *service.Services, cfg *config.Config) *Handlers {
	return &Handlers{
		Auth:          NewAuth(services.Auth, cfg.IsDevelopment()),
		Job:           NewJob(services.Job),
		JobCallback:   NewJobCallback(services.Job, services.Reservation),
		User:          NewUser(services.User, services.Token),
		PlatformToken: NewPlatformToken(services.PlatformToken),
	}
}
