package model

import (
	"time"

	"github.com/google/uuid"
)

type DropConfig struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Platform     string    `gorm:"type:platform;not null"`
	RestaurantID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_drop_configs_restaurant_days"`

	DaysInAdvance int16      `gorm:"type:smallint;not null;uniqueIndex:idx_drop_configs_restaurant_days"`
	DropTime      string     `gorm:"type:time;not null"` // "15:04"
	Timezone      *Timezone  `gorm:"type:varchar(64)"`
	Confidence    int16      `gorm:"type:smallint;not null;default:0"`
	LastUsedAt    *time.Time `gorm:"type:timestamptz"`

	CreatedBy *uuid.UUID `gorm:"type:uuid"`

	// Relations
	Restaurant *Restaurant `gorm:"foreignKey:RestaurantID"`
	Jobs       []Job       `gorm:"foreignKey:DropConfigID"`

	CreatedAt time.Time `gorm:"not null;default:now()"`
	UpdatedAt time.Time `gorm:"not null;default:now()"`
}
