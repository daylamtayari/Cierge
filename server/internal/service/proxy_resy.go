package service

import (
	"context"

	"github.com/daylamtayari/cierge/resy"
)

type ProxyResy struct {
	resyClient *resy.Client
}

func NewProxyResy(resyClient *resy.Client) *ProxyResy {
	return &ProxyResy{
		resyClient: resyClient,
	}
}

func (s *ProxyResy) Auth(ctx context.Context, email string, password string) (resy.Tokens, error) {
	authResyClient := resy.NewClient(nil, resy.Tokens{}, "")
	return authResyClient.Login(email, password)
}
