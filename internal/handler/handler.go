package handler

import (
	"github.com/daylamtayari/cierge/internal/config"
	"github.com/daylamtayari/cierge/internal/service"
)

type Handlers struct {
	Auth        *AuthHandler
	JobCallback *JobCallbackHandler
}

func New(services *service.Services, cfg *config.Config) *Handlers {
	return &Handlers{
		Auth:        NewAuthHandler(services.Auth, cfg.IsDevelopment()),
		JobCallback: NewJobCallbackHandler(services.Job),
	}
}
