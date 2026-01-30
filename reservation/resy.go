package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/daylamtayari/cierge/resy"
)

var (
	ErrNoMatchingSlotsFound = errors.New("no slots matching the preferred times found")
	ErrNoSlotsFound         = errors.New("no reservation slots found")
	ErrFailedToBookSlots    = errors.New("failed to book any of the slots")
	ErrUnmarshalToken       = errors.New("failed to unmarshal token")
)

type ResyClient struct {
	client *resy.Client
	tokens resy.Tokens
}

// Returns a Resy booking client
func NewResyClient(token string) (*ResyClient, error) {
	resyClient := ResyClient{}
	// Unmarshal token string into resy.Tokens
	err := json.Unmarshal([]byte(token), &resyClient.tokens)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUnmarshalToken, err)
	}

	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	resyClient.client = resy.NewClient(httpClient, resyClient.tokens, "")

	return &resyClient, nil
}

// Performs pre-booking checks for the Resy client
// - Test if the tokens are valid
func (c *ResyClient) PreBookingCheck(ctx context.Context, event LambdaEvent) error {
	// Test token validity by retrieving the current user
	_, err := c.client.GetUser()
	if err != nil {
		return err
	}
	return nil
}

// Handles the booking logic for Resy
// - Retrieve slots
// - Filter slots to matching slots
// - Attempt to book slots in order of preference
func (c *ResyClient) Book(ctx context.Context, event LambdaEvent) (*BookingResult, error) {
	venueId, err := strconv.Atoi(event.PlatformVenueId)
	if err != nil {
		return nil, err
	}

	// Get slots
	slotDeadline := 30 * time.Second
	slots, err := c.getSlotsUntilDeadline(ctx, event, venueId, slotDeadline)
	if err != nil {
		return nil, err
	}

	// Find matching slots and sort them in order of preference
	matchingSlots := matchSlots(slots, event.PreferredTimes)
	if len(matchingSlots) == 0 {
		return nil, ErrNoMatchingSlotsFound
	}

	var bookingResult *BookingResult

	// Attempt to book slots in order of preference
	// Ignore 404 errors as that can be simply due to
	// the reservation no longer being available
	for _, slot := range matchingSlots {
		bookingResult, err = c.bookSlot(slot, event.PartySize)
		if err != nil && !errors.Is(err, resy.ErrNotFound) {
			return nil, err
		}
		if err == nil {
			// If no errors are returned, the reservation
			// was successfully booked
			return bookingResult, nil
		}
	}

	return nil, ErrFailedToBookSlots
}

// Books a given slot for a given party size
// Returns a BookingResult if successful or an error if not
// If an ErrNotFound is returned, that is due to the slot no longer being available
func (c *ResyClient) bookSlot(slot resy.Slot, partySize int) (*BookingResult, error) {
	// Get the slot details to get the booking token
	slotDetails, err := c.client.GetSlotDetails(slot.Config.Token, slot.Date.Start.Time, partySize)
	if err != nil {
		return nil, err
	}

	var bookingConfirmation *resy.BookingConfirmation

	// If no payment methods are configured, book without
	// This will work fine if the restaurant does not require a deposit
	// but if it does, a resy.ErrPaymentRequired error will be returned
	if len(slotDetails.User.PaymentMethods) == 0 {
		bookingConfirmation, err = c.client.BookReservation(slotDetails.BookingToken.Value, nil)
	} else {
		paymentMethodId := strconv.Itoa(slotDetails.User.PaymentMethods[0].Id)
		bookingConfirmation, err = c.client.BookReservation(slotDetails.BookingToken.Value, &paymentMethodId)
	}
	if err != nil {
		return nil, err
	}

	return &BookingResult{
		ReservationTime: slot.Date.Start.Time,
		PlatformConfirmation: map[string]any{
			"resy_token":     bookingConfirmation.ReservationToken,
			"reservation_id": bookingConfirmation.ReservationId,
			"venue_opt_in":   bookingConfirmation.VenueOptIn,
		},
	}, nil
}

// Retrieves slots with a 0.05s pause between requests until either slots are found or the deadline after the drop time is expired
// This is to handle if there is a slight delay in the API in marking slots as available after the drop time and ensuring this does
// not cause the lambda to fail
func (c *ResyClient) getSlotsUntilDeadline(ctx context.Context, event LambdaEvent, venueId int, deadline time.Duration) ([]resy.Slot, error) {
	deadlineTime := event.DropTime.Add(deadline)
	pauseDuration := 50 * time.Millisecond // 0.05s

	for {
		if time.Now().UTC().After(deadlineTime) {
			return nil, ErrNoSlotsFound
		}

		// Handle context cancellation
		// (used by Lambda)
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		slots, _, err := c.client.GetSlots(venueId, event.ReservationDate, event.PartySize)
		if err != nil {
			// Exit if an error is returned in the request
			return nil, err
		}

		if len(slots) > 0 {
			return slots, nil
		}

		time.Sleep(pauseDuration)
	}
}

// Accepts a slice of slots representing valid reservation slots, and an ordered slice of
// time objects that represent the preferred slot times in order of preference.
// Returns a slice of matching slots in order of preference
func matchSlots(slots []resy.Slot, preferredTimes []time.Time) []resy.Slot {
	slotsByTime := make(map[[3]int]resy.Slot)
	for _, slot := range slots {
		h, m, s := slot.Date.Start.Clock()
		key := [3]int{h, m, s}
		slotsByTime[key] = slot
	}

	matchingSlots := make([]resy.Slot, 0)

	for _, preferredTime := range preferredTimes {
		h, m, s := preferredTime.Clock()
		key := [3]int{h, m, s}

		if matchingSlot, exists := slotsByTime[key]; exists {
			matchingSlots = append(matchingSlots, matchingSlot)
		}
	}

	return matchingSlots
}
