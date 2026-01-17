package service

import (
	"github.com/daylamtayari/cierge/internal/config"
	"github.com/daylamtayari/cierge/internal/repository"
)

type Services struct {
	Token  *TokenService
	User   *UserService
	Health *HealthService
	Auth   *AuthService
}

func New(repos *repository.Repositories, cfg *config.Config) *Services {
	userService := NewUserService(repos.User)
	tokenService := NewTokenService(userService, cfg.Auth, repos.Revocation)

	return &Services{
		User:   userService,
		Token:  tokenService,
		Health: NewHealthService(repos.DB(), repos.Timeout()),
		Auth:   NewAuthService(userService, tokenService, &cfg.Auth),
	}
}
