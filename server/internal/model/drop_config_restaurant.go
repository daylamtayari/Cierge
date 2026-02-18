package model

import (
	"time"

	"github.com/google/uuid"
)

// Junction table between DropConfig and Restaurant
type DropConfigRestaurant struct {
	DropConfigID uuid.UUID `gorm:"type:uuid;primaryKey;index:idx_drop_config_restaurants_restaurant,priority:2"`
	RestaurantID uuid.UUID `gorm:"type:uuid;primaryKey;index:idx_drop_config_restaurants_restaurant,priority:1"`
	Confidence   int16     `gorm:"type:smallint;not null;default:0"`
	CreatedAt    time.Time `gorm:"not null;default:now()"`
}
