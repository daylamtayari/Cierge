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
	Restaurant    *Restaurant
	PlatformToken *PlatformToken
	DropConfig    *DropConfig
}

func New(services *service.Services, cfg *config.Config) *Handlers {
	return &Handlers{
		Auth:          NewAuth(services.Auth, cfg.IsDevelopment()),
		Job:           NewJob(services.Job, services.Restaurant, services.DropConfig),
		JobCallback:   NewJobCallback(services.Job, services.Reservation),
		User:          NewUser(services.User, services.Token),
		Restaurant:    NewRestaurant(services.Restaurant),
		PlatformToken: NewPlatformToken(services.PlatformToken),
		DropConfig:    NewDropConfig(services.DropConfig),
	}
}
