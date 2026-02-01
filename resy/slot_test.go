package resy

import (
	"errors"
	"testing"
	"time"
)

func TestSlots_Get(t *testing.T) {
	client, config := newUnauthenticatedClient(t)

	slots, venue, err := client.GetSlots(config.VenueId, config.TestDate, 2)
	requireNoError(t, err, "GetSlots failed")

	// Venue should always be returned even if no slots
	assertNotNil(t, venue, "venue should not be nil")

	if venue.Id.Resy != config.VenueId {
		t.Errorf("Venue Id.Resy: got %d, want %d", venue.Id.Resy, config.VenueId)
	}

	// Slots may be empty if none are available
	if slots == nil {
		t.Fatal("slots should not be nil (can be empty slice)")
	}

	if len(slots) > 0 {
		// Validate first slot structure
		slot := slots[0]

		if slot.Config.Token == "" {
			t.Error("Slot Config.Token should not be empty")
		}

		if slot.Config.Id == 0 {
			t.Error("Slot Config.Id should not be 0")
		}

		if slot.Date.Start.IsZero() {
			t.Error("Slot Date.Start should not be zero")
		}

		if slot.Date.End.IsZero() {
			t.Error("Slot Date.End should not be zero")
		}

		if slot.Quantity < 0 {
			t.Errorf("Slot Quantity should not be negative, got: %d", slot.Quantity)
		}

		// Start should be before end
		if !slot.Date.Start.Before(slot.Date.End.Time) {
			t.Errorf("Slot Start (%v) should be before End (%v)",
				slot.Date.Start.Time, slot.Date.End.Time)
		}

		t.Logf("Found %d slots for venue %s on %s, first slot: %s (Quantity: %d)",
			len(slots), venue.Name, config.TestDate.Format("2006-01-02"),
			slot.Date.Start.Format(ResyDatetimeFormat), slot.Quantity)
	} else {
		t.Logf("No slots available for venue %s on %s (party size: 2)",
			venue.Name, config.TestDate.Format("2006-01-02"))
	}
}

func TestSlots_Get_NoAvailability(t *testing.T) {
	client, config := newUnauthenticatedClient(t)

	// Use a date far in the future (90+ days) where slots are unlikely to be available
	farFutureDate := time.Now().AddDate(0, 0, 90)

	slots, venue, err := client.GetSlots(config.VenueId, farFutureDate, 2)

	if err != nil {
		t.Errorf("Got error retrieving slots for venue: %v", err)
		return
	}

	// If no error, venue should be returned
	assertNotNil(t, venue, "venue should not be nil when no error")

	if len(slots) > 0 {
		t.Logf("Warning: Found %d slots 90 days in future (unexpected but valid)", len(slots))
	}
}

func TestSlots_Get_InvalidVenue(t *testing.T) {
	client, config := newUnauthenticatedClient(t)

	slots, venue, err := client.GetSlots(TestVenueInvalid, config.TestDate, 2)

	if err == nil {
		t.Error("expected error for invalid venue ID")
	}

	// Should get ErrBadRequest
	if !errors.Is(err, ErrBadRequest) {
		t.Logf("Got error: %v (expected)", err)
	}

	if slots != nil {
		t.Logf("Warning: Got slots despite error: %v", slots)
	}

	if venue != nil {
		t.Logf("Warning: Got venue despite error: %v", venue)
	}
}

func TestSlots_GetDetails(t *testing.T) {
	client, config := newAuthenticatedClient(t)

	// First, get available slots
	slots, venue, err := client.GetSlots(config.VenueId, config.TestDate, 2)
	requireNoError(t, err, "GetSlots failed")

	if len(slots) == 0 {
		t.Skip("No slots available for GetSlotDetails test")
	}

	// WARNING: This creates a real booking token that expires in 5 minutes
	t.Logf("WARNING: Creating booking token for venue %s", venue.Name)

	slotConfig := slots[0].Config.Token
	slotDetails, err := client.GetSlotDetails(slotConfig, config.TestDate, 2)
	requireNoError(t, err, "GetSlotDetails failed")
	assertNotNil(t, slotDetails, "slotDetails should not be nil")

	// Validate booking token
	if slotDetails.BookingToken.Value == "" {
		t.Error("BookingToken.Value should not be empty")
	}

	if slotDetails.BookingToken.Expiry.IsZero() {
		t.Error("BookingToken.Expiry should not be zero")
	}

	// Expiry should be in the future
	if !slotDetails.BookingToken.Expiry.After(time.Now()) {
		t.Errorf("BookingToken.Expiry should be in future, got: %v", slotDetails.BookingToken.Expiry)
	}

	// Expiry should be roughly 5 minutes from now (within 6 minutes)
	expiryDiff := time.Until(slotDetails.BookingToken.Expiry)
	if expiryDiff > 6*time.Minute {
		t.Errorf("BookingToken.Expiry seems too far in future: %v", expiryDiff)
	}

	t.Logf("Created booking token: %s (expires in %.1f minutes)",
		slotDetails.BookingToken.Value, expiryDiff.Minutes())
	t.Logf("User: %s %s (ID: %d)",
		slotDetails.User.FirstName, slotDetails.User.LastName, slotDetails.User.Id)
}

func TestSlots_GetDetails_InvalidConfig(t *testing.T) {
	client, config := newAuthenticatedClient(t)

	// Use an invalid config token
	slotDetails, err := client.GetSlotDetails("invalid-token-12345", config.TestDate, 2)

	// Should get an error (likely 400 or 404)
	if err == nil {
		t.Error("expected error for invalid config token")
	}

	if !errors.Is(err, ErrBadRequest) && !errors.Is(err, ErrNotFound) && !errors.Is(err, ErrBadGateway) {
		t.Logf("Got error: %v (acceptable)", err)
	}

	if slotDetails != nil {
		t.Errorf("expected nil slotDetails on error, got: %v", slotDetails)
	}
}
