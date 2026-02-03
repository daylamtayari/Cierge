package service

import (
	"github.com/daylamtayari/cierge/server/internal/config"
	"github.com/daylamtayari/cierge/server/internal/repository"
)

type Services struct {
	Token       *TokenService
	User        *UserService
	Health      *HealthService
	Auth        *AuthService
	Job         *JobService
	Reservation *ReservationService
}

func New(repos *repository.Repositories, cfg *config.Config) *Services {
	userService := NewUserService(repos.User)
	tokenService := NewTokenService(userService, cfg.Auth, repos.Revocation)

	return &Services{
		User:        userService,
		Token:       tokenService,
		Health:      NewHealthService(repos.DB(), repos.Timeout()),
		Auth:        NewAuthService(userService, tokenService, &cfg.Auth),
		Job:         NewJobService(repos.Job),
		Reservation: NewReservationService(repos.Reservation),
	}
}
