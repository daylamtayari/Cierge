package database

import (
	"context"
	"errors"
	"time"

	"github.com/rs/zerolog"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type Logger struct {
	logger        zerolog.Logger
	slowThreshold time.Duration
	logLevel      gormlogger.LogLevel
}

func NewLogger(logger zerolog.Logger, isDevelopment bool) *Logger {
	logLevel := gormlogger.Warn
	if isDevelopment {
		logLevel = gormlogger.Info
	}
	return &Logger{
		logger:        logger,
		slowThreshold: 200 * time.Millisecond,
		logLevel:      logLevel,
	}
}

func (l *Logger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	return &Logger{
		logger:        l.logger,
		slowThreshold: l.slowThreshold,
		logLevel:      level,
	}
}

func (l *Logger) Info(ctx context.Context, msg string, args ...any) {
	l.logger.Info().Msgf(msg, args...)
}

func (l *Logger) Warn(ctx context.Context, msg string, args ...any) {
	l.logger.Warn().Msgf(msg, args...)
}

func (l *Logger) Error(ctx context.Context, msg string, args ...any) {
	l.logger.Error().Msgf(msg, args...)
}

func (l *Logger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.logLevel == gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()
	log := l.logger.With().Dur("duration", elapsed).Int64("rows", rows).Logger()

	switch {
	case err != nil && !errors.Is(err, gorm.ErrRecordNotFound):
		log.Error().Err(err).Str("sql", sql).Msg("query error")
	case elapsed > l.slowThreshold:
		log.Warn().Str("sql", sql).Msg("slow query")
	case l.logLevel >= gormlogger.Info:
		log.Debug().Str("sql", sql).Msg("query")
	}
}
