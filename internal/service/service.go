package service

import "github.com/daylamtayari/cierge/internal/repository"

type Services struct {
	Token *TokenService
}

func New(repos repository.Repositories) *Services {
	return &Services{
		Token: NewTokenService(repos.User),
	}
}
