package api

import (
	"time"

	"github.com/google/uuid"
)

type Restaurant struct {
	ID         uuid.UUID      `json:"id"`
	Platform   string         `json:"platform"`
	PlatformID string         `json:"platform_id"`
	Name       string         `json:"name"`
	Address    *string        `json:"address,omitempty"`
	City       *string        `json:"city,omitempty"`
	State      *string        `json:"state,omitempty"`
	Timezone   *time.Location `json:"timezone,omitempty"`
	Rating     *float32       `json:"rating,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}
