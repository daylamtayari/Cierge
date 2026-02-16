package api

import (
	"net/http"
	"time"

	"github.com/google/uuid"
)

// NOTE: Token values are not retrievable via the API
type PlatformToken struct {
	ID               uuid.UUID  `json:"id"`
	UserID           uuid.UUID  `json:"user_id"`
	Platform         string     `json:"platform"`
	ExpiresAt        *time.Time `json:"expires_at"`
	HasRefresh       bool       `json:"has_refresh"`
	RefreshExpiresAt *time.Time `json:"refresh_expires_at"`
	CreatedAt        time.Time  `json:"created_at"`
}

// Retrieve's a users platform tokens for either the specified platform or all
// platforms if platform is nil
func (c *Client) GetPlatformTokens(platform *string) ([]PlatformToken, error) {
	pltfrm := ""
	if platform != nil {
		pltfrm = *platform
	}
	reqUrl := c.host + "/api/user/token?platform=" + pltfrm
	req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
	if err != nil {
		return nil, err
	}

	var platformTokens []PlatformToken
	err = c.Do(req, &platformTokens)
	if err != nil {
		return nil, err
	}

	return platformTokens, nil
}

// Create a new platform token for the specified platform using the specified token
// The token structure must match that of the platform and is validated server-side
func (c *Client) CreatePlatformToken(platform string, token any) (PlatformToken, error) {
	reqUrl := c.host + "/api/user/token?platform=" + platform
	req, err := c.NewJsonRequest(http.MethodPost, reqUrl, token)
	if err != nil {
		return PlatformToken{}, err
	}

	var platformToken PlatformToken
	err = c.Do(req, &platformToken)
	if err != nil {
		return PlatformToken{}, err
	}

	return platformToken, nil
}
