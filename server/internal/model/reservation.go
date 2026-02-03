package model

import (
	"time"

	"github.com/google/uuid"
)

type Reservation struct {
	ID           uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID       uuid.UUID  `gorm:"type:uuid;not null;index:idx_reservations_user;index:idx_reservations_user_at"`
	JobID        *uuid.UUID `gorm:"type:uuid"`
	RestaurantID uuid.UUID  `gorm:"type:uuid;not null;index:idx_reservations_restaurant"`

	Platform      string    `gorm:"type:platform;not null"`
	Confirmation  *string   `gorm:"type:varchar(255)"`
	ReservationAt time.Time `gorm:"type:timestamptz;not null;index:idx_reservations_user_at"`
	PartySize     int16     `gorm:"type:smallint;not null"`

	// Relations
	User       *User       `gorm:"foreignKey:UserID"`
	Job        *Job        `gorm:"foreignKey:JobID"`
	Restaurant *Restaurant `gorm:"foreignKey:RestaurantID"`

	CreatedAt time.Time `gorm:"not null;default:now()"`
	UpdatedAt time.Time `gorm:"not null;default:now()"`
}
