package tokenstore

import (
	"runtime/debug"

	"github.com/rs/zerolog"
)

type logger struct {
	log zerolog.Logger
}

// Create a new logger for the token store that includes version info and sets level
func newLogger(log zerolog.Logger, isDevelopment bool) *logger {
	logLevel := zerolog.WarnLevel
	if isDevelopment {
		logLevel = zerolog.InfoLevel
	}

	var badgerVersion string
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, dep := range info.Deps {
			if dep.Path == "github.com/dgraph-io/badger/v4" {
				badgerVersion = dep.Version
				break
			}
		}
	}

	return &logger{
		log: log.With().Str("component", "token_store").Str("badger_version", badgerVersion).Logger().Level(logLevel),
	}
}

func (l *logger) Errorf(format string, args ...any) {
	l.log.Error().Msgf(format, args...)
}

func (l *logger) Warningf(format string, args ...any) {
	l.log.Warn().Msgf(format, args...)
}

func (l *logger) Infof(format string, args ...any) {
	l.log.Info().Msgf(format, args...)
}

func (l *logger) Debugf(format string, args ...any) {
	l.log.Debug().Msgf(format, args...)
}
