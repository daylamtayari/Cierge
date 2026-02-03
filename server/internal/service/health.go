package service

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

var (
	ErrDBConnectionFail = errors.New("failed to get database connection")
	ErrDBPingFail       = errors.New("failed to ping the database")
)

type Health struct {
	db      *gorm.DB
	timeout time.Duration
}

func NewHealth(db *gorm.DB, timeout time.Duration) *Health {
	return &Health{
		db:      db,
		timeout: timeout,
	}
}

func (s *Health) GetDBConnectivity(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	sqlDB, err := s.db.DB()
	if err != nil {
		return ErrDBConnectionFail
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		return ErrDBPingFail
	}
	return nil
}
