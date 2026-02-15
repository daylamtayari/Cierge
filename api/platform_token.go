package api

import (
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
