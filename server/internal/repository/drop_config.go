package repository

import (
	"context"
	"time"

	"github.com/daylamtayari/cierge/server/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DropConfig struct {
	db      *gorm.DB
	timeout time.Duration
}

func NewDropConfig(db *gorm.DB, timeout time.Duration) *DropConfig {
	return &DropConfig{
		db:      db,
		timeout: timeout,
	}
}

// Gets a drop config with a given ID
func (r *DropConfig) GetByID(ctx context.Context, id uuid.UUID) (*model.DropConfig, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	var dropConfig model.DropConfig
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&dropConfig).Error; err != nil {
		return nil, err
	}
	return &dropConfig, nil
}

// Get all drop configs for a specified restaurant ID
func (r *DropConfig) GetByRestaurant(ctx context.Context, restaurantId uuid.UUID) ([]*model.DropConfig, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	var dropConfigs []*model.DropConfig
	err := r.db.WithContext(ctx).Where("restaurant_id = ?", restaurantId).Find(&dropConfigs).Error
	return dropConfigs, err
}

// Gets a drop config for a specified restaurant with the same drop time and amount of days in advance
func (r *DropConfig) GetByConfig(ctx context.Context, restaurantId uuid.UUID, daysInAdvance int16, dropTime string) (*model.DropConfig, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	var dropConfig model.DropConfig
	err := r.db.WithContext(ctx).Where("restaurant_id = ?", restaurantId).Where("days_in_advance = ?", daysInAdvance).Where("drop_time", dropTime).First(&dropConfig).Error
	return &dropConfig, err
}

// Create a drop config
func (r *DropConfig) Create(ctx context.Context, dropConfig *model.DropConfig) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	return r.db.WithContext(ctx).Create(dropConfig).Error
}

// Increments the confidence of a specified drop config
func (r *DropConfig) IncrementConfidence(ctx context.Context, id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	return r.db.WithContext(ctx).Model(&model.DropConfig{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"confidence": gorm.Expr("confidence + 1"),
		}).Error
}
