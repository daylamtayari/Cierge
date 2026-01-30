package reservation

import (
	"time"

	"github.com/google/uuid"
)

// Represents the data that is provided to the handler
// NOTE: Reservation date must have UTC timezone
// NOTE: Preferred times must have UTC timezone and in order of preference
// NOTE: Callback signifies whether or not a callback should be made
type Event struct {
	JobID                   *uuid.UUID  `json:"job_id"`
	Platform                string      `json:"platform"`
	PlatformVenueId         string      `json:"platform_venue_id"`
	EncryptedToken          string      `json:"encrypted_token"`
	EncryptedCallbackSecret string      `json:"encrypted_callback_secret"`
	ReservationDate         time.Time   `json:"reservation_date"`
	PartySize               int         `json:"party_size"`
	PreferredTimes          []time.Time `json:"preferred_times"`
	DropTime                time.Time   `json:"drop_time"`
	ServerEndpoint          string      `json:"server_endpoint"`
	Callback                bool        `json:"callback"`
}

// Result of a booking
type BookingResult struct {
	ReservationTime      time.Time      `json:"reservation_time"`
	PlatformConfirmation map[string]any `json:"platform_confirmation"`
}

// Represents the output value of the job
// as well as the log event for the Lambda
// It is sent back to the server at completion
// and logged to stdout
type Output struct {
	JobId        *uuid.UUID    `json:"job_id"`
	Success      bool          `json:"success"`
	Duration     time.Duration `json:"duration"`
	Message      string        `json:"error_message"`
	Error        string        `json:"error,omitempty"`
	Level        string        `json:"level"`
	StartTime    time.Time     `json:"start_time"`
	BookingStart time.Time     `json:"booking_start"`
	DriftNs      int64         `json:"drift_ns"`
	BookingResult
}
