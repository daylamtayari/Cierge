package middleware

import (
	"github.com/gin-gonic/gin"
)

// Sets the security headers used by the application
func Secure(isDevelopment bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("Referrer-Policy", "same-origin")
		c.Header("Content-Security-Policy", "default-src 'self'; object-src 'none'")

		// Only implement HSTS if not in development mode
		if !isDevelopment {
			c.Header("Strict-Transport-Security", "max-age=31536000")
		}

		c.Next()
	}
}
