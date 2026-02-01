package reservation

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrUnsupportedPlatform = errors.New("unsupported platform")
	ErrFailedToBookSlots   = errors.New("failed to book any of the slots")
)

type BookingClient interface {
	// Handles any pre-booking checks such as token validity
	PreBookingCheck(ctx context.Context, event Event) error

	// Retrieve matching slots
	// The slice of any represents slot types and should be
	// type asserted in the respective methods
	FetchSlots(ctx context.Context, event Event) (any, error)

	// Attempts to perform a booking for a single slot
	Book(ctx context.Context, event Event, slot any) (Attempt, error)

	// Books slots using the generic booking handler
	BookSlots(ctx context.Context, event Event, slots any) (*BookingResult, []Attempt, error)
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

// Generic booking handler that handles the core booking logic
// - Implements soft and strict preference
// - Generics to handle any slot type
// The slots slice is to be in preference order
// Returns the result if successful, a slice of all attempts, and an error
func bookingHandler[T any](
	ctx context.Context,
	client BookingClient,
	event Event,
	slots []T,
) (*BookingResult, []Attempt, error) {
	var attempts []Attempt

	// Strict preference: attempt each reservation sequentially
	if event.StrictPreference {
		for _, slot := range slots {
			attempt, err := client.Book(ctx, event, slot)
			attempts = append(attempts, attempt)
			if err == nil {
				return attempt.Result, attempts, nil
			}
		}
		return nil, attempts, ErrFailedToBookSlots
	}

	// Soft preference: concurrent booking with 1-second stagger
	// Create cancellable context (inherits parent cancellation)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	resultCh := make(chan Attempt, len(slots))
	doneCh := make(chan struct{})
	var wg sync.WaitGroup

	var finalResult *BookingResult
	var finalError error

	// Process results concurrently as they arrive
	go func() {
		defer close(doneCh)
		var errors []error
		for attempt := range resultCh {
			attempts = append(attempts, attempt)
			if attempt.Error == "" {
				// Cancel all other goroutines on success
				cancel()
				finalResult = attempt.Result
				return
			}

			errors = append(errors, fmt.Errorf("slot at %s: %s",
				attempt.SlotTime.Format("15:04"), attempt.Error))
		}

		// All attempts failed
		if len(errors) > 0 {
			finalError = fmt.Errorf("%w: %v", ErrFailedToBookSlots, errors)
		} else {
			finalError = ErrFailedToBookSlots
		}
	}()

	// Launch goroutines with 1-second stagger to maintain soft preference
launchLoop:
	for i, slot := range slots {
		// Stagger launches (except first one)
		if i > 0 {
			select {
			case <-time.After(1 * time.Second):
				// Continue to launch next goroutine
			case <-ctx.Done():
				// Context cancelled
				break launchLoop
			}
		}

		// Check if context was cancelled
		select {
		case <-ctx.Done():
			break launchLoop
		default:
			wg.Add(1)
			go bookSlotConcurrent(ctx, &wg, resultCh, client, event, slot)
		}
	}

	// Close channel after all attempts complete
	wg.Wait()
	close(resultCh)

	// Wait for result processing to complete
	<-doneCh
	return finalResult, attempts, finalError
}

// Wraps Book call for goroutine execution with context cancellation support
func bookSlotConcurrent[T any](
	ctx context.Context,
	wg *sync.WaitGroup,
	resultCh chan<- Attempt,
	client BookingClient,
	event Event,
	slot T,
) {
	defer wg.Done()

	// Check if context already cancelled, skip if cancelled
	select {
	case <-ctx.Done():
		return
	default:
	}

	attempt, err := client.Book(ctx, event, slot)

	// Send result back
	if err != nil {
		attempt.Error = err.Error()
	}

	resultCh <- attempt
}
