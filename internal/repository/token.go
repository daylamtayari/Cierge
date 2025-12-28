package repository

import (
	"time"

	"gorm.io/gorm"
)

type TokenRepository struct {
	db      *gorm.DB
	timeout time.Duration
}

func NewTokenRepository(db *gorm.DB, timeout time.Duration) *TokenRepository {
	return &TokenRepository{
		db:      db,
		timeout: timeout,
	}
}
