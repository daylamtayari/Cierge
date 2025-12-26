package models

import (
	"time"

	"github.com/google/uuid"
)

type UserNotificationPreferences struct {
	ID     uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`

	// Notification methods
	Email      *string `gorm:"type:varchar(255)"`
	Phone      *string `gorm:"type:varchar(15)"`
	WebhookURL *string `gorm:"type:varchar(2048)"`

	// Email notification toggles
	EmailTokenExpiry bool `gorm:"not null;default:true"`
	EmailJobStarted  bool `gorm:"not null;default:false"`
	EmailJobSuccess  bool `gorm:"not null;default:true"`
	EmailJobFailed   bool `gorm:"not null;default:true"`

	// SMS notification toggles
	SMSTokenExpiry bool `gorm:"not null;default:true"`
	SMSJobStarted  bool `gorm:"not null;default:false"`
	SMSJobSuccess  bool `gorm:"not null;default:true"`
	SMSJobFailed   bool `gorm:"not null;default:true"`

	// Webhook notification toggles
	WebhookTokenExpiry bool `gorm:"not null;default:false"`
	WebhookJobStarted  bool `gorm:"not null;default:false"`
	WebhookJobSuccess  bool `gorm:"not null;default:true"`
	WebhookJobFailed   bool `gorm:"not null;default:true"`

	CreatedAt time.Time `gorm:"not null;default:now()"`
	UpdatedAt time.Time `gorm:"not null;default:now()"`
}

// Checks if a user wants a notification type via a delivery method
func (p *UserNotificationPreferences) WantsEmail(notifType string) bool {
	if p.Email == nil || *p.Email == "" {
		return false
	}
	switch notifType {
	case "token_expiry":
		return p.EmailTokenExpiry
	case "job_started":
		return p.EmailJobStarted
	case "job_success":
		return p.EmailJobSuccess
	case "job_failed":
		return p.EmailJobFailed
	default:
		return false
	}
}

func (p *UserNotificationPreferences) WantsSMS(notifType string) bool {
	if p.Phone == nil || *p.Phone == "" {
		return false
	}
	switch notifType {
	case "token_expiry":
		return p.SMSTokenExpiry
	case "job_started":
		return p.SMSJobStarted
	case "job_success":
		return p.SMSJobSuccess
	case "job_failed":
		return p.SMSJobFailed
	default:
		return false
	}
}

func (p *UserNotificationPreferences) WantsWebhook(notifType string) bool {
	if p.WebhookURL == nil || *p.WebhookURL == "" {
		return false
	}
	switch notifType {
	case "token_expiry":
		return p.WebhookTokenExpiry
	case "job_started":
		return p.WebhookJobStarted
	case "job_success":
		return p.WebhookJobSuccess
	case "job_failed":
		return p.WebhookJobFailed
	default:
		return false
	}
}
