package model

import (
	"time"

	"github.com/google/uuid"
)

type Revocation struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
	JTI       string    `gorm:"column:jti;type:varchar(255);not null;uniqueIndex"`
	RevokedAt time.Time `gorm:"type:timestamptz;default:now()"`
	RevokedBy string    `gorm:"type:varchar(255)"`
}
