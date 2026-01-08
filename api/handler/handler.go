package handlers

import (
	"github.com/daylamtayari/cierge/internal/config"
	"github.com/daylamtayari/cierge/internal/service"
)

type Handlers struct {
	Auth *AuthHandler
}

func New(services *service.Services, cfg *config.Config) *Handlers {
	return &Handlers{
		Auth: NewAuthHandler(services.Auth, cfg.IsDevelopment()),
	}
}
