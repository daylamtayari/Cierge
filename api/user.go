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

type passwordChangeRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
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

// Change a user's password
// NOTE: Password requirements: upper case, number, special character, min 8, max 128
func (c *Client) ChangePassword(oldPassword string, newPassword string) error {
	reqUrl := c.host + "/api/user/password"
	passwordChangeReq := passwordChangeRequest{
		OldPassword: oldPassword,
		NewPassword: newPassword,
	}
	req, err := c.NewJsonRequest(http.MethodPost, reqUrl, passwordChangeReq)
	if err != nil {
		return err
	}

	err = c.Do(req, nil)
	if err != nil {
		return err
	}

	return nil
}
