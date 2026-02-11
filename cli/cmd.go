package main

import (
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var logger zerolog.Logger

var rootCmd = &cobra.Command{
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
		debug, err := cmd.Flags().GetBool("debug")
		if err != nil {
			return err
		}
		if debug {
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		} else {
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
		}

		if cfg.HostURL == "" {
			return fmt.Errorf("no Cierge host specified")
		}
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().Bool("debug", false, "Enable debug logging")
	rootCmd.AddCommand(initLoginCmd())
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(versionCmd)
}
