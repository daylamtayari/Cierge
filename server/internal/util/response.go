package util

import (
	"net/http"

	"github.com/gin-gonic/gin"

	appctx "github.com/daylamtayari/cierge/server/internal/context"
)

// Return a Bad Request error response with a custom message
func RespondBadRequest(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
		"error":      "Bad Requeest",
		"message":    message,
		"request_id": appctx.RequestID(c.Request.Context()),
	})
}

// Return an Unauthorized error response
func RespondUnauthorized(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"error":      "Unauthorized",
		"request_id": appctx.RequestID(c.Request.Context()),
	})
}

// Return a Forbidden error response
func RespondForbidden(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
		"error":      "Forbidden",
		"request_id": appctx.RequestID(c.Request.Context()),
	})
}

// Returns a not found error message
func RespondNotFound(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
		"error":      "Not Found",
		"message":    message,
		"request_id": appctx.RequestID(c.Request.Context()),
	})
}

// Return an Internal Server Error response
func RespondInternalServerError(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
		"error":      "Internal Server Error",
		"request_id": appctx.RequestID(c.Request.Context()),
	})
}
