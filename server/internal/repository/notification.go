package repository

import (
	"time"

	"gorm.io/gorm"
)

type Notification struct {
	db      *gorm.DB
	timeout time.Duration
}

func NewNotification(db *gorm.DB, timeout time.Duration) *Notification {
	return &Notification{
		db:      db,
		timeout: timeout,
	}
}
