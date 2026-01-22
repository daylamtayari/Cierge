package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type JobStatus string

const (
	JobStatusScheduled JobStatus = "scheduled"
	JobStatusRunning   JobStatus = "running"
	JobStatusSuccess   JobStatus = "success"
	JobStatusFailed    JobStatus = "failed"
	JobStatusCancelled JobStatus = "cancelled"
)

type Job struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID       uuid.UUID `gorm:"type:uuid;not null;index:idx_jobs_user;index:idx_jobs_user_status"`
	RestaurantID uuid.UUID `gorm:"type:uuid;not null;index:idx_jobs_restaurant"`
	Platform     Platform  `gorm:"type:varchar(50);not null;index:idx_jobs_platform"`

	ReservationDate string         `gorm:"type:date;not null"`
	PartySize       int16          `gorm:"type:smallint;not null"`
	PreferredTimes  pq.StringArray `gorm:"type:varchar(10)[];not null"`

	ScheduledAt        time.Time  `gorm:"not null;index:idx_jobs_scheduled,where:status = 'scheduled'"`
	DropConfigID       *uuid.UUID `gorm:"type:uuid"`
	CallbackSecretHash *string    `gorm:"type:varchar(255)"`
	Callbacked         bool       `gorm:"not null;default:false"`

	Status      JobStatus `gorm:"type:job_status;not null;default:'scheduled';index:idx_jobs_status;index:idx_jobs_user_status"`
	StartedAt   *time.Time
	CompletedAt *time.Time

	ReservedTime     *string `gorm:"type:varchar(10)"`
	ConfirmationCode *string `gorm:"type:varchar(255)"`
	ErrorMessage     *string `gorm:"type:text"`
	Logs             *string `gorm:"type:text"`

	// Relations
	User       *User       `gorm:"foreignKey:UserID"`
	Restaurant *Restaurant `gorm:"foreignKey:RestaurantID"`

	CreatedAt time.Time `gorm:"not null;default:now()"`
	UpdatedAt time.Time `gorm:"not null;default:now()"`
}
