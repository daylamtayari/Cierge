package main

import (
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var logger zerolog.Logger

var (
	debugLog bool
	host     string

	rootCmd = &cobra.Command{
		Use:   "cierge",
		Short: "Cierge CLI",
		Long:  ``,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Skip persistent pre run for version retrieval
			if cmd.Name() == "version" {
				return nil
			}

			var err error
			cfg, err = initConfig()
			if err != nil {
				return err
			}

			logger = zerolog.New(os.Stderr).Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).With().Timestamp().Logger()
			if debugLog {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			} else {
				zerolog.SetGlobalLevel(zerolog.InfoLevel)
			}

			if cmd.Flags().Changed("host") {
				cfg.HostURL = host
			}
			if cfg.HostURL == "" {
				return fmt.Errorf("no server host specified")
			}

			return nil
		},
	}
)

func init() {
	rootCmd.PersistentFlags().BoolVar(&debugLog, "debug", false, "Enable debug logging")
	rootCmd.PersistentFlags().StringVar(&host, "host", "", "Override the server host")
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(initLoginCmd())
	rootCmd.AddCommand(initJobCmd())
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(versionCmd)
}
