package database

import (
	"context"
	"time"

	appctx "github.com/daylamtayari/cierge/internal/context"
	gormlogger "gorm.io/gorm/logger"
)

type Logger struct {
	logLevel gormlogger.LogLevel
}

func NewLogger(isDevelopment bool) *Logger {
	logLevel := gormlogger.Warn
	if isDevelopment {
		logLevel = gormlogger.Info
	}

	return &Logger{
		logLevel: logLevel,
	}
}

func (l *Logger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	return &Logger{
		logLevel: level,
	}
}

func (l *Logger) Info(ctx context.Context, msg string, args ...any) {
	// Silent - handled by Trace
}

func (l *Logger) Warn(ctx context.Context, msg string, args ...any) {
	// Silent - handled by Trace
}

func (l *Logger) Error(ctx context.Context, msg string, args ...any) {
	// Silent - handled by Trace
}

func (l *Logger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.logLevel == gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	// Add query to the query collector
	appctx.QueryCollector(ctx).Add(sql, elapsed, rows, err)
}
