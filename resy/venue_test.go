package resy

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestVenue_Search(t *testing.T) {
	client, _ := newUnauthenticatedClient(t)

	// Search for a well-known restaurant
	venues, err := client.SearchVenue("Carbone", nil)
	requireNoError(t, err, "SearchVenue failed")

	if venues == nil {
		t.Fatal("venues should not be nil (can be empty slice)")
	}

	if len(venues) == 0 {
		t.Skip("No venues found for 'Carbone' - this is acceptable but unexpected")
	}

	// Validate first venue
	venue := venues[0]
	if venue.Id.Resy == 0 {
		t.Error("Venue Id.Resy should not be 0")
	}

	if venue.Name == "" {
		t.Error("Venue Name should not be empty")
	}

	// Name should contain the search query (case-insensitive)
	if !strings.Contains(strings.ToLower(venue.Name), "carbone") {
		t.Logf("Warning: Venue name %q doesn't contain 'carbone'", venue.Name)
	}

	if venue.UrlSlug == "" {
		t.Error("UrlSlug should not be empty")
	}

	t.Logf("Found %d venues, first: %s (ID: %d)", len(venues), venue.Name, venue.Id.Resy)
}

func TestVenue_Search_WithPageLimit(t *testing.T) {
	client, _ := newUnauthenticatedClient(t)

	// Test with different page limits
	limits := []int{5, 10, 20}

	for _, limit := range limits {
		t.Run("limit="+string(rune(limit+'0')), func(t *testing.T) {
			venues, err := client.SearchVenue("restaurant", &limit)
			requireNoError(t, err, "SearchVenue failed")

			if venues == nil {
				t.Fatal("venues should not be nil")
			}

			if len(venues) > limit {
				t.Errorf("Expected at most %d venues, got %d", limit, len(venues))
			}

			t.Logf("Limit %d: returned %d venues", limit, len(venues))
		})
	}
}

func TestVenue_Search_NoResults(t *testing.T) {
	client, _ := newUnauthenticatedClient(t)

	// Search for nonsense query that should return no results
	venues, err := client.SearchVenue("xyzabc123notarestaurant999", nil)
	requireNoError(t, err, "SearchVenue should not error on no results")

	if venues == nil {
		t.Fatal("venues should not be nil (should be empty slice)")
	}

	if len(venues) != 0 {
		t.Logf("Warning: Expected 0 venues for nonsense query, got %d", len(venues))
	}
}

func TestVenue_Get(t *testing.T) {
	client, config := newUnauthenticatedClient(t)

	venue, err := client.GetVenue(config.VenueId)
	requireNoError(t, err, "GetVenue failed")
	assertNotNil(t, venue, "venue should not be nil")

	// Validate venue structure
	if venue.Id.Resy != config.VenueId {
		t.Errorf("Venue Id.Resy: got %d, want %d", venue.Id.Resy, config.VenueId)
	}

	if venue.Name == "" {
		t.Error("Venue Name should not be empty")
	}

	// Validate location coordinates
	if venue.Location.Geo.Lat < -90 || venue.Location.Geo.Lat > 90 {
		t.Errorf("Latitude out of range: %v", venue.Location.Geo.Lat)
	}

	if venue.Location.Geo.Lon < -180 || venue.Location.Geo.Lon > 180 {
		t.Errorf("Longitude out of range: %v", venue.Location.Geo.Lon)
	}

	// Rating should be valid
	if venue.Rating.Score < 0 || venue.Rating.Score > 5 {
		t.Errorf("Rating.Score out of range: %v (expected 0-5)", venue.Rating.Score)
	}

	// Contact info should be present
	if venue.Contact.PhoneNumber == "" && venue.Contact.Website == "" {
		t.Log("Warning: Both PhoneNumber and Website are empty")
	}

	t.Logf("Venue: %s (ID: %d, Rating: %.1f, Location: %s)",
		venue.Name, venue.Id.Resy, venue.Rating.Score, venue.Locality)
}

func TestVenue_Get_InvalidId(t *testing.T) {
	client, _ := newUnauthenticatedClient(t)

	venue, err := client.GetVenue(TestVenueInvalid)

	// Should get not found error (404)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}

	if venue != nil {
		t.Errorf("expected nil venue on error, got: %v", venue)
	}
}

