package service

import (
	"github.com/daylamtayari/cierge/server/internal/config"
	"github.com/daylamtayari/cierge/server/internal/repository"
	tokenstore "github.com/daylamtayari/cierge/server/internal/token_store"
)

type Services struct {
	Token       *Token
	User        *User
	Health      *Health
	Auth        *Auth
	Job         *Job
	Reservation *Reservation
}

func New(repos *repository.Repositories, cfg *config.Config, tokenStore *tokenstore.Store) *Services {
	userService := NewUser(repos.User)
	tokenService := NewToken(userService, cfg.Auth, tokenStore)

	return &Services{
		User:        userService,
		Token:       tokenService,
		Health:      NewHealth(repos.DB(), repos.Timeout()),
		Auth:        NewAuth(userService, tokenService, &cfg.Auth),
		Job:         NewJob(repos.Job),
		Reservation: NewReservation(repos.Reservation),
	}
}
