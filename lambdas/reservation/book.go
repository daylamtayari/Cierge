package main

import "context"

type BookingClient interface {
	// Handles any pre-booking checks such as token validaty
	PreBookingCheck(ctx context.Context, event LambdaEvent) error

	// Attempts to perform a booking
	Book(ctx context.Context, event LambdaEvent) (*BookingResult, error)
}
