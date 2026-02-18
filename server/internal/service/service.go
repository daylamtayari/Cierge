package service

import (
	"errors"

	"github.com/daylamtayari/cierge/resy"
	"github.com/daylamtayari/cierge/server/cloud"
	"github.com/daylamtayari/cierge/server/internal/config"
	"github.com/daylamtayari/cierge/server/internal/repository"
	tokenstore "github.com/daylamtayari/cierge/server/internal/token_store"
)

var (
	ErrUserNotInContext = errors.New("user object not in context")
)

type Services struct {
	Token         *Token
	User          *User
	Health        *Health
	Auth          *Auth
	Job           *Job
	Reservation   *Reservation
	Restaurant    *Restaurant
	PlatformToken *PlatformToken
	DropConfig    *DropConfig
}

func New(repos *repository.Repositories, cfg *config.Config, tokenStore *tokenstore.Store, cloudProvider cloud.Provider) *Services {
	resyClient := resy.NewClient(nil, resy.Tokens{ApiKey: resy.DefaultApiKey}, "")
	userService := NewUser(repos.User)
	tokenService := NewToken(userService, cfg.Auth, tokenStore)

	return &Services{
		User:          userService,
		Token:         tokenService,
		Health:        NewHealth(repos.DB(), repos.Timeout()),
		Auth:          NewAuth(userService, tokenService, &cfg.Auth),
		Job:           NewJob(repos.Job),
		Reservation:   NewReservation(repos.Reservation),
		Restaurant:    NewRestaurant(repos.Restaurant, resyClient),
		PlatformToken: NewPlatformToken(repos.PlatformToken, cloudProvider),
		DropConfig:    NewDropConfig(repos.DropConfig, repos.Restaurant),
	}
}
