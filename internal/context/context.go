package context

import (
	"context"
	"fmt"
	"os"

	"github.com/daylamtayari/cierge/pkg/errcol"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type ctxKey string

const (
	loggerKey         ctxKey = "logger"
	requestIDKey      ctxKey = "request_id"
	userIDKey         ctxKey = "user_id"
	errorCollectorKey ctxKey = "error_collector"
)

// Adds a logger to the context
func WithLogger(ctx context.Context, logger *zerolog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// Returns the logger from the context
func Logger(ctx context.Context) *zerolog.Logger {
	if logger, ok := ctx.Value(loggerKey).(*zerolog.Logger); ok {
		return logger
	}
	fallback := zerolog.New(os.Stdout).With().Timestamp().Logger()
	fallback.Error().Msg("logger not in context")
	return &fallback
}

// Adds a request ID to the context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// Returns the request ID from the context
func RequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(requestIDKey).(string); ok {
		return requestID
	}
	// If no error, generate a fallback UUID and log
	fallbackID := uuid.NewString()
	logger := Logger(ctx)
	logger.Error().Msgf("request ID not in context, generating fallback ID: %s", fallbackID)
	return fallbackID
}

// Adds a user ID to the context
func WithUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// Returns the user ID from the context
func UserID(ctx context.Context) uuid.UUID {
	if userID, ok := ctx.Value(userIDKey).(uuid.UUID); ok {
		return userID
	}
	// If the value was not set, return a zero value
	// This *should* never happen, but this zero value
	// UUID makes sure that any future user operations
	// should fail as there can never be a corresponding
	// user ID. A UUIDv4 can never be all zeros due to
	// the version and variant bits.
	return uuid.UUID{}
}

// Adds an error collector to the context
func WithErrorCollector(ctx context.Context, errorCollector *errcol.ErrorCollector) context.Context {
	return context.WithValue(ctx, errorCollectorKey, errorCollector)
}

// Returns the error collector from the context
func ErrorCollector(ctx context.Context) *errcol.ErrorCollector {
	if errorCollector, ok := ctx.Value(errorCollectorKey).(*errcol.ErrorCollector); ok {
		return errorCollector
	}
	errorCollector := &errcol.ErrorCollector{}
	errorCollector.Add(fmt.Errorf("error collector not in context, creating new"), zerolog.ErrorLevel, false, nil, "error collector not found in context")
	return errorCollector
}
