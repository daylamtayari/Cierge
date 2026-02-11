package api

import (
	"net/http"
	"time"

	"github.com/google/uuid"
)

type AuthMethod string

const (
	LocalAuthMethod = AuthMethod("local")
	OIDCAuthMethod  = AuthMethod("oidc")
)

type User struct {
	ID          uuid.UUID  `json:"id"`
	Email       string     `json:"email"`
	HasApiKey   bool       `json:"has_api_key"`
	IsAdmin     bool       `json:"is_admin"`
	AuthMethod  AuthMethod `json:"auth_method"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// Returns a User pointer representing the authenticated
// user making the request
func (c *Client) GetMe() (*User, error) {
	reqUrl := c.host + "/api/user/me"
	req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
	if err != nil {
		return nil, err
	}

	var user User
	err = c.Do(req, &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
