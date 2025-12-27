package logging

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

func New(level zerolog.Level, isDevelopment bool) zerolog.Logger {
	zerolog.SetGlobalLevel(level)
	zerolog.TimestampFieldName = "timestamp"

	if isDevelopment {
		// Pretty output for when running in development
		return zerolog.New(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}).With().Timestamp().Logger()
	}
	return zerolog.New(os.Stdout).With().Timestamp().Logger()
}
