package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/daylamtayari/cierge/server/internal/config"
	"github.com/daylamtayari/cierge/server/internal/database"
	"github.com/daylamtayari/cierge/server/internal/logging"
	"github.com/daylamtayari/cierge/server/internal/repository"
	"github.com/daylamtayari/cierge/server/internal/router"
	"github.com/daylamtayari/cierge/server/internal/service"
	"github.com/daylamtayari/cierge/server/internal/version"
	"github.com/rs/zerolog"
	"github.com/spf13/pflag"

	// Registering of cloud providers
	_ "github.com/daylamtayari/cierge/server/internal/cloud/aws"
)

func main() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	if len(os.Args) > 1 && os.Args[1] == "version" {
		printVersion()
		return
	}

	devMode := pflag.Bool("dev", false, "Run in development mode (overrides config)")
	jsonOutput := pflag.Bool("json", false, "Force output logs to be JSON even during development mode")
	pflag.Parse()

	cfg, err := config.Load()
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to load config")
	}

	if *devMode {
		cfg.Environment = config.EnvironmentDev
		logger.Info().Msg("development mode enabled via command flag")
	}

	prettyOutput := cfg.IsDevelopment()
	if *jsonOutput {
		prettyOutput = false
	}

	logger = logging.New(cfg.LogLevel, prettyOutput).With().Str("environment", string(cfg.Environment)).Str("version", version.Version).Logger()
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

// Prints the version information
func printVersion() {
	ver := version.Version
	if ver == "dev" {
		if info, ok := debug.ReadBuildInfo(); ok {
			if info.Main.Version != "" && info.Main.Version != "(devel)" {
				ver = info.Main.Version
			}
		}
	}
	fmt.Printf("version %s", ver)
}
