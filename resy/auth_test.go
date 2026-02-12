package resy

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestLogin_Success(t *testing.T) {
	// Skip if credentials aren't provided
	email := os.Getenv("RESY_TEST_EMAIL")
	password := os.Getenv("RESY_TEST_PASSWORD")

	if email == "" || password == "" {
		t.Skip("Skipping test: RESY_TEST_EMAIL and RESY_TEST_PASSWORD not set")
	}

	// Create unauthenticated client (API key only)
	client, _ := newUnauthenticatedClient(t)

	// Perform login
	tokens, err := client.Login(email, password)
	requireNoError(t, err, "Login failed")

	// Validate auth token
	if tokens.Token == "" {
		t.Error("Auth token should not be empty")
	}

	// Validate refresh token
	if tokens.Refresh == "" {
		t.Error("Refresh token should not be empty")
	}

	// Validate token is a valid JWT
	expiry, err := getTokenExpiry(tokens.Token)
	requireNoError(t, err, "Failed to parse auth token expiry")

	// Token should expire in the future
	if expiry.Before(time.Now()) {
		t.Errorf("Auth token should expire in the future, got: %v", expiry)
	}

	// Token should expire within ~45 days (allow some variance)
	expectedExpiry := time.Now().Add(45 * 24 * time.Hour)
	if expiry.After(expectedExpiry.Add(24*time.Hour)) || expiry.Before(expectedExpiry.Add(-24*time.Hour)) {
		t.Logf("Warning: Auth token expiry is outside expected range (45 days), got: %v", expiry.Sub(time.Now()))
	}

	t.Logf("Login successful - Token expires: %v (in %v)", expiry, time.Until(expiry).Round(time.Hour))
}

func TestLogin_InvalidCredentials(t *testing.T) {
	client, _ := newUnauthenticatedClient(t)

	// Use clearly invalid credentials
	tokens, err := client.Login("invalid@example.com", "wrongpassword")

	// Should get an error
	if err == nil {
		t.Error("Expected error for invalid credentials, got nil")
	}

	// Token should be empty on error
	if tokens.Token != "" {
		t.Errorf("Expected empty token on error, got: %s", tokens.Token)
	}

	if tokens.Refresh != "" {
		t.Errorf("Expected empty refresh token on error, got: %s", tokens.Refresh)
	}

	t.Logf("Invalid credentials correctly rejected with error: %v", err)
}

func TestLogin_EmptyCredentials(t *testing.T) {
	client, _ := newUnauthenticatedClient(t)

	// Test with empty credentials
	tokens, err := client.Login("", "")

	// Should get an error
	if err == nil {
		t.Error("Expected error for empty credentials, got nil")
	}

	// Tokens should be empty on error
	if tokens.Token != "" || tokens.Refresh != "" {
		t.Error("Expected empty tokens on error")
	}
}

func TestRefreshToken_Success(t *testing.T) {
	// This test requires a valid refresh token
	// First, let's try to get one via login
	email := os.Getenv("RESY_TEST_EMAIL")
	password := os.Getenv("RESY_TEST_PASSWORD")

	if email == "" || password == "" {
		t.Skip("Skipping test: RESY_TEST_EMAIL and RESY_TEST_PASSWORD not set")
	}

	client, _ := newUnauthenticatedClient(t)

	// Get initial tokens
	initialTokens, err := client.Login(email, password)
	requireNoError(t, err, "Login failed")

	time.Sleep(1 * time.Second)

	// Now use refresh token to get new tokens
	newTokens, err := client.RefreshToken(initialTokens.Refresh)
	requireNoError(t, err, "RefreshToken failed")

	// Validate new auth token
	if newTokens.Token == "" {
		t.Error("New auth token should not be empty")
	}

	// Validate new refresh token
	if newTokens.Refresh == "" {
		t.Error("New refresh token should not be empty")
	}

	// New tokens should be different from initial tokens
	if newTokens.Token == initialTokens.Token {
		t.Error("New auth token should be different from initial token")
	}

	if newTokens.Refresh == initialTokens.Refresh {
		t.Error("New refresh token should be different from initial refresh token")
	}

	t.Log("Token refresh successful")
}

func TestRefreshToken_InvalidToken(t *testing.T) {
	client, _ := newUnauthenticatedClient(t)

	// Use an invalid refresh token
	tokens, err := client.RefreshToken("invalid-refresh-token")

	// Should get an error
	if err == nil {
		t.Error("Expected error for invalid refresh token, got nil")
	}

	// Tokens should be empty on error
	if tokens.Token != "" || tokens.Refresh != "" {
		t.Error("Expected empty tokens on error")
	}

	t.Logf("Invalid refresh token correctly rejected with error: %v", err)
}

func TestGetTokenExpiry(t *testing.T) {
	testCases := []struct {
		name        string
		setupToken  func() string
		expectError bool
		expectedErr error
	}{
		{
			name: "Valid JWT with expiration",
			setupToken: func() string {
				expTime := time.Now().Add(24 * time.Hour)
				claims := jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(expTime),
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString([]byte("test-secret"))
				return tokenString
			},
			expectError: false,
		},
		{
			name: "JWT without expiration",
			setupToken: func() string {
				claims := jwt.RegisteredClaims{
					Subject: "test",
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString([]byte("test-secret"))
				return tokenString
			},
			expectError: true,
			expectedErr: ErrNoExpirationTime,
		},
		{
			name: "Invalid JWT",
			setupToken: func() string {
				return "not.a.valid.jwt"
			},
			expectError: true,
		},
		{
			name: "Empty string",
			setupToken: func() string {
				return ""
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			token := tc.setupToken()
			expiry, err := getTokenExpiry(token)

			if tc.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				}
				if tc.expectedErr != nil && !errors.Is(err, tc.expectedErr) {
					t.Errorf("Expected error %v, got: %v", tc.expectedErr, err)
				}
			} else {
				requireNoError(t, err, "getTokenExpiry failed")
				if expiry.IsZero() {
					t.Error("Expected non-zero expiry time")
				}
			}
		})
	}
}
