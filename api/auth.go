package api

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrNoAPIKey      = errors.New("no API key was returned by the server")
	ErrNoAuthCookies = errors.New("no auth cookies were found")
)

type AuthCookies struct {
	AccessToken  string
	RefreshToken string
}

// Login and get a user's authentication cookies if successful
// This is designed to be used for providing a smooth way of
// retrieving a user's API key if they don't already have it
// NOTE: Only supports username:password auth at this time
func (c *Client) Login(email string, password string) (*AuthCookies, error) {
	reqUrl := c.host + "/auth/login"
	loginReq := struct {
		Email    string
		Password string
	}{
		Email:    email,
		Password: password,
	}

	req, err := c.NewJsonRequest(http.MethodPost, reqUrl, loginReq)
	if err != nil {
		return nil, err
	}

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close() //nolint:errcheck

	switch res.StatusCode {
	case 200:

		authCookies := AuthCookies{}

		cookies := res.Cookies()
		for _, cookie := range cookies {
			switch cookie.Name {
			case "access_token":
				authCookies.AccessToken = cookie.Value
			case "refresh_token":
				authCookies.RefreshToken = cookie.Value
			}
		}

		if authCookies.AccessToken == "" && authCookies.RefreshToken == "" {
			return nil, ErrNoAuthCookies
		}
		return &authCookies, nil
	case 401:
		return nil, ErrUnauthenticated
	case 500:
		return nil, ErrServerError
	default:
		return nil, fmt.Errorf("%w: %d", ErrUnhandledStatus, res.StatusCode)
	}
}

// Generates a new API key for a user
// NOTE: Requires authentication (if using cookie auth
// it must be set separately)
// NOTE: This will replace the existing API key and as such
// will invalidate any past API keys. Highly recommend fetching
// the user and checking if they have an active API key first and
// if so, getting explicit confirmation.
func (c *Client) GenerateAPIKey() (string, error) {
	reqUrl := c.host + "/api/user/api-key"
	req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
	if err != nil {
		return "", err
	}

	var apiKey struct {
		ApiKey string `json:"api_key"`
	}
	err = c.Do(req, &apiKey)
	if err != nil {
		return "", err
	}

	if apiKey.ApiKey == "" {
		return "", ErrNoAPIKey
	}
	return apiKey.ApiKey, nil
}
