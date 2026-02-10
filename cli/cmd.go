package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

type contextKey string

var (
	configContextKey = contextKey("config")
	loggerContextKey = contextKey("logger")
)

var rootCmd = &cobra.Command{
	Use:   "cierge",
	Short: "Cierge CLI",
	Long:  ``,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip persistent pre run for version retrieval
		if cmd.Name() == "version" {
			return nil
		}

		config, err := initConfig()
		if err != nil {
			return err
		}
		cmd.SetContext(context.WithValue(cmd.Context(), configContextKey, config))

		logger := zerolog.New(os.Stderr).Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).With().Timestamp().Logger()
		debug, err := cmd.Flags().GetBool("debug")
		if err != nil {
			return err
		}
		if debug {
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		} else {
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
		}
		cmd.SetContext(context.WithValue(cmd.Context(), loggerContextKey, &logger))

		if config.HostURL == "" {
			return fmt.Errorf("no Cierge host specified")
		}
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().Bool("debug", false, "Enable debug logging")
	rootCmd.AddCommand(versionCmd)
}

func getConfig(cmd *cobra.Command) *config {
	return cmd.Context().Value(configContextKey).(*config)
}

func getLogger(cmd *cobra.Command) *zerolog.Logger {
	return cmd.Context().Value(loggerContextKey).(*zerolog.Logger)
}
