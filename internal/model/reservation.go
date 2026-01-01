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

	Platform         Platform `gorm:"type:varchar(50);not null"`
	ConfirmationCode *string  `gorm:"type:varchar(255)"`
	ReservationAt    string   `gorm:"type:timestamptz;not null;index:idx_reservation_user_at"`
	PartySize        int16    `gorm:"type:smallint;not null"`

	// Relations
	User       *User       `gorm:"foreignKey:UserID"`
	Job        *Job        `gorm:"foreignKey:JobID"`
	Restaurant *Restaurant `gorm:"foreignKey:RestaurantID"`

	CreatedAt time.Time `gorm:"not null;default:now()"`
	UpdatedAt time.Time `gorm:"not null;default:now()"`
}
