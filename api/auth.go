package api

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrNoAuthCookies = errors.New("no auth cookies were found")
)

type AuthCookies struct {
	AccessToken  string
	RefreshToken string
}

// Login
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
