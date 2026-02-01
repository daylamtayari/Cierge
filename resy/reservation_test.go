package resy

import (
	"errors"
	"os"
	"testing"
	"time"
)

func TestReservations_GetAll(t *testing.T) {
	client, _ := newAuthenticatedClient(t)

	reservations, err := client.GetReservations(nil, nil, nil, nil)
	requireNoError(t, err, "GetReservations failed")

	if reservations == nil {
		t.Fatal("reservations should not be nil (can be empty slice)")
	}

	// Validate structure if any reservations exist
	if len(reservations) > 0 {
		res := reservations[0]

		if res.ReservationId == 0 {
			t.Error("ReservationId should not be 0")
		}

		if res.ReservationToken == "" {
			t.Error("ReservationToken should not be empty")
		}

		if res.Venue.VenueId == 0 {
			t.Error("Venue.VenueId should not be 0")
		}

		if res.NumSeats <= 0 {
			t.Errorf("NumSeats should be positive, got: %d", res.NumSeats)
		}

		t.Logf("Found %d total reservations, first: ID=%d, Venue=%d, Seats=%d, When=%s",
			len(reservations), res.ReservationId, res.Venue.VenueId, res.NumSeats,
			res.When.Format(ResyDatetimeFormat))
	} else {
		t.Log("No reservations found (this is acceptable)")
	}
}

func TestReservations_GetUpcoming(t *testing.T) {
	client, _ := newAuthenticatedClient(t)

	resType := UpcomingReservation
	reservations, err := client.GetReservations(&resType, nil, nil, nil)
	requireNoError(t, err, "GetReservations failed")

	if reservations == nil {
		t.Fatal("reservations should not be nil (can be empty slice)")
	}

	now := time.Now()

	// Validate all returned reservations are in the future
	for i, res := range reservations {
		if res.When.Time.Before(now) {
			t.Errorf("Reservation %d: When (%v) should be in future",
				i, res.When.Format(ResyDatetimeFormat))
		}

		// Finished status should be 0 for upcoming
		if res.Status.Finished != 0 {
			t.Logf("Warning: Reservation %d has Finished=%d (expected 0 for upcoming)",
				i, res.Status.Finished)
		}
	}

	t.Logf("Found %d upcoming reservations", len(reservations))
}

func TestReservations_GetPast(t *testing.T) {
	client, _ := newAuthenticatedClient(t)

	resType := PastReservation
	reservations, err := client.GetReservations(&resType, nil, nil, nil)
	requireNoError(t, err, "GetReservations failed")

	if reservations == nil {
		t.Fatal("reservations should not be nil (can be empty slice)")
	}

	now := time.Now()

	// Validate all returned reservations are in the past
	for i, res := range reservations {
		if res.When.Time.After(now) {
			t.Errorf("Reservation %d: When (%v) should be in past",
				i, res.When.Format(ResyDatetimeFormat))
		}

		// Finished status should be 1 for past
		if res.Status.Finished != 1 {
			t.Logf("Warning: Reservation %d has Finished=%d (expected 1 for past)",
				i, res.Status.Finished)
		}
	}

	t.Logf("Found %d past reservations", len(reservations))
}

func TestReservations_GetWithPagination(t *testing.T) {
	client, _ := newAuthenticatedClient(t)

	// First call with limit 5, offset 0
	limit := 5
	offset := 0
	reservations1, err := client.GetReservations(nil, nil, &limit, &offset)
	requireNoError(t, err, "GetReservations (first call) failed")

	if len(reservations1) > limit {
		t.Errorf("First call: expected at most %d reservations, got %d", limit, len(reservations1))
	}

	// Second call with limit 5, offset 5
	offset = 5
	reservations2, err := client.GetReservations(nil, nil, &limit, &offset)
	requireNoError(t, err, "GetReservations (second call) failed")

	if len(reservations2) > limit {
		t.Errorf("Second call: expected at most %d reservations, got %d", limit, len(reservations2))
	}

	// If both calls returned reservations, they should be different
	if len(reservations1) > 0 && len(reservations2) > 0 {
		if reservations1[0].ReservationId == reservations2[0].ReservationId {
			t.Log("Warning: First reservation from both calls is the same (might indicate not enough total reservations)")
		}
	}

	t.Logf("Pagination test: First call returned %d, second call returned %d",
		len(reservations1), len(reservations2))
}

