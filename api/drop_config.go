package api

import (
	"time"

	"github.com/google/uuid"
)

type DropConfig struct {
	ID           uuid.UUID `json:"id"`
	Platform     string    `json:"platform"`
	RestaurantID uuid.UUID `json:"restaurant_id"`

	DaysInAdvance int16          `json:"days_in_advance"`
	DropTime      string         `json:"drop_time"`
	Timezone      *time.Location `json:"timezone"`
	Confidence    int16          `json:"confidence"`

	CreatedAt time.Time `json:"created_at"`
}
