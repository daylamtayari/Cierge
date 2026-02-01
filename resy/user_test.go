package resy

import (
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestUser_Get(t *testing.T) {
	client, _ := newAuthenticatedClient(t)

	user, err := client.GetUser()
	requireNoError(t, err, "GetUser failed")
	assertNotNil(t, user, "user should not be nil")

	// Validate basic user fields
	if user.Id <= 0 {
		t.Errorf("User Id should be positive, got: %d", user.Id)
	}

	if user.FirstName == "" {
		t.Error("FirstName should not be empty")
	}

	if user.LastName == "" {
		t.Error("LastName should not be empty")
	}

	if user.EmailAddress == "" {
		t.Error("EmailAddress should not be empty")
	}

	if !strings.Contains(user.EmailAddress, "@") {
		t.Errorf("EmailAddress should contain @, got: %s", user.EmailAddress)
	}

	// PaymentMethods slice should exist (may be empty)
	if user.PaymentMethods == nil {
		t.Error("PaymentMethods should not be nil (can be empty slice)")
	}

	t.Logf("User: %s %s (ID: %d, Email: %s, Payment Methods: %d)",
		user.FirstName, user.LastName, user.Id, user.EmailAddress, len(user.PaymentMethods))
}

func TestUser_Unauthorized(t *testing.T) {
	// Create client without auth token (API key only)
	tokens := Tokens{
		ApiKey: DefaultApiKey,
		Token:  "", // No user token
	}

	client := NewClient(http.DefaultClient, tokens, "")

	user, err := client.GetUser()

	// Should get unauthorized error (419 status code)
	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("expected ErrUnauthorized, got: %v", err)
	}

	if user != nil {
		t.Errorf("expected nil user on error, got: %v", user)
	}
}

func TestUser_GetDefaultPaymentMethod(t *testing.T) {
	client, _ := newAuthenticatedClient(t)

	user, err := client.GetUser()
	requireNoError(t, err, "GetUser failed")

	// Get default payment method
	paymentMethod := GetDefaultPaymentMethod(user)

	// If payment methods exist, validate the default one
	if len(user.PaymentMethods) > 0 {
		// Should have found a default payment method
		if paymentMethod.Id == 0 {
			t.Error("Expected default payment method, but got empty PaymentMethod")
		}

		// Verify it's actually marked as default
		found := false
		for _, pm := range user.PaymentMethods {
			if pm.Id == paymentMethod.Id {
				if !pm.IsDefault {
					t.Errorf("Payment method ID %d should be marked as default", pm.Id)
				}
				found = true
				break
			}
		}

		if !found {
			t.Errorf("Default payment method ID %d not found in user's payment methods", paymentMethod.Id)
		}

		// Validate payment method fields
		if paymentMethod.Display == "" {
			t.Error("Payment method Display should not be empty")
		}

		// Expiration year should be reasonable (current year - 5 to current year + 10)
		currentYear := time.Now().Year()
		if paymentMethod.ExpirationYear != 0 {
			if paymentMethod.ExpirationYear < currentYear-5 || paymentMethod.ExpirationYear > currentYear+10 {
				t.Errorf("ExpirationYear seems unreasonable: %d", paymentMethod.ExpirationYear)
			}
		}

		// Expiration month should be 1-12
		if paymentMethod.ExpirationMonth != 0 {
			if paymentMethod.ExpirationMonth < 1 || paymentMethod.ExpirationMonth > 12 {
				t.Errorf("ExpirationMonth out of range: %d", paymentMethod.ExpirationMonth)
			}
		}

		t.Logf("Default Payment Method: %s (ID: %d, Type: %s, Expires: %d/%d)",
			paymentMethod.Display, paymentMethod.Id, paymentMethod.Type,
			paymentMethod.ExpirationMonth, paymentMethod.ExpirationYear)
	} else {
		// No payment methods, should return empty PaymentMethod
		if paymentMethod.Id != 0 {
			t.Errorf("Expected empty PaymentMethod when no payment methods exist, got ID: %d", paymentMethod.Id)
		}
		t.Log("User has no payment methods")
	}
}

func TestUser_GetDefaultPaymentMethod_NoPaymentMethods(t *testing.T) {
	// Create a user with empty payment methods
	user := &User{
		Id:             123,
		FirstName:      "Test",
		LastName:       "User",
		PaymentMethods: []PaymentMethod{},
	}

	paymentMethod := GetDefaultPaymentMethod(user)

	// Should return empty PaymentMethod
	if paymentMethod.Id != 0 {
		t.Errorf("Expected empty PaymentMethod, got ID: %d", paymentMethod.Id)
	}

	if paymentMethod.Display != "" {
		t.Errorf("Expected empty Display, got: %s", paymentMethod.Display)
	}
}

func TestUser_GetDefaultPaymentMethod_MultiplePaymentMethods(t *testing.T) {
	// Create a user with multiple payment methods, one marked as default
	user := &User{
		Id:        123,
		FirstName: "Test",
		LastName:  "User",
		PaymentMethods: []PaymentMethod{
			{
				Id:        1,
				IsDefault: false,
				Display:   "1234",
				Type:      "visa",
			},
			{
				Id:        2,
				IsDefault: true,
				Display:   "5678",
				Type:      "mastercard",
			},
			{
				Id:        3,
				IsDefault: false,
				Display:   "9012",
				Type:      "amex",
			},
		},
	}

	paymentMethod := GetDefaultPaymentMethod(user)

	// Should return the default payment method (ID 2)
	if paymentMethod.Id != 2 {
		t.Errorf("Expected payment method ID 2, got: %d", paymentMethod.Id)
	}

	if !paymentMethod.IsDefault {
		t.Error("Expected IsDefault to be true")
	}

	if paymentMethod.Display != "5678" {
		t.Errorf("Expected Display '5678', got: %s", paymentMethod.Display)
	}
}