func TestReservations_SetOccasion(t *testing.T) {
	client, _ := newAuthenticatedClient(t)

	// Check if a test reservation token is provided
	testToken := os.Getenv("RESY_TEST_RESERVATION_TOKEN")
	if testToken == "" {
		// Try to get an upcoming reservation
		resType := UpcomingReservation
		limit := 1
		reservations, err := client.GetReservations(&resType, nil, &limit, nil)
		requireNoError(t, err, "GetReservations failed")

		if len(reservations) == 0 {
			t.Skip("No upcoming reservations available for SetReservationOccasion test")
		}

		testToken = reservations[0].ReservationToken
		t.Logf("WARNING: Modifying real reservation ID=%d", reservations[0].ReservationId)
	}

	// Set occasion to Birthday
	err := client.SetReservationOccasion(testToken, BirthdayOccasion)
	requireNoError(t, err, "SetReservationOccasion failed")

	t.Log("Successfully set reservation occasion to Birthday")
}

func TestReservations_SetOccasion_InvalidToken(t *testing.T) {
	client, _ := newAuthenticatedClient(t)

	// Use an invalid reservation token
	err := client.SetReservationOccasion("invalid-token-12345", BirthdayOccasion)

	// Should get an error (likely 404 or 419)
	if err == nil {
		t.Error("expected error for invalid reservation token")
	}

	if !errors.Is(err, ErrNotFound) && !errors.Is(err, ErrUnauthorized) && !errors.Is(err, ErrBadRequest) {
		t.Logf("Got error: %v (acceptable)", err)
	}
}

func TestReservations_SetSpecialRequest(t *testing.T) {
	client, _ := newAuthenticatedClient(t)

	// Check if a test reservation token is provided
	testToken := os.Getenv("RESY_TEST_RESERVATION_TOKEN")
	if testToken == "" {
		// Try to get an upcoming reservation
		resType := UpcomingReservation
		limit := 1
		reservations, err := client.GetReservations(&resType, nil, &limit, nil)
		requireNoError(t, err, "GetReservations failed")

		if len(reservations) == 0 {
			t.Skip("No upcoming reservations available for SetReservationSpecialRequest test")
		}

		testToken = reservations[0].ReservationToken
		t.Logf("WARNING: Modifying real reservation ID=%d", reservations[0].ReservationId)
	}

	// Set a test special request
	specialRequest := "Test special request - please ignore"
	err := client.SetReservationSpecialRequest(testToken, specialRequest)
	requireNoError(t, err, "SetReservationSpecialRequest failed")

	t.Log("Successfully set special request")
}

func TestReservations_SetSpecialRequest_Empty(t *testing.T) {
	client, _ := newAuthenticatedClient(t)

	// Check if a test reservation token is provided
	testToken := os.Getenv("RESY_TEST_RESERVATION_TOKEN")
	if testToken == "" {
		// Try to get an upcoming reservation
		resType := UpcomingReservation
		limit := 1
		reservations, err := client.GetReservations(&resType, nil, &limit, nil)
		requireNoError(t, err, "GetReservations failed")

		if len(reservations) == 0 {
			t.Skip("No upcoming reservations available for SetReservationSpecialRequest test")
		}

		testToken = reservations[0].ReservationToken
	}

	// Set empty special request (clearing it)
	err := client.SetReservationSpecialRequest(testToken, "")
	requireNoError(t, err, "SetReservationSpecialRequest with empty string failed")

	t.Log("Successfully cleared special request")
}

func TestReservations_SetSpecialRequest_InvalidToken(t *testing.T) {
	client, _ := newAuthenticatedClient(t)

	// Use an invalid reservation token
	err := client.SetReservationSpecialRequest("invalid-token-12345", "Test request")

	// Should get an error (likely 404 or 419)
	if err == nil {
		t.Error("expected error for invalid reservation token")
	}

	if !errors.Is(err, ErrNotFound) && !errors.Is(err, ErrUnauthorized) && !errors.Is(err, ErrBadRequest) {
		t.Logf("Got error: %v (acceptable)", err)
	}
}

func TestReservations_OccasionConstants(t *testing.T) {
	// Validate that predefined occasions have proper values
	occasions := []ReservationOccasion{
		AnniversaryOccasion,
		BirthdayOccasion,
		BusinessOccasion,
		GraduationOccasion,
		NoOccasion,
	}

	for i, occasion := range occasions[:len(occasions)-1] { // Exclude NoOccasion
		if occasion.OccasionId == "" {
			t.Errorf("Occasion %d: OccasionId should not be empty", i)
		}
		if occasion.Occasion == "" {
			t.Errorf("Occasion %d: Occasion should not be empty", i)
		}
	}

	// NoOccasion should be empty
	if NoOccasion.OccasionId != "" || NoOccasion.Occasion != "" {
		t.Error("NoOccasion should have empty values")
	}
}
