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

func (s *ProxyResy) Restaurant(ctx context.Context, query string) ([]resy.Venue, error) {
	// Use the default page limit of 10
	return s.resyClient.SearchVenue(query, nil)
}
