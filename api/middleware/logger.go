package middleware

import (
	"time"

	"github.com/gin-gonic/gin"

	appctx "github.com/daylamtayari/cierge/internal/context"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		requestID := appctx.RequestID(c.Request.Context())

		// Create request-scoped logger
		baseLogger := appctx.Logger(c.Request.Context())
		logger := baseLogger.With().
			Str("request_id", requestID).
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Logger()

		logger.Debug().Msg("received request")

		ctx := appctx.WithLogger(c.Request.Context(), &logger)
		c.Request = c.Request.WithContext(ctx)

		c.Next()

		// Log request completion
		duration := time.Since(start)
		statusCode := c.Writer.Status()

		logEvent := logger.Info()
		if statusCode >= 500 {
			logEvent = logger.Error()
		} else if statusCode >= 400 {
			logEvent = logger.Warn()
		}

		// This project should be used at a small enough scale that these
		// log events become too significant to justify log sampling.
		logEvent.
			Int("status", statusCode).
			Dur("duration", duration).
			Msg("request completed")
	}
}
