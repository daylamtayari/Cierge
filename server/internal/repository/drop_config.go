package repository

import (
	"context"
	"time"

	"github.com/daylamtayari/cierge/server/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

// Get all drop configs for a specified restaurant ID, ordered by confidence descending
func (r *DropConfig) GetByRestaurant(ctx context.Context, restaurantId uuid.UUID) ([]*model.DropConfig, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	var dropConfigs []*model.DropConfig
	err := r.db.WithContext(ctx).
		Select("drop_configs.*, drop_config_restaurants.confidence").
		Joins("JOIN drop_config_restaurants ON drop_config_restaurants.drop_config_id = drop_configs.id").
		Where("drop_config_restaurants.restaurant_id = ?", restaurantId).
		Order("drop_config_restaurants.confidence DESC").
		Find(&dropConfigs).Error
	return dropConfigs, err
}

// Gets a drop config matching the given days in advance and drop time
func (r *DropConfig) GetByConfig(ctx context.Context, daysInAdvance int16, dropTime string) (*model.DropConfig, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	var dropConfig model.DropConfig
	err := r.db.WithContext(ctx).
		Where("days_in_advance = ? AND drop_time = ?", daysInAdvance, dropTime).
		Take(&dropConfig).Error
	return &dropConfig, err
}

// Create a drop config
func (r *DropConfig) Create(ctx context.Context, dropConfig *model.DropConfig) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	return r.db.WithContext(ctx).Create(dropConfig).Error
}

// AddRestaurant associates a restaurant with a drop config.
// If the association already exists it is a no-op.
func (r *DropConfig) AddRestaurant(ctx context.Context, dropConfigId uuid.UUID, restaurantId uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&model.DropConfigRestaurant{
			DropConfigID: dropConfigId,
			RestaurantID: restaurantId,
		}).Error
}

// Increments the confidence for the association between a drop config and a restaurant
func (r *DropConfig) IncrementConfidence(ctx context.Context, dropConfigId uuid.UUID, restaurantId uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	return r.db.WithContext(ctx).Model(&model.DropConfigRestaurant{}).
		Where("drop_config_id = ? AND restaurant_id = ?", dropConfigId, restaurantId).
		Updates(map[string]any{
			"confidence": gorm.Expr("confidence + 1"),
		}).Error
}
