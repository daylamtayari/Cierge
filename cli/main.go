package main

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		logger := zerolog.New(os.Stderr).Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).With().Timestamp().Logger()
		logger.Fatal().Err(err).Msg("Root command encountered an error")
	}
}
