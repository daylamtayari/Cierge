package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	appctx "github.com/daylamtayari/cierge/server/internal/context"
)

// Return an Unauthorized error response
func respondUnauthorized(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"error":      "Unauthorized",
		"request_id": appctx.RequestID(c.Request.Context()),
	})
}

// Return a Forbidden error response
func respondForbidden(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
		"error":      "Forbidden",
		"request_id": appctx.RequestID(c.Request.Context()),
	})
}

// Return an Internal Server Error response
func respondInternalServerError(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
		"error":      "Internal Server Error",
		"request_id": appctx.RequestID(c.Request.Context()),
	})
}
