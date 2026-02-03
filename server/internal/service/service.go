package service

import (
	"github.com/daylamtayari/cierge/server/internal/config"
	"github.com/daylamtayari/cierge/server/internal/repository"
)

type Services struct {
	Token       *Token
	User        *User
	Health      *Health
	Auth        *Auth
	Job         *Job
	Reservation *Reservation
}

func New(repos *repository.Repositories, cfg *config.Config) *Services {
	userService := NewUser(repos.User)
	tokenService := NewToken(userService, cfg.Auth, repos.Revocation)

	return &Services{
		User:        userService,
		Token:       tokenService,
		Health:      NewHealth(repos.DB(), repos.Timeout()),
		Auth:        NewAuth(userService, tokenService, &cfg.Auth),
		Job:         NewJob(repos.Job),
		Reservation: NewReservation(repos.Reservation),
	}
}
