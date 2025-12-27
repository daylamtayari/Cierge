package database

import (
	"context"
	"errors"
	"time"

	"github.com/rs/zerolog"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	appctx "github.com/daylamtayari/cierge/internal/context"
)

type Logger struct {
	logger        zerolog.Logger
	slowThreshold time.Duration
	logLevel      gormlogger.LogLevel
}

func NewLogger(logger zerolog.Logger) *Logger {
	return &Logger{
		logger:        logger,
		slowThreshold: 200 * time.Millisecond,
		logLevel:      gormlogger.Info,
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
	appctx.Logger(ctx).Info().Msgf(msg, args...)
}

func (l *Logger) Warn(ctx context.Context, msg string, args ...any) {
	appctx.Logger(ctx).Warn().Msgf(msg, args...)
}

func (l *Logger) Error(ctx context.Context, msg string, args ...any) {
	appctx.Logger(ctx).Error().Msgf(msg, args...)
}

func (l *Logger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.logLevel == gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()
	logger := appctx.Logger(ctx)
	log := logger.With().Dur("duration", elapsed).Int64("rows", rows).Logger()

	switch {
	case err != nil && !errors.Is(err, gorm.ErrRecordNotFound):
		log.Error().Err(err).Str("sql", sql).Msg("query error")
	case elapsed > l.slowThreshold:
		log.Warn().Str("sql", sql).Msg("slow query")
	case l.logLevel >= gormlogger.Info:
		log.Debug().Str("sql", sql).Msg("query")
	}
}
