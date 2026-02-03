package repository

import (
	"time"

	"gorm.io/gorm"
)

type DropConfig struct {
	db      *gorm.DB
	timeout time.Duration
}

func NewDropConfig(db *gorm.DB, timeout time.Duration) *DropConfig {
	return &DropConfig{
		db:      db,
		timeout: timeout,
	}
}
