package resy

import (
	"errors"
	"os"
	"testing"
)

// bookingTestsEnabled controls whether dangerous booking tests are run
var bookingTestsEnabled = os.Getenv("RESY_ENABLE_BOOKING_TESTS") == "true"

// skipIfBookingTestsDisabled skips the test if booking tests are not explicitly enabled
func skipIfBookingTestsDisabled(t *testing.T) {
	t.Helper()
	if !bookingTestsEnabled {
		t.Skip("Booking tests disabled. Set RESY_ENABLE_BOOKING_TESTS=true to enable.")
	}
}

func TestBooking_Book_WithoutPayment(t *testing.T) {
	skipIfBookingTestsDisabled(t)

	t.Log("WARNING: This test may create a real reservation!")
	t.Log("It will attempt to book without a payment method and should fail for most restaurants")

	client, config := newAuthenticatedClient(t)

	// First, get available slots to get a booking token
	slots, venue, err := client.GetSlots(config.VenueId, config.TestDate, 2)
	if err != nil || len(slots) == 0 {
		t.Skip("No slots available for booking test")
	}

	// Get slot details to create a booking token
	slotConfig := slots[0].Config.Token
	slotDetails, err := client.GetSlotDetails(slotConfig, config.TestDate, 2)
	if err != nil {
		t.Skipf("Could not get slot details: %v", err)
	}

	bookingToken := slotDetails.BookingToken.Value

	t.Logf("Attempting to book at %s without payment method", venue.Name)

	// Attempt booking without payment method
	confirmation, err := client.BookReservation(bookingToken, nil)

	// Two possible outcomes:
	// 1. ErrPaymentRequired if restaurant requires payment
	// 2. Success if restaurant doesn't require payment (then we must cancel)

	if err != nil {
		if errors.Is(err, ErrPaymentRequired) {
			t.Log("Got ErrPaymentRequired as expected for restaurant requiring payment")
			return
		}
		// Other error
		t.Logf("Got error: %v", err)
		return
	}

	// Booking succeeded - MUST CANCEL IMMEDIATELY
	if confirmation != nil {
		t.Logf("WARNING: Booking succeeded! Reservation ID: %s, Token: %s",
			confirmation.ReservationId, confirmation.ReservationToken)
		t.Log("Attempting to cancel immediately...")

		cancelErr := client.CancelBooking(confirmation.ReservationToken, nil)
		if cancelErr != nil {
			t.Fatalf("CRITICAL: Failed to cancel booking! Manual cancellation required. Error: %v", cancelErr)
		}

		t.Log("Successfully cancelled the test booking")
	}
}

func TestBooking_Book_WithPayment(t *testing.T) {
	skipIfBookingTestsDisabled(t)

	t.Skip("VERY DANGEROUS: This test creates real reservations with payment. Skipped by default.")

	// This test is intentionally not implemented to prevent accidental real bookings
	// If you need to test this functionality:
	// 1. Ensure you have a test reservation that is safe to book
	// 2. Ensure you have a test payment method
	// 3. Implement with extreme caution
	// 4. Always cancel immediately after booking
}

func TestBooking_Cancel(t *testing.T) {
	skipIfBookingTestsDisabled(t)

	client, _ := newAuthenticatedClient(t)

	// Check if a test cancellable reservation token is provided
	testToken := os.Getenv("RESY_TEST_CANCELLABLE_RESERVATION")
	if testToken == "" {
		t.Skip("No cancellable reservation token provided. Set RESY_TEST_CANCELLABLE_RESERVATION to enable this test.")
	}

	t.Logf("WARNING: Cancelling reservation with token: %s", testToken)

	// Cancel the booking
	var responseBody []byte
	err := client.CancelBooking(testToken, &responseBody)
	requireNoError(t, err, "CancelBooking failed")

	t.Logf("Successfully cancelled booking. Response: %s", string(responseBody))
}

func TestBooking_Cancel_InvalidToken(t *testing.T) {
	client, _ := newAuthenticatedClient(t)

	// Use an invalid reservation token
	err := client.CancelBooking("invalid-token-12345", nil)

	// Should get an error (likely 404)
	if err == nil {
		t.Error("expected error for invalid reservation token")
	}

	if errors.Is(err, ErrUnauthorized) {
		t.Logf("Got error: %v (expected)", err)
	} else {
		t.Errorf("Got unexpected error: %v", err)
	}
}

func TestBooking_Book_InvalidToken(t *testing.T) {
	client, _ := newAuthenticatedClient(t)

	// Use an invalid booking token
	confirmation, err := client.BookReservation("invalid-token-12345", nil)

	// Should get an error
	if err == nil {
		t.Error("expected error for invalid booking token")
	}

	if !errors.Is(err, ErrBadRequest) && !errors.Is(err, ErrNotFound) && !errors.Is(err, ErrBadGateway) {
		t.Logf("Got error: %v (acceptable)", err)
	}

	if confirmation != nil {
		t.Errorf("expected nil confirmation on error, got: %v", confirmation)
	}
}
