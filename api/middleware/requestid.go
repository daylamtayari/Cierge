package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	appctx "github.com/daylamtayari/cierge/internal/context"
)

const RequestIDHeader = "X-Request-ID"

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := uuid.NewString()
		c.Header(RequestIDHeader, requestID)

		ctx := appctx.WithRequestID(c.Request.Context(), requestID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
