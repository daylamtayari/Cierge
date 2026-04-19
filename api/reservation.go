package api

import (
	"time"

	"github.com/google/uuid"
)

type Reservation struct {
	ID           uuid.UUID  `json:"id"`
	UserID       uuid.UUID  `json:"user_id"`
	JobID        *uuid.UUID `json:"job_id,omitempty"`
	RestaurantID uuid.UUID  `json:"restaurant_id"`

	Platform      string    `json:"platform"`
	Confirmation  *string   `json:"confirmation,omitempty"`
	ReservationAt time.Time `json:"reservation_at"`
	PartySize     int16     `json:"party_size"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
