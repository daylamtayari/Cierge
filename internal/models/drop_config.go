package models

import (
	"time"

	"github.com/google/uuid"
)

type DropConfig struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Platform     Platform  `gorm:"type:varchar(50);not null;index:idx_drop_config_platform"`
	RestaurantID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_drop_configs_restaurant_days"`

	DaysInAdvance  int16      `gorm:"type:smallint;not null;uniqueIndex:idx_drop_configs_restaurant_days"`
	DropTime       string     `gorm:"type:time;not null"` // "09:00:00"
	Timezone       string     `gorm:"type:varchar(50);not null"`
	Confidence     int16      `gorm:"type:smallint;not null;default:0"`
	LastVerifiedAt *time.Time `gorm:"type:timestamptz"`

	CreatedBy     *uuid.UUID `gorm:"type:uuid"`
	LastUpdatedBy *uuid.UUID `gorm:"type:uuid"`

	// Relations
	Restaurant *Restaurant `gorm:"foreignKey:RestaurantID"`
	Jobs       []Job       `gorm:"foreignKey:DropConfigID"`

	CreatedAt time.Time `gorm:"not null;default:now()"`
	UpdatedAt time.Time `gorm:"not null;default:now()"`
}
