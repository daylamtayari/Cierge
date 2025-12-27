package context

import (
	"context"
	"os"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type ctxKey string

const (
	loggerKey    ctxKey = "logger"
	requestIDKey ctxKey = "request_id"
	claimsKey    ctxKey = "claims"
)

// Adds a logger to the context
func WithLogger(ctx context.Context, logger zerolog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// Returns the logger from the context
func Logger(ctx context.Context) zerolog.Logger {
	if logger, ok := ctx.Value(loggerKey).(zerolog.Logger); ok {
		return logger
	}
	fallback := zerolog.New(os.Stdout).With().Timestamp().Logger()
	fallback.Error().Msg("logger not in context")
	return fallback
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

// TODO: Implement claims
