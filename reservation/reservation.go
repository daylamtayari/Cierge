package reservation

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type ctxKey string

const (
	startTimeKey = ctxKey("start_time")
)

var (
	ErrBase64Decode = errors.New("failed to decode base64")
	ErrDecrypt      = errors.New("failed to decrypt token")
)

// Main handler of the reservation job and handles the core logic
// - Decrypts token
// - Creates booking client
// - Performs pre-booking checks
// - Performs booking
// Returns an Output type representing the output of the reservation job
func Handle(ctx context.Context, event Event, decrypter Decrypter) Output {
	startTime := time.Now().UTC()

	output := Output{
		JobId:     event.JobID,
		StartTime: startTime,
	}

	ctx = context.WithValue(ctx, startTimeKey, startTime)

	token, err := decryptToken(ctx, event.EncryptedToken, decrypter)
	if err != nil {
		output.Message = "failed to decrypt token"
		output.Success = false
		output.Error = err.Error()
		output.Level = "error"
		return complete(ctx, event, output, decrypter)
	}

	bookingClient, err := NewBookingClient(event.Platform, token)
	if err != nil {
		output.Message = "failed to create booking client"
		output.Success = false
		output.Error = err.Error()
		output.Level = "error"
		return complete(ctx, event, output, decrypter)
	}

	err = bookingClient.PreBookingCheck(ctx, event)
	if err != nil {
		output.Message = "failed to perform pre-booking checks"
		output.Success = false
		output.Error = err.Error()
		output.Level = "error"
		return complete(ctx, event, output, decrypter)
	}

	waitUntil(ctx, event.DropTime)
	output.BookingStart = time.Now().UTC()
	output.DriftNs = time.Since(event.DropTime).Nanoseconds()

	bookingResult, err := bookingClient.Book(ctx, event)
	if err != nil {
		output.Message = "failed to perform booking"
		output.Success = false
		output.Error = err.Error()
		output.Level = "error"
		return complete(ctx, event, output, decrypter)
	}

	output.ReservationTime = bookingResult.ReservationTime
	output.PlatformConfirmation = bookingResult.PlatformConfirmation
	output.Success = true
	output.Message = "reservation completed successfully"
	output.Level = "info"

	return complete(ctx, event, output, decrypter)
}

// Decrypts the users token using the Decrypter interface
func decryptToken(ctx context.Context, encryptedToken string, decrypter Decrypter) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedToken)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrBase64Decode, err)
	}

	decrypted, err := decrypter.Decrypt(ctx, ciphertext)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrDecrypt, err)
	}

	return string(decrypted), nil
}

// Exit handler of the Lambda
// Calculates duration, notifies server of output, and outputs job output to stdout
func complete(ctx context.Context, event Event, output Output, decrypter Decrypter) Output {
	if startTime, ok := ctx.Value(startTimeKey).(time.Time); ok {
		output.Duration = time.Now().UTC().Sub(startTime)
	} else {
		output.Duration = time.Duration(0)
	}

	if !event.Callback {
		return output
	}

	callbackSecret, err := decryptToken(ctx, event.EncryptedCallbackSecret, decrypter)
	if err != nil {
		// Keep success as true if the reservation completed
		// as that is the core goal of this lambda and the
		// output will still be sent to stdout
		output.Message += " - error: failed to decrypt token"
		output.Error = err.Error()
		output.Level = "error"
	} else {
		marshalledOutput, _ := json.Marshal(output)
		err = notifyServer(ctx, event.ServerEndpoint, callbackSecret, marshalledOutput)
		if err != nil {
			// Keep success as true if the reservation completed
			// as that is the core goal of this lambda and the
			// output will still be sent to stdout
			output.Message += " - error: failed to notify server"
			output.Error = err.Error()
			output.Level = "error"
		}
	}

	return output
}

// Waits until a specified time
func waitUntil(ctx context.Context, target time.Time) {
	for {
		remaining := time.Until(target)

		// If at or past target time, return
		if remaining <= 0 {
			return
		}

		// If more than 1 second away, sleep for 0.5s
		if remaining > 1*time.Second {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		// If more than 0.1s left, sleep for 0.1s
		if remaining > 100*time.Millisecond {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		for time.Now().Before(target) {
			// Busy wait until the exact time is hit
			if ctx.Err() != nil {
				return
			}
		}
		return
	}
}
