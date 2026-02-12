package resy

import (
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrFailedToGetClaims = errors.New("failed to get claims from a parsed JWT token")
	ErrNoAuthToken       = errors.New("no auth token was included in the response")
	ErrNoExpirationTime  = errors.New("JWT token has no expiration time specified")
	ErrNoRefreshToken    = errors.New("no refresh token cookie was included in the response")
)

type AuthTokens struct {
	// Authentication token for the user to be used in requests - 45 day expiration
	Token string
	// Refresh token that can be used to get an updated refresh and auth token - 90 day expiration
	Refresh string
}

// Performs authentication using username and password auth, returning
// the user's auth JWT token and an error that is nil if successful
func (c *Client) Login(email string, password string) (AuthTokens, error) {
	reqUrl := Host + "/4/auth/password"

	reqForm := url.Values{
		"email":    []string{email},
		"password": []string{password},
	}

	req, err := c.NewFormRequest(http.MethodPost, reqUrl, &reqForm)
	if err != nil {
		return AuthTokens{}, err
	}

	return c.makeAuthRequest(req)
}

// Uses a provided refresh token to retrieve a new auth token (45 day expiration)
// and a new refresh token (additional 90 day expiration)
// Returns an error that is nil if successful
func (c *Client) RefreshToken(refreshToken string) (AuthTokens, error) {
	reqUrl := Host + "/3/auth/refresh"

	req, err := http.NewRequest(http.MethodPost, reqUrl, nil)
	if err != nil {
		return AuthTokens{}, err
	}

	return c.makeAuthRequest(req)
}

// Handles an authentication request and retrieves the auth and refresh tokens
func (c *Client) makeAuthRequest(req *http.Request) (AuthTokens, error) {
	type loginResponse struct {
		Token string `json:"token"`
	}
	var loginRes loginResponse

	cookies, err := c.DoWithCookies(req, &loginRes)
	if err != nil {
		return AuthTokens{}, err
	}

	if loginRes.Token == "" {
		return AuthTokens{}, ErrNoAuthToken
	}

	refreshToken := ""
	for _, cookie := range cookies {
		if cookie.Name == "production_refresh_token" {
			refreshToken = cookie.Value
		}
	}
	if refreshToken == "" {
		return AuthTokens{}, ErrNoRefreshToken
	}

	return AuthTokens{Token: loginRes.Token, Refresh: refreshToken}, nil
}

// Retrieves the expiration time of a JWT token
func getTokenExpiry(jwtToken string) (time.Time, error) {
	token, _, err := jwt.NewParser().ParseUnverified(jwtToken, &jwt.RegisteredClaims{})
	if err != nil {
		return time.Time{}, err
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return time.Time{}, ErrFailedToGetClaims
	}

	if claims.ExpiresAt == nil {
		return time.Time{}, ErrNoExpirationTime
	}

	return claims.ExpiresAt.Time, nil
}
