package models

import (
	"time"

	"github.com/google/uuid"
)

type Restaurant struct {
	ID                   uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Platform             Platform  `gorm:"type:varchar(50);not null;uniqueIndex:idx_restaurants_platform_id"`
	PlatformRestaurantID string    `gorm:"type:varchar(255);not null;uniqueIndex:idx_restaurants_platform_id"`

	Name    string   `gorm:"type:varchar(255);not null;index"`
	Address *string  `gorm:"type:varchar(1024)"`
	City    *string  `gorm:"type:varchar(255);index"`
	State   *string  `gorm:"type:varchar(100)"`
	Rating  *float32 `gorm:"type:real"`

	// Relations
	DropConfigs  []DropConfig  `gorm:"foreignKey:RestaurantID;constraint:OnDelete:CASCADE"`
	Favourites   []Favourite   `gorm:"foreignKey:RestaurantID;constraint:OnDelete:CASCADE"`
	Jobs         []Job         `gorm:"foreignKey:RestaurantID;constraint:OnDelete:RESTRICT"`
	Reservations []Reservation `gorm:"foreignKey:RestaurantID;constraint:OnDelete:RESTRICT"`

	CreatedAt time.Time `gorm:"not null;default:now()"`
	UpdatedAt time.Time `gorm:"not null;default:now()"`
}
