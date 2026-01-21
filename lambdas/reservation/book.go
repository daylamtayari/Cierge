package main

import (
	"context"
	"errors"
)

var (
	ErrUnsupportedPlatform = errors.New("unsupported platform")
)

type BookingClient interface {
	// Handles any pre-booking checks such as token validaty
	PreBookingCheck(ctx context.Context, event LambdaEvent) error

	// Attempts to perform a booking
	Book(ctx context.Context, event LambdaEvent) (*BookingResult, error)
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
