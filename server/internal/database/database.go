package database

import (
	"fmt"
	"time"

	"github.com/daylamtayari/cierge/server/internal/config"
	"github.com/daylamtayari/cierge/server/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func New(cfg config.DatabaseConfig, isDevelopment bool) (*gorm.DB, error) {
	gormCfg := &gorm.Config{
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		Logger: NewLogger(isDevelopment),
	}

	db, err := gorm.Open(postgres.Open(cfg.DSN()), gormCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the datbase: %w", err)
	}
	return db, nil
}

// AutoMigrate handles any additive database migrations
// NOTE: Does not handle deletions and updating existing
func AutoMigrate(db *gorm.DB) error {
	if err := createCustomTypes(db); err != nil {
		return fmt.Errorf("failed to create custom types: %w", err)
	}

	if err := db.AutoMigrate(
		&model.User{},
		&model.UserNotificationPreferences{},
		&model.PlatformToken{},
		&model.Job{},
		&model.Restaurant{},
		&model.DropConfig{},
		&model.Reservation{},
		&model.Favourite{},
		&model.Notification{},
		&model.Revocation{},
	); err != nil {
		return fmt.Errorf("failed to automigrate: %w", err)
	}
	return nil
}

// createCustomTypes creates PostgreSQL custom types (enums).
func createCustomTypes(db *gorm.DB) error {
	types := []string{
		`DO $$ BEGIN
			CREATE TYPE job_status AS ENUM ('scheduled', 'running', 'success', 'failed', 'cancelled');
		EXCEPTION
			WHEN duplicate_object THEN null;
		END $$`,
		`DO $$ BEGIN
			CREATE TYPE notification_type AS ENUM ('token_expiry', 'job_started', 'job_success', 'job_failed');
		EXCEPTION
			WHEN duplicate_object THEN null;
		END $$`,
	}

	for _, t := range types {
		if err := db.Exec(t).Error; err != nil {
			return err
		}
	}

	return nil
}
