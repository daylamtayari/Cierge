package models

import (
	"time"

	"github.com/google/uuid"
)

type Favourite struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID       uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_favourites_user_restaurant;index:idx_favourites_user"`
	RestaurantID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_favourites_user_restaurant;index:idx_favourites_restaurant"`

	// Relations
	User       *User       `gorm:"foreignKey:UserID"`
	Restaurant *Restaurant `gorm:"foreignKey:RestaurantID"`

	CreatedAt time.Time `gorm:"not null;default:now()"`
}
