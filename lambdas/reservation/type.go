package main

import (
	"time"

	"github.com/google/uuid"
)

// Represents the data that is provided to the lambda
// NOTE: Reservation date must have UTC timezone
// NOTE: Preferred times must have UTC timezone and in order of preference
type LambdaEvent struct {
	JobID           *uuid.UUID  `json:"job_id"`
	Platform        string      `json:"platform"`
	PlatformVenueId string      `json:"platform_venue_id"`
	EncryptedToken  string      `json:"encrypted_token"`
	ReservationDate time.Time   `json:"reservation_date"`
	PartySize       int         `json:"party_size"`
	PreferredTimes  []time.Time `json:"preferred_times"`
	DropTime        time.Time   `json:"drop_time"`
	ServerEndpoint  string      `json:"server_endpoint"`
}

// Result of a booking
type BookingResult struct {
	Success              bool           `json:"success"`
	ReservationTime      time.Time      `json:"reservation_time"`
	PlatformConfirmation map[string]any `json:"platform_confirmation"`
}

type Status string

const (
	StatusSuccess = Status("success")
	StatusFail    = Status("fail")
)

// Output value of the job that is sent back to the
// server and set as the output
type JobOutput struct {
	JobId        *uuid.UUID     `json:"job_id"`
	Status       Status         `json:"status"`
	Result       *BookingResult `json:"result"`
	Duration     time.Duration  `json:"duration"`
	ErrorMessage string         `json:"error_message"`
	Log          map[string]any `json:"log"`
}
