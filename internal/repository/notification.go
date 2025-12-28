package repository

import (
	"time"

	"gorm.io/gorm"
)

type NotificationRepository struct {
	db      *gorm.DB
	timeout time.Duration
}

func NewNotificationRepository(db *gorm.DB, timeout time.Duration) *NotificationRepository {
	return &NotificationRepository{
		db:      db,
		timeout: timeout,
	}
}
