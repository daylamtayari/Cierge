package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	appctx "github.com/daylamtayari/cierge/internal/context"
	"github.com/daylamtayari/cierge/internal/service"
)

func Health(healthService *service.HealthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		errorCol := appctx.ErrorCollector(c.Request.Context())

		err := healthService.GetDBConnectivity(c.Request.Context())
		if err != nil {
			errorCol.Add(err, zerolog.ErrorLevel, false, nil, "DB connectivity check failed")
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"status": "unavailable",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	}
}
