package api

import (
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
