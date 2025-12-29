package repository

import (
	"time"

	"gorm.io/gorm"
)

type PlatformTokenRepository struct {
	db      *gorm.DB
	timeout time.Duration
}

func NewPlatformTokenRepository(db *gorm.DB, timeout time.Duration) *PlatformTokenRepository {
	return &PlatformTokenRepository{
		db:      db,
		timeout: timeout,
	}
}
