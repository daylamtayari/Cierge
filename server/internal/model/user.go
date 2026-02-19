package model

import (
	"time"

	"github.com/daylamtayari/cierge/api"
	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Email        string    `gorm:"type:varchar(255);not null;uniqueIndex"`
	PasswordHash *string   `gorm:"type:varchar(255)"`
	ApiKey       *string   `gorm:"type:varchar(64);uniqueIndex"`

	// User management
	IsAdmin             bool       `gorm:"not null;default:false"`
	PasswordChangedAt   *time.Time `gorm:"type:timestamptz"`
	LastLoginAt         *time.Time `gorm:"type:timestamptz"`
	FailedLoginAttempts int        `gorm:"type:int;not null;default:0"`
	LastFailedLogin     *time.Time `gorm:"type:timestamptz"`
	LockedUntil         *time.Time `gorm:"type:timestamptz"`
	ApiKeyLastCreated   *time.Time `gorm:"type:timestamptz"`

	// OIDC fields
	OIDCProvider *string `gorm:"column:oidc_provider;type:varchar(50);index:idx_users_oidc,where:oidc_provider IS NOT NULL"`
	OIDCSubject  *string `gorm:"column:oidc_subject;type:varchar(255);index:idx_users_oidc,where:oidc_provider IS NOT NULL"`

	// Relations
	NotificationPreferences *NotificationPreferences `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	PlatformTokens          []PlatformToken          `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Favourites              []Favourite              `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Jobs                    []Job                    `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Reservations            []Reservation            `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Notifications           []Notification           `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`

	CreatedAt time.Time `gorm:"not null;default:now()"`
	UpdatedAt time.Time `gorm:"not null;default:now()"`
}

func (u *User) IsAccountLocked() bool {
	if u.LockedUntil == nil {
		return false
	}
	return time.Now().UTC().Before(*u.LockedUntil)
}

func (u *User) NeedsPasswordChange() bool {
	return u.PasswordChangedAt == nil
}

// ToAPI converts the internal User model to the public API representation.
func (u *User) ToAPI() *api.User {
	authMethod := api.LocalAuthMethod
	// TODO: Add check if OIDC or even social auth is enabled for the user

	lastLoginAt := u.LastLoginAt
	if u.LastLoginAt != nil {
		lastLoginAtUTC := u.LastLoginAt.UTC()
		lastLoginAt = &lastLoginAtUTC
	}

	return &api.User{
		ID:          u.ID,
		Email:       u.Email,
		HasApiKey:   u.ApiKey != nil && *u.ApiKey != "",
		IsAdmin:     u.IsAdmin,
		AuthMethod:  authMethod,
		LastLoginAt: lastLoginAt,
		CreatedAt:   u.CreatedAt.UTC(),
	}
}
