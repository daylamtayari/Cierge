package service

import (
	"github.com/daylamtayari/cierge/internal/config"
	"github.com/daylamtayari/cierge/internal/repository"
)

type Services struct {
	Token  *TokenService
	User   *UserService
	Health *HealthService
}

func New(repos *repository.Repositories, cfg *config.Config) *Services {
	return &Services{
		User:   NewUserService(repos.User),
		Token:  NewTokenService(repos.User, cfg.Auth),
		Health: NewHealthService(repos.DB(), repos.Timeout()),
	}
}
