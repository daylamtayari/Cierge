package reservation

import (
	"context"
	"errors"
)

var (
	ErrUnsupportedPlatform = errors.New("unsupported platform")
)

type BookingClient interface {
	// Handles any pre-booking checks such as token validity
	PreBookingCheck(ctx context.Context, event Event) error

	// Retrieve matching slots
	// The slice of any represents slot types and should be
	// type asserted in the respective methods
	FetchSlot(ctx context.Context, event Event) (any, error)

	// Attempts to perform a booking
	Book(ctx context.Context, event Event, slots any) (*BookingResult, error)
}

// Returns a new booking client for the specified platform
// NOTE: Add OpenTable client when created
func NewBookingClient(platform string, token string) (BookingClient, error) {
	switch platform {
	case "resy":
		return NewResyClient(token)
	default:
		return nil, ErrUnsupportedPlatform
	}
}
