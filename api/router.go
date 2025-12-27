package api

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/daylamtayari/cierge/api/middleware"
	"github.com/daylamtayari/cierge/internal/config"
)

func NewRouter(cfg *config.Config, logger zerolog.Logger) *gin.Engine {
	// Set gin mode based on environment
	if cfg.IsDevelopment() {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Global middleware
	router.Use(middleware.RequestID())
	router.Use(middleware.Logger())
	router.Use(middleware.CORS(cfg.Server.CORSOrigins))
	router.Use(middleware.Secure(cfg.IsDevelopment()))

	// Set trusted proxies to specified or nil, unless in dev
	// mode where it will trust all proxies (gin default, INSECURE)
	// NOTE: If you run cierge behind a proxy, you NEED to
	// specify trusted proxies
	if len(cfg.Server.TrustedProxies) > 0 {
		router.SetTrustedProxies(cfg.Server.TrustedProxies)
	} else if !cfg.IsDevelopment() {
		router.SetTrustedProxies(nil)
	}

	return router
}
