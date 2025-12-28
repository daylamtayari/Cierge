package models

import (
	"time"

	"github.com/google/uuid"
)

type TokenType string

const (
	TokenTypeCookie = "Cookie"
	TokenTypeApiKey = "api_key"
)

type PlatformToken struct {
	ID       uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID   uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_platform_tokens_user_platform"`
	Platform Platform  `gorm:"type:varchar(50);not null;uniqueIndex:idx_platform_tokens_user_platform"`

	EncryptedToken string     `gorm:"type:text;not null"`
	ExpiresAt      *time.Time `gorm:"index:idx_platform_tokens_expiry,where:expires_at IS NOT NULL"`
	TokenType      TokenType  `gorm:"type:varchar(50);not null"`

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
