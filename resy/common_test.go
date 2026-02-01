package resy

import (
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"
)

// TestConfig holds configuration for integration tests
type TestConfig struct {
	ApiKey    string
	AuthToken string
	VenueId   int
	TestDate  time.Time
	HasAuth   bool
}

// Test venue IDs for different cities
const (
	TestVenue        = 54602 // A known valid venue
	TestVenueInvalid = 99999
)

// loadTestConfig loads test configuration from environment variables
func loadTestConfig(t *testing.T) *TestConfig {
	t.Helper()

	apiKey := os.Getenv("RESY_API_KEY")
	if apiKey == "" {
		apiKey = DefaultApiKey
	}

	authToken := os.Getenv("RESY_AUTH_TOKEN")

	config := &TestConfig{
		ApiKey:    apiKey,
		AuthToken: authToken,
		HasAuth:   authToken != "",
		VenueId:   TestVenue, // Default venue
		TestDate:  getDefaultTestDate(),
	}

	// Parse optional test venue ID
	if venueIdStr := os.Getenv("RESY_TEST_VENUE_ID"); venueIdStr != "" {
		if id, err := strconv.Atoi(venueIdStr); err == nil {
			config.VenueId = id
		}
	}

	// Parse optional test date
	if dateStr := os.Getenv("RESY_TEST_DATE"); dateStr != "" {
		if date, err := time.Parse("2006-01-02", dateStr); err == nil {
			config.TestDate = date
		}
	}

	return config
}

// newUnauthenticatedClient creates a client with API key only (no user token)
// Skips the test if API key cannot be loaded
func newUnauthenticatedClient(t *testing.T) (*Client, *TestConfig) {
	t.Helper()

	config := loadTestConfig(t)

	tokens := Tokens{
		ApiKey: config.ApiKey,
		Token:  "", // No user token
	}

	client := NewClient(http.DefaultClient, tokens, "")
	return client, config
}

// newAuthenticatedClient creates a client with API key + user token
// Skips the test if RESY_AUTH_TOKEN is not set
func newAuthenticatedClient(t *testing.T) (*Client, *TestConfig) {
	t.Helper()

	config := loadTestConfig(t)
	if !config.HasAuth {
		t.Skip("Skipping test: RESY_AUTH_TOKEN not set")
	}

	tokens := Tokens{
		ApiKey: config.ApiKey,
		Token:  config.AuthToken,
	}

	client := NewClient(http.DefaultClient, tokens, "")
	return client, config
}

// skipIfNoAuth skips the test if authentication is not available
func skipIfNoAuth(t *testing.T, config *TestConfig) {
	t.Helper()
	if !config.HasAuth {
		t.Skip("Skipping test: RESY_AUTH_TOKEN not set")
	}
}

// requireNoError fails the test if err is not nil
func requireNoError(t *testing.T, err error, msgAndArgs ...interface{}) {
	t.Helper()
	if err != nil {
		if len(msgAndArgs) > 0 {
			format, ok := msgAndArgs[0].(string)
			if ok {
				t.Fatalf(format+": %v", append(msgAndArgs[1:], err)...)
			} else {
				t.Fatalf("unexpected error: %v (context: %v)", err, msgAndArgs)
			}
		} else {
			t.Fatalf("unexpected error: %v", err)
		}
	}
}

// assertNotNil fails the test if value is nil
func assertNotNil(t *testing.T, value interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	if value == nil {
		if len(msgAndArgs) > 0 {
			format, ok := msgAndArgs[0].(string)
			if ok {
				t.Fatalf(format, msgAndArgs[1:]...)
			} else {
				t.Fatalf("expected non-nil value, got nil (context: %v)", msgAndArgs)
			}
		} else {
			t.Fatal("expected non-nil value, got nil")
		}
	}
}

// getDefaultTestDate returns a date 14 days in the future
func getDefaultTestDate() time.Time {
	return time.Now().AddDate(0, 0, 14)
}
