package model

import (
	"time"

	"github.com/daylamtayari/cierge/api"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type JobStatus string

const (
	JobStatusCreated   JobStatus = "created"
	JobStatusScheduled JobStatus = "scheduled"
	JobStatusSuccess   JobStatus = "success"
	JobStatusFailed    JobStatus = "failed"
	JobStatusCancelled JobStatus = "cancelled"
)

type Job struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID       uuid.UUID `gorm:"type:uuid;not null;index:idx_jobs_user;index:idx_jobs_user_status"`
	RestaurantID uuid.UUID `gorm:"type:uuid;not null;index:idx_jobs_restaurant"`
	Platform     string    `gorm:"type:platform;not null;index:idx_jobs_platform"`

	ReservationDate string         `gorm:"type:date;not null"` // YYYY-MM-DD
	PartySize       int16          `gorm:"type:smallint;not null"`
	PreferredTimes  pq.StringArray `gorm:"type:varchar(5)[];not null"` // HH:mm

	ScheduledAt        time.Time `gorm:"not null;index:idx_jobs_scheduled,where:status = 'scheduled'"`
	DropConfigID       uuid.UUID `gorm:"type:uuid"`
	CallbackSecretHash *string   `gorm:"type:varchar(255)"`
	Callbacked         bool      `gorm:"not null;default:false"`

	Status      JobStatus  `gorm:"type:job_status;not null;default:'scheduled';index:idx_jobs_status;index:idx_jobs_user_status"`
	StartedAt   *time.Time `gorm:"type:timestamptz"`
	CompletedAt *time.Time `gorm:"type:timestamptz"`

	ReservedTime *time.Time `gorm:"type:timestamptz"`
	Confirmation *string    `gorm:"type:text"`
	ErrorMessage *string    `gorm:"type:text"`
	Logs         *string    `gorm:"type:text"`

	// Relations
	User       *User       `gorm:"foreignKey:UserID"`
	Restaurant *Restaurant `gorm:"foreignKey:RestaurantID"`

	CreatedAt time.Time `gorm:"not null;default:now()"`
	UpdatedAt time.Time `gorm:"not null;default:now()"`
}

func (m *Job) ToAPI() *api.Job {
	return &api.Job{
		ID:           m.ID,
		UserID:       m.UserID,
		RestaurantID: m.RestaurantID,
		Platform:     m.Platform,

		ReservationDate: m.ReservationDate,
		PartySize:       m.PartySize,
		PreferredTimes:  m.PreferredTimes,

		ScheduledAt:  m.ScheduledAt,
		DropConfigID: m.DropConfigID,
		Callbacked:   m.Callbacked,

		Status:      api.JobStatus(m.Status),
		StartedAt:   m.StartedAt,
		CompletedAt: m.CompletedAt,

		ReservedTime: m.ReservedTime,
		Confirmation: m.Confirmation,
		ErrorMessage: m.ErrorMessage,
		Logs:         m.Logs,

		CreatedAt: m.CreatedAt.UTC(),
		UpdatedAt: m.UpdatedAt.UTC(),
	}
}
