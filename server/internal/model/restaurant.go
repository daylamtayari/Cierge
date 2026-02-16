package model

import (
	"time"

	"github.com/daylamtayari/cierge/api"
	"github.com/google/uuid"
)

type Restaurant struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Platform   string    `gorm:"type:platform;not null;uniqueIndex:idx_restaurants_platform_id"`
	PlatformID string    `gorm:"type:varchar(255);not null;uniqueIndex:idx_restaurants_platform_id"`

	Name     string    `gorm:"type:varchar(255);not null;index"`
	Address  *string   `gorm:"type:varchar(1024)"`
	City     *string   `gorm:"type:varchar(255);index"`
	State    *string   `gorm:"type:varchar(100)"`
	Timezone *Timezone `gorm:"type:varchar(64)"`

	Rating *float32 `gorm:"type:real"`

	// Relations
	DropConfigs  []DropConfig  `gorm:"foreignKey:RestaurantID;constraint:OnDelete:CASCADE"`
	Favourites   []Favourite   `gorm:"foreignKey:RestaurantID;constraint:OnDelete:CASCADE"`
	Jobs         []Job         `gorm:"foreignKey:RestaurantID;constraint:OnDelete:RESTRICT"`
	Reservations []Reservation `gorm:"foreignKey:RestaurantID;constraint:OnDelete:RESTRICT"`

	CreatedAt time.Time `gorm:"not null;default:now()"`
	UpdatedAt time.Time `gorm:"not null;default:now()"`
}

func (m *Restaurant) ToAPI() *api.Restaurant {
	return &api.Restaurant{
		ID:         m.ID,
		Platform:   m.Platform,
		PlatformID: m.PlatformID,
		Name:       m.Name,
		Address:    m.Address,
		City:       m.City,
		State:      m.State,
		Timezone:   m.Timezone.Location,
		Rating:     m.Rating,
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
	}
}
