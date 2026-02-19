package reservation

import (
	"time"

	"github.com/google/uuid"
)

// Represents the data that is provided to the handler
// NOTE: Callback signifies whether or not a callback should be made
// NOTE: Drop time must have a UTC timezone
// NOTE: Strict preference represents whether the preference should be
// absolutely respected or not (not recommended for highly competitive reservations)
type Event struct {
	JobID                   uuid.UUID `json:"job_id"`
	Platform                string    `json:"platform"`
	PlatformVenueId         string    `json:"platform_venue_id"`
	EncryptedToken          string    `json:"encrypted_token"`
	EncryptedCallbackSecret string    `json:"encrypted_callback_secret"`
	ReservationDate         string    `json:"reservation_date"` // YYYY-MM-DD
	PartySize               int16     `json:"party_size"`
	PreferredTimes          []string  `json:"preferred_times"` // HH:mm
	DropTime                time.Time `json:"drop_time"`
	ServerEndpoint          string    `json:"server_endpoint"`
	Callback                bool      `json:"callback"`
	StrictPreference        bool      `json:"strict_preference"`
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
	JobId           uuid.UUID     `json:"job_id"`
	Success         bool          `json:"success"`
	Duration        time.Duration `json:"duration"`
	Message         string        `json:"message"`
	Error           string        `json:"error,omitempty"`
	Level           string        `json:"level"`
	StartTime       time.Time     `json:"start_time"`
	BookingStart    time.Time     `json:"booking_start"`
	DriftNs         int64         `json:"drift_ns"`
	BookingAttempts []Attempt     `json:"booking_attempts"`
	BookingResult
}

// Booking attempt
// NOTE: Slot time is in a UTC timezone
type Attempt struct {
	Result    *BookingResult `json:"result"`
	Error     string         `json:"error,omitempty"`
	SlotTime  time.Time      `json:"slot_time"`
	StartTime time.Time      `json:"start_time"`
	Duration  time.Duration  `json:"duration"`
}
