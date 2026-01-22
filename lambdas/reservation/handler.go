package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/kms"
)

type ctxKey string

const (
	startTimeKey = ctxKey("start_time")
)

var (
	ErrBase64Decode = errors.New("failed to decode base64")
	ErrKmsDecrypt   = errors.New("failed to decrypt token")
)

func HandleRequest(ctx context.Context, event LambdaEvent) error {
	logs := make([]LogEntry, 0)
	ctx = context.WithValue(ctx, startTimeKey, time.Now().UTC())

	logEntry(&logs, slog.LevelInfo, "starting job", map[string]any{
		"job_id": event.JobID,
	})

	token, err := decryptToken(ctx, event.EncryptedToken)
	if err != nil {
		logEntry(&logs, slog.LevelError, "failed to decrypt token", map[string]any{
			"error": err.Error(),
		})
		complete(ctx, event, JobOutput{
			Status:       StatusFail,
			ErrorMessage: err.Error(),
			Logs:         logs,
		})
	}

	bookingClient, err := NewBookingClient(event.Platform, token)
	if err != nil {
		logEntry(&logs, slog.LevelError, "failed to create booking client", map[string]any{
			"error": err.Error(),
		})
		complete(ctx, event, JobOutput{
			Status:       StatusFail,
			ErrorMessage: err.Error(),
			Logs:         logs,
		})
	}

	err = bookingClient.PreBookingCheck(ctx, event)
	if err != nil {
		logEntry(&logs, slog.LevelError, "failed to perform pre-booking checks", map[string]any{
			"error": err.Error(),
		})
		complete(ctx, event, JobOutput{
			Status:       StatusFail,
			ErrorMessage: err.Error(),
			Logs:         logs,
		})
	}

	waitUntil(event.DropTime)
	logEntry(&logs, slog.LevelInfo, "starting booking attempts", map[string]any{
		"actual_time": time.Now().UTC().Format(time.RFC3339Nano),
		"drift_ns":    time.Since(event.DropTime).Nanoseconds(),
	})

	bookingResult, err := bookingClient.Book(ctx, event)
	if err != nil {
		logEntry(&logs, slog.LevelError, "failed to perform booking", map[string]any{
			"error": err.Error(),
		})
		complete(ctx, event, JobOutput{
			Status:       StatusFail,
			ErrorMessage: err.Error(),
			Logs:         logs,
		})
	}

	logEntry(&logs, slog.LevelInfo, "booking attempt successful", map[string]any{
		"reservation_time":      bookingResult.ReservationTime,
		"platform_confirmation": bookingResult.PlatformConfirmation,
	})

	complete(ctx, event, JobOutput{
		Status: StatusSuccess,
		Logs:   logs,
	})
	return nil
}

func decryptToken(ctx context.Context, encryptedToken string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedToken)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrBase64Decode, err)
	}

	decrypted, err := kmsClient.Decrypt(ctx, &kms.DecryptInput{
		CiphertextBlob: ciphertext,
	})
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrKmsDecrypt, err)
	}

	return string(decrypted.Plaintext), nil
}

func logEntry(logs *[]LogEntry, level slog.Level, msg string, fields map[string]any) {
	*logs = append(*logs, LogEntry{
		Time:   time.Now().UTC(),
		Level:  level,
		Msg:    msg,
		Fields: fields,
	})
}

func complete(ctx context.Context, event LambdaEvent, output JobOutput) {
	if startTime, ok := ctx.Value(startTimeKey).(time.Time); ok {
		output.Duration = time.Now().UTC().Sub(startTime)
	} else {
		output.Duration = time.Duration(0)
	}

	output.JobId = event.JobID
	// TODO: Notify server
	marshalledOutput, _ := json.Marshal(output)
	fmt.Print(string(marshalledOutput))
}

// Waits until a specified time
func waitUntil(target time.Time) {
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
		}
		return
	}
}