func TestVenue_GetConfig(t *testing.T) {
	client, config := newUnauthenticatedClient(t)

	venue, err := client.GetVenueConfig(config.VenueId, nil)
	requireNoError(t, err, "GetVenueConfig failed")
	assertNotNil(t, venue, "venue should not be nil")

	// Validate config fields
	if venue.LeadTimeInDays <= 0 {
		t.Errorf("LeadTimeInDays should be positive, got: %d", venue.LeadTimeInDays)
	}

	if venue.MinPartySize <= 0 || venue.MinPartySize > 20 {
		t.Errorf("MinPartySize seems unreasonable: %d", venue.MinPartySize)
	}

	if venue.MaxPartySize <= 0 || venue.MaxPartySize > 100 {
		t.Errorf("MaxPartySize seems unreasonable: %d", venue.MaxPartySize)
	}

	if venue.MinPartySize > venue.MaxPartySize {
		t.Errorf("MinPartySize (%d) > MaxPartySize (%d)", venue.MinPartySize, venue.MaxPartySize)
	}

	if len(venue.ServiceTypes) == 0 {
		t.Error("ServiceTypes should not be empty")
	}

	if venue.Name == "" {
		t.Error("Name should not be empty")
	}

	t.Logf("Venue Config: %s (Lead Time: %d days, Party Size: %d-%d, Service Types: %d)",
		venue.Name, venue.LeadTimeInDays, venue.MinPartySize, venue.MaxPartySize, len(venue.ServiceTypes))
}

func TestVenue_GetConfig_WithProvidedVenue(t *testing.T) {
	client, config := newUnauthenticatedClient(t)

	// First get the venue
	originalVenue, err := client.GetVenue(config.VenueId)
	requireNoError(t, err, "GetVenue failed")

	originalName := originalVenue.Name
	originalRating := originalVenue.Rating

	// Now augment it with config
	augmentedVenue, err := client.GetVenueConfig(config.VenueId, originalVenue)
	requireNoError(t, err, "GetVenueConfig failed")

	// Should be the same pointer
	if augmentedVenue != originalVenue {
		t.Error("Expected same venue pointer to be returned")
	}

	// Config fields should be populated
	if augmentedVenue.LeadTimeInDays <= 0 {
		t.Error("LeadTimeInDays should be populated")
	}

	// Original fields should be preserved (though Name gets overwritten)
	if augmentedVenue.Rating.Score != originalRating.Score {
		t.Logf("Note: Rating changed from %.1f to %.1f", originalRating.Score, augmentedVenue.Rating.Score)
	}

	t.Logf("Augmented venue: %s (was: %s)", augmentedVenue.Name, originalName)
}

func TestVenue_GetCalendar(t *testing.T) {
	client, config := newUnauthenticatedClient(t)

	// Get calendar for a week range (7 to 14 days from now)
	startDate := ResyDate{Time: time.Now().AddDate(0, 0, 7)}
	endDate := ResyDate{Time: time.Now().AddDate(0, 0, 14)}
	numSeats := 2

	slots, err := client.GetVenueCalendar(config.VenueId, numSeats, startDate, endDate)
	requireNoError(t, err, "GetVenueCalendar failed")

	if slots == nil {
		t.Fatal("slots should not be nil (can be empty slice)")
	}

	// Validate slots if any are returned
	if len(slots) > 0 {
		for i, slot := range slots {
			// Date should be valid
			if slot.Date.Time.IsZero() {
				t.Errorf("Slot %d: Date is zero", i)
			}

			// Inventory status should be one of the expected values
			validStatuses := []string{"available", "not available", "sold-out"}
			if !contains(validStatuses, slot.Inventory.Reservation) {
				t.Errorf("Slot %d: Unexpected Reservation status: %s", i, slot.Inventory.Reservation)
			}

			if i == 0 {
				t.Logf("First slot: %s - Reservation: %s, Event: %s, WalkIn: %s",
					slot.Date.Format(ResyDateFormat),
					slot.Inventory.Reservation,
					slot.Inventory.Event,
					slot.Inventory.WalkIn)
			}
		}
	}

	t.Logf("Found %d calendar slots for %d seats", len(slots), numSeats)
}

func TestVenue_GetCalendar_InvalidVenue(t *testing.T) {
	client, _ := newUnauthenticatedClient(t)

	startDate := ResyDate{Time: time.Now().AddDate(0, 0, 7)}
	endDate := ResyDate{Time: time.Now().AddDate(0, 0, 14)}

	slots, err := client.GetVenueCalendar(TestVenueInvalid, 2, startDate, endDate)

	// Should get an error (likely 502 or 404)
	if err == nil {
		t.Error("expected error for invalid venue ID")
	}

	// Some error types are acceptable
	if !errors.Is(err, ErrBadGateway) && !errors.Is(err, ErrNotFound) {
		t.Logf("Got error: %v (acceptable)", err)
	}

	if slots != nil {
		t.Logf("Warning: Got slots despite error: %v", slots)
	}
}

// Helper function to check if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
