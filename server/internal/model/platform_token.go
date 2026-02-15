package model

import (
	"time"

	"github.com/google/uuid"
)

type PlatformToken struct {
	ID       uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID   uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_platform_tokens_user_platform"`
	Platform string    `gorm:"type:platform;not null;uniqueIndex:idx_platform_tokens_user_platform"`

	EncryptedToken   string     `gorm:"type:text;not null"`
	ExpiresAt        *time.Time `gorm:"type:timestamptz;index:idx_platform_tokens_expiry,where:expires_at IS NOT NULL"`
	HasRefresh       bool       `gorm:"default:false"`
	RefreshExpiresAt *time.Time `gorm:"type:timestamptz"`

	CreatedAt time.Time `gorm:"not null;default:now()"`
	UpdatedAt time.Time `gorm:"not null;default:now()"`
}

func (t *PlatformToken) IsExpired() bool {
	if t.ExpiresAt == nil {
		return false
	}
	return time.Now().UTC().After(*t.ExpiresAt)
}

func (t *PlatformToken) ExpiresIn() time.Duration {
	if t.ExpiresAt == nil {
		return 0
	}
	return time.Until(*t.ExpiresAt)
}
