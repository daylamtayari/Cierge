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

type HealthService struct {
	db      *gorm.DB
	timeout time.Duration
}

func NewHealthService(db *gorm.DB, timeout time.Duration) *HealthService {
	return &HealthService{
		db:      db,
		timeout: timeout,
	}
}

func (s *HealthService) GetDBConnectivity(ctx context.Context) error {
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
