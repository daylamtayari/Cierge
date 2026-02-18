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

// Returns drop configs for a specified restaurant ordered by confidence in descending order
func (s *DropConfig) GetByRestaurant(ctx context.Context, restaurantId uuid.UUID) ([]*model.DropConfig, error) {
	return s.dcRepo.GetByRestaurant(ctx, restaurantId)
}

// Creates or reuses a drop config for the given restaurant.
// If a config with the same days_in_advance and drop_time already exists,
// the restaurant is associated with it and the existing config is returned.
func (s *DropConfig) Create(ctx context.Context, restaurantId uuid.UUID, daysInAdvance int16, dropTime string) (*model.DropConfig, error) {
	_, err := time.Parse("15:04", dropTime)
	if err != nil {
		return nil, ErrInvalidDropTime
	}

	if _, err := s.restaurantRepo.GetByID(ctx, restaurantId); err != nil {
		return nil, err
	}

	existing, err := s.dcRepo.GetByConfig(ctx, daysInAdvance, dropTime)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existing != nil {
		if err := s.dcRepo.AddRestaurant(ctx, existing.ID, restaurantId); err != nil {
			return nil, err
		}
		return existing, nil
	}

	user, ok := ctx.Value("user").(*model.User)
	if !ok {
		return nil, ErrUserNotInContext
	}

	dropConfig := model.DropConfig{
		DaysInAdvance: daysInAdvance,
		DropTime:      dropTime,
		CreatedBy:     &user.ID,
	}

	if err := s.dcRepo.Create(ctx, &dropConfig); err != nil {
		return nil, err
	}

	if err := s.dcRepo.AddRestaurant(ctx, dropConfig.ID, restaurantId); err != nil {
		return nil, err
	}

	return &dropConfig, nil
}

// Increment the confidence of a drop config for a specific restaurant
func (s *DropConfig) IncrementConfidence(ctx context.Context, dropConfigId uuid.UUID, restaurantId uuid.UUID) error {
	return s.dcRepo.IncrementConfidence(ctx, dropConfigId, restaurantId)
}
