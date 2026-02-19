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
	callbackAuthMiddleware := middleware.NewCallbackAuth(services.Job, services.Token)

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
		authRoutes.POST("/logout", authMiddleware.RequireAuth(), handlers.Auth.Logout)
		authRoutes.POST("/refresh", handlers.Auth.Refresh)
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
			users.GET("/token", handlers.PlatformToken.Get)
			users.POST("/token", handlers.PlatformToken.Create)
			users.POST("/api-key", handlers.User.APIKey)
			users.POST("/password", handlers.User.ChangePassword)
		}

		// Job routes
		jobs := api.Group("/job")
		{
			jobs.POST("", handlers.Job.Create)
			jobs.GET("/list", handlers.Job.List)
			jobs.POST("/:job/cancel", handlers.Job.Cancel)
		}

		// Restaurant route
		restaurants := api.Group("/restaurant")
		{
			restaurants.GET("", handlers.Restaurant.Get)
			restaurants.GET("/:id", handlers.Restaurant.GetByID)
		}

		// Drop config routes
		dropConfig := api.Group("/drop-config")
		{
			dropConfig.GET("", handlers.DropConfig.Get)
			dropConfig.POST("", handlers.DropConfig.Create)
		}

		// Admin routes
		admin := api.Group("/admin")
		admin.Use(authMiddleware.RequireAdmin())
		{
			admin.PUT("/user")
		}
	}

	return router
}
