package api

import (
	"time"

	"github.com/google/uuid"
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
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	RestaurantID uuid.UUID `json:"restaurant_id"`
	Platform     string    `json:"platform"`

	ReservationDate time.Time `json:"reservation_date"`
	PartySize       int16     `json:"party_size"`
	PreferredTimes  []string  `json:"preferred_times"`

	ScheduledAt  time.Time  `json:"scheduled_at"`
	DropConfigID *uuid.UUID `json:"drop_config_id,omitempty"`
	Callbacked   bool       `json:"callbacked"`

	Status      JobStatus  `json:"status"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	ReservedTime *time.Time `json:"reserved_time,omitempty"`
	Confirmation *string    `json:"confirmation,omitempty"`
	ErrorMessage *string    `json:"error_message,omitempty"`
	Logs         *string    `json:"logs,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
