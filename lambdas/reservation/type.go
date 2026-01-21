package main

import (
	"time"

	"github.com/daylamtayari/cierge/pkg/resy"
)

// Reservation date is a UTC time
type LambdaEvent struct {
	JobID           string      `json:"job_id"`
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
	Success         bool      `json:"success"`
	ReservationTime time.Time `json:"reservation_time"`
	resy.BookingConfirmation
}

// Output value of the job that is sent back to the
// server and set as the output
type JobOutput struct {
	Status       string        `json:"status"`
	Result       BookingResult `json:"result"`
	ErrorMessage string        `json:"error_message"`
	Logs         string        `json:"logs"`
}
