package model

import (
	"time"

	"github.com/daylamtayari/cierge/api"
	"github.com/google/uuid"
)

type DropConfig struct {
	ID uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`

	DaysInAdvance int16      `gorm:"type:smallint;not null;uniqueIndex:idx_drop_configs_days_time"`
	DropTime      string     `gorm:"type:varchar(5);not null;uniqueIndex:idx_drop_configs_days_time"` // "HH:mm"
	Confidence    int16      `gorm:"<-:false;-:migration"` // populated from drop_config_restaurants when querying by restaurant
	LastUsedAt    *time.Time `gorm:"type:timestamptz"`

	CreatedBy *uuid.UUID `gorm:"type:uuid"`

	// Relations
	Restaurants []*Restaurant `gorm:"many2many:drop_config_restaurants;"`
	Jobs        []Job         `gorm:"foreignKey:DropConfigID"`

	CreatedAt time.Time `gorm:"not null;default:now()"`
	UpdatedAt time.Time `gorm:"not null;default:now()"`
}

func (m *DropConfig) ToAPI() *api.DropConfig {
	return &api.DropConfig{
		ID:            m.ID,
		DaysInAdvance: m.DaysInAdvance,
		DropTime:      m.DropTime,
		Confidence:    m.Confidence,
		CreatedAt:     m.CreatedAt,
	}
}
