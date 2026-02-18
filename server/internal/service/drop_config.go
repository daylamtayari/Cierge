package service

import (
	"context"
	"errors"
	"time"

	"github.com/daylamtayari/cierge/server/internal/model"
	"github.com/daylamtayari/cierge/server/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrInvalidDropTime = errors.New("drop time is not in HH:mm format")
)

type DropConfig struct {
	dcRepo         *repository.DropConfig
	restaurantRepo *repository.Restaurant
}

func NewDropConfig(dropConfigRepo *repository.DropConfig, restaurantRepo *repository.Restaurant) *DropConfig {
	return &DropConfig{
		dcRepo:         dropConfigRepo,
		restaurantRepo: restaurantRepo,
	}
}

func (s *DropConfig) GetByRestaurant(ctx context.Context, restaurantId uuid.UUID) ([]*model.DropConfig, error) {
	return s.dcRepo.GetByRestaurant(ctx, restaurantId)
}

// Creates a new drop config but not before checking that it does not match an existing config
func (s *DropConfig) Create(ctx context.Context, restaurantId uuid.UUID, daysInAdvance int16, dropTime string) (*model.DropConfig, error) {
	_, err := time.Parse("15:04", dropTime)
	if err != nil {
		return nil, ErrInvalidDropTime
	}

	existing, err := s.dcRepo.GetByConfig(ctx, restaurantId, daysInAdvance, dropTime)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	restaurant, err := s.restaurantRepo.GetByID(ctx, restaurantId)
	if err != nil {
		return nil, err
	}

	user, ok := ctx.Value("user").(*model.User)
	if !ok {
		return nil, ErrUserNotInContext
	}

	dropConfig := model.DropConfig{
		Platform:      restaurant.Platform,
		RestaurantID:  restaurantId,
		DaysInAdvance: daysInAdvance,
		DropTime:      dropTime,
		CreatedBy:     &user.ID,
	}

	if restaurant.Timezone != nil {
		dropConfig.Timezone = restaurant.Timezone
	}

	return &dropConfig, s.dcRepo.Create(ctx, &dropConfig)
}

// Increment the confidence of a drop config
func (s *DropConfig) IncrementConfidence(ctx context.Context, configId uuid.UUID) error {
	return s.dcRepo.IncrementConfidence(ctx, configId)
}
