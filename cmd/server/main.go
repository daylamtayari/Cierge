package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/daylamtayari/cierge/internal/config"
	"github.com/daylamtayari/cierge/internal/database"
	"github.com/daylamtayari/cierge/internal/logging"
	"github.com/daylamtayari/cierge/internal/repository"
	"github.com/daylamtayari/cierge/internal/router"
	"github.com/daylamtayari/cierge/internal/service"
	"github.com/daylamtayari/cierge/internal/version"
	"github.com/rs/zerolog"
)

func main() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	cfg, err := config.Load()
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to load config")
	}

	logger = logging.New(cfg.LogLevel, cfg.IsDevelopment()).With().Str("environment", string(cfg.Environment)).Str("version", version.Version).Logger()
	logger.Info().Msg("starting cierge server")

	db, err := database.New(cfg.Database, cfg.IsDevelopment())
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to connect to the database")
	}
	logger.Debug().Msg("connected to the database")
	if cfg.Database.AutoMigrate {
		logger.Debug().Msg("running database migrations")
		if err := database.AutoMigrate(db); err != nil {
			logger.Fatal().Err(err).Msg("failed to run database migrations")
		}
		logger.Info().Msg("database migrations completed successfully")
	}

	repos := repository.New(db, cfg.Database.Timeout.Duration())
	services := service.New(repos, cfg)

	// Handle default admin user creation if no users exist
	userCount, err := services.User.GetUserCount(context.Background())
	if err != nil {
		logger.Error().Err(err).Msg("Failed to retrieve user count")
	} else if userCount == 0 {
		defaultAdmin := cfg.DefaultAdmin
		hashedPassword := services.Auth.HashPassword(defaultAdmin.Password)
		_, err = services.User.Create(context.Background(), defaultAdmin.Email, hashedPassword, true)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to create default admin user")
			logger.Warn().Msg("Server does not have any valid users")
		}
	}

	router := router.NewRouter(cfg, logger, services)
	server := &http.Server{
		Addr:    cfg.Server.Address(),
		Handler: router,
	}

	// Use a goroutine to run the server to handle graceful shutdowns with a context
	// This application is expected to be ran in a containerised environment
	// and as such, to shutdown generally an interrupt signal will be sent.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	serverErrors := make(chan error, 1)
	go func() {
		logger.Info().Str("address", cfg.Server.Address()).Msg("starting http server")
		var err error
		if cfg.Server.TLS.Enabled {
			err = server.ListenAndServeTLS(cfg.Server.TLS.CertFile, cfg.Server.TLS.KeyFile)
		} else {
			err = server.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			serverErrors <- err
		}
	}()

	// If server failed to start, read the error, otherwise wait for the shutdown signal
	select {
	case err := <-serverErrors:
		stop()
		logger.Error().Err(err).Msg("error serving http server")
	case <-ctx.Done():
		stop()
		logger.Warn().Msg("shutdown signal received")
		logger.Warn().Msg("shutting server down")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error().Err(err).Msg("graceful shutdown failed")
			if closeErr := server.Close(); closeErr != nil {
				logger.Error().Err(closeErr).Msg("force closure failed")
			}
		}
		logger.Info().Msg("server exiting")
	}
}
