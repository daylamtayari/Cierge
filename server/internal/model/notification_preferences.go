package model

import (
	"time"

	"github.com/google/uuid"
)

type NotificationPreferences struct {
	ID     uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`

	Method  string `gorm:"type:varchar(255)"`
	Action  string `gorm:"type:varchar(255)"`
	Enabled bool   `gorm:"not null;default:false"`

	CreatedAt time.Time `gorm:"not null;default:now()"`
	UpdatedAt time.Time `gorm:"not null;default:now()"`
}
