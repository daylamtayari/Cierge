package tokenstore

import "github.com/rs/zerolog"

type logger struct {
	log zerolog.Logger
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
