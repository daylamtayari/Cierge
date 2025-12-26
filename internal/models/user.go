package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Email        string    `gorm:"type:varchar(255);not null;uniqueIndex"`
	PasswordHash *string   `gorm:"type:varchar(255)"`

	// OIDC fields
	OIDCProvider *string `gorm:"type:varchar(50);index:idx_users_oidc,where:oidc_provider IS NOT NULL"`
	OIDCSubject  *string `gorm:"type:varchar(255);index:idx_users_oidc,where:oidc_provider IS NOT NULL"`

	// Relations
	NotificationPreferences *UserNotificationPreferences `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	PlatformTokens          []PlatformToken              `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Favourites              []Favourite                  `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Jobs                    []Job                        `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Reservations            []Reservation                `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Notifications           []Notification               `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`

	CreatedAt time.Time `gorm:"not null;default:now()"`
	UpdatedAt time.Time `gorm:"not null;default:now()"`
}
