package models

import (
	"time"

	"github.com/google/uuid"
)

type NotificationType string

const (
	NotificationTypeTokenExpiry NotificationType = "token_expiry"
	NotificationTypeJobStarted  NotificationType = "job_started"
	NotificationTypeJobSuccess  NotificationType = "job_success"
	NotificationTypeJobFailed   NotificationType = "job_failed"
)

type Notification struct {
	ID     uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID uuid.UUID `gorm:"type:uuid;not null;index:idx_notifications_user"`

	Type    NotificationType `gorm:"type:notification_type;not null"`
	Title   string           `gorm:"type:varchar(255);not null"`
	Message string           `gorm:"type:text;not null"`

	JobID *uuid.UUID `gorm:"type:uuid"`
	Job   *Job       `gorm:"foreignKey:JobID"`

	ReadAt *time.Time `gorm:"index:idx_notifications_user_unread,where:read_at IS NULL"`

	CreatedAt time.Time `gorm:"not null;default:now()"`
}
