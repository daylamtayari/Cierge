package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	appctx "github.com/daylamtayari/cierge/internal/context"
	"github.com/daylamtayari/cierge/internal/version"
	"github.com/daylamtayari/cierge/pkg/errcol"
)

func Logger(baseLogger zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		requestID := appctx.RequestID(c.Request.Context())
		var contentType string
		if contentType = c.ContentType(); contentType == "" {
			contentType = "unspecified"
		}

		// Create request-scoped logger
		logger := baseLogger.With().
			Str("request_id", requestID).
			Str("version", version.Version).
			Str("gin_version", gin.Version).
			// Request information
			// Intentional no logging of IP for privacy
			Dict("request", zerolog.Dict().
				Str("method", c.Request.Method).
				Str("proto", c.Request.Proto).
				Str("path", c.Request.URL.Path).
				Str("user_agent", c.Request.UserAgent()).
				Int64("content_length", c.Request.ContentLength).
				Str("content_type", contentType).
				Interface("headers", sanitizeHeaders(c.Request.Header.Clone())),
			).Logger()

		// Create our error collector as well
		errorCol := errcol.NewErrorCollector()

		ctx := appctx.WithLogger(c.Request.Context(), &logger)
		ctx = appctx.WithErrorCollector(ctx, errorCol)
		c.Request = c.Request.WithContext(ctx)

		c.Next()

		// Log request completion
		duration := time.Since(start)
		statusCode := c.Writer.Status()

		logger = *appctx.Logger(c.Request.Context())
		errorCol = appctx.ErrorCollector(c.Request.Context())

		var logLevel zerolog.Level
		var highestSeverity errcol.ErrorInfo
		logMessage := "request completed"
		if errorCol.HasErrors() {
			highestSeverity = errorCol.HighestSeverity()
			logMessage = highestSeverity.Message
		}

		if highestSeverity.Severity >= zerolog.ErrorLevel || statusCode >= 500 {
			logLevel = zerolog.ErrorLevel
		} else if highestSeverity.Severity == zerolog.WarnLevel {
			logLevel = zerolog.WarnLevel
		} else if highestSeverity.Severity == zerolog.InfoLevel || statusCode >= 400 {
			logLevel = zerolog.InfoLevel
		} else {
			// Duplicate as above but kept for readability
			logLevel = zerolog.InfoLevel
		}

		// This project should be used at a small enough scale that these
		// log events become too significant to justify log sampling.
		logEvent := logger.WithLevel(logLevel)
		if errorCol.HasErrors() {
			logEvent = errorCol.ApplyToEvent(logEvent)
		}
		logEvent.
			Dict("response", zerolog.Dict().
				Int("status", statusCode).
				Int("body_size", c.Writer.Size()).
				Dur("duration", duration),
			).Msg(logMessage)
	}
}

// Sanitize the request headers to remove any sensitive information
// Handles headers:
// - Authorization
func sanitizeHeaders(headers http.Header) http.Header {
	if _, ok := headers["Authorization"]; ok {
		headers["Authorization"] = []string{"*****"}
	}
	return headers
}
