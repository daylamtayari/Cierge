package repository

import (
	"time"

	"gorm.io/gorm"
)

type DropConfigRepository struct {
	db      *gorm.DB
	timeout time.Duration
}

func NewDropConfigRepository(db *gorm.DB, timeout time.Duration) *DropConfigRepository {
	return &DropConfigRepository{
		db:      db,
		timeout: timeout,
	}
}
