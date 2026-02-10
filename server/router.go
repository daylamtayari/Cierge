package main

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/daylamtayari/cierge/server/internal/config"
	"github.com/daylamtayari/cierge/server/internal/handler"
	"github.com/daylamtayari/cierge/server/internal/middleware"
	"github.com/daylamtayari/cierge/server/internal/service"
)

func NewRouter(cfg *config.Config, logger zerolog.Logger, services *service.Services) *gin.Engine {
	// Set gin mode based on environment
	if cfg.IsDevelopment() {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Redirect gin's output to our logger
	ginLogger := logger.With().Str("component", "gin").Logger()
	gin.DefaultWriter = ginLogger
	gin.DefaultErrorWriter = ginLogger

	router := gin.New()

	// Global middleware
	router.Use(middleware.RequestID())
	router.Use(middleware.Logger(logger, cfg.IsDevelopment()))
	router.Use(middleware.CORS(cfg.Server.CORSOrigins))
	router.Use(middleware.Secure(cfg.IsDevelopment()))
	router.Use(middleware.Recovery())

	authMiddleware := middleware.NewAuth(services.Token, services.User)
	callbackAuthMiddleware := middleware.NewCallbackAuth(services.Job)

	handlers := handler.New(services, cfg)

	// Set trusted proxies to specified or nil, unless in dev
	// mode where it will trust all proxies (gin default, INSECURE)
	// NOTE: If you run cierge behind a proxy, you NEED to
	// specify trusted proxies
	if len(cfg.Server.TrustedProxies) > 0 {
		router.SetTrustedProxies(cfg.Server.TrustedProxies) //nolint:errcheck
	} else if !cfg.IsDevelopment() {
		router.SetTrustedProxies(nil) //nolint:errcheck
	}

	// Public routes
	router.GET("/health", handler.Health(services.Health))

	// Auth routes
	authRoutes := router.Group("/auth")
	{
		authRoutes.POST("/login", handlers.Auth.Login)
	}

	// Internal callback routes
	internalRoutes := router.Group("/internal")
	{
		internalRoutes.POST("/job/status", callbackAuthMiddleware.RequireCallbackAuth(), handlers.JobCallback.HandleJobCallback)
	}

	api := router.Group("/api")
	api.Use(authMiddleware.RequireAuth())
	{
		// User routes
		users := api.Group("/user")
		{
			users.GET("/me", handlers.User.Me)
			users.GET("/api-key", handlers.User.APIKey)
		}
	}

	return router
}
