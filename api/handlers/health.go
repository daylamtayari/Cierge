package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	appctx "github.com/daylamtayari/cierge/internal/context"
)

type HealthResponse struct {
	Status string `json:"status"`
}

func Health(db *gorm.DB, timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := appctx.Logger(c.Request.Context())
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		// Check database connectivity
		sqlDB, err := db.DB()
		if err != nil {
			logger.Error().Err(err).Msg("health check: failed to get database connection")
			c.JSON(http.StatusServiceUnavailable, HealthResponse{
				Status: "unavailable",
			})
			return
		}
		if err := sqlDB.PingContext(ctx); err != nil {
			logger.Error().Err(err).Msg("health check: failed to ping database")
			c.JSON(http.StatusServiceUnavailable, HealthResponse{
				Status: "unavailable",
			})
			return
		}

		c.JSON(http.StatusOK, HealthResponse{
			Status: "ok",
		})
	}
}
