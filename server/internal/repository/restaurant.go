package repository

import (
	"context"
	"time"

	"github.com/daylamtayari/cierge/server/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Restaurant struct {
	db      *gorm.DB
	timeout time.Duration
}

func NewRestaurant(db *gorm.DB, timeout time.Duration) *Restaurant {
	return &Restaurant{
		db:      db,
		timeout: timeout,
	}
}

// Get restaurant by ID
func (r *Restaurant) GetByID(ctx context.Context, id uuid.UUID) (*model.Restaurant, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	var restaurant model.Restaurant
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&restaurant).Error; err != nil {
		return nil, err
	}
	return &restaurant, nil
}

// Get restaurants from a given platform by their platform specific ID
func (r *Restaurant) GetByPlatformID(ctx context.Context, platform string, platformID string) (*model.Restaurant, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	var restaurant model.Restaurant
	if err := r.db.WithContext(ctx).Where("platform = ?", platform).Where("platform_id = ?", platformID).First(&restaurant).Error; err != nil {
		return nil, err
	}
	return &restaurant, nil
}

// Create restaurant
func (r *Restaurant) Create(ctx context.Context, restaurant *model.Restaurant) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	return r.db.WithContext(ctx).Create(restaurant).Error
}

// Update restuarant
func (r *Restaurant) Update(ctx context.Context, restaurant *model.Restaurant) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	return r.db.WithContext(ctx).Save(restaurant).Error
}

// Delete restaurant
func (r *Restaurant) Delete(ctx context.Context, id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	return r.db.WithContext(ctx).Delete(&model.Restaurant{}, "id = ?", id).Error
}
