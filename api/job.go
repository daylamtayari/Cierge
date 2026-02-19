package api

import (
	"net/http"
	"strconv"
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

	ReservationDate string   `json:"reservation_date"` // YYYY-MM-DD
	PartySize       int16    `json:"party_size"`
	PreferredTimes  []string `json:"preferred_times"` // HH:mm

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

// Request type for a new job
type JobCreationRequest struct {
	RestaurantID    uuid.UUID `json:"restaurant_id"`
	ReservationDate string    `json:"reservation_date"` // YYYY-MM-DD
	PartySize       int16     `json:"party_size"`
	PreferredTimes  []string  `json:"preferred_times"` // HH:mm
	DropConfigID    uuid.UUID `json:"drop_config_id"`
}

// Retrieve jobs for the user
// If upcomingOnly is set to true, only upcoming jobs are returned
func (c *Client) GetJobs(upcomingOnly bool) ([]Job, error) {
	reqUrl := c.host + "/api/job/list?upcoming=" + strconv.FormatBool(upcomingOnly)
	req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
	if err != nil {
		return nil, err
	}

	var jobs []Job
	err = c.Do(req, &jobs)
	if err != nil {
		return nil, err
	}

	return jobs, nil
}
