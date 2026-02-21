package service

import (
	"context"
	"errors"
	"time"

	appctx "github.com/daylamtayari/cierge/server/internal/context"
	"github.com/daylamtayari/cierge/server/internal/model"
	"github.com/daylamtayari/cierge/server/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrDropConfigDNE   = errors.New("drop config does not exist")
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

// Gets a drop config from its ID
func (s *DropConfig) GetByID(ctx context.Context, dropConfigId uuid.UUID) (*model.DropConfig, error) {
	dropConfig, err := s.dcRepo.GetByID(ctx, dropConfigId)
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrDropConfigDNE
	} else if err != nil {
		return nil, err
	}
	return dropConfig, nil
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

	userID := appctx.UserID(ctx)

	dropConfig := model.DropConfig{
		DaysInAdvance: daysInAdvance,
		DropTime:      dropTime,
		CreatedBy:     &userID,
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

// Returns true if the scheduled job time would be in the past
func (s *DropConfig) IsScheduledAtPast(dropConfig *model.DropConfig, reservationDate time.Time, restaurantTimezone *time.Location) bool {
	scheduledAtDate := reservationDate.Add(-time.Duration(dropConfig.DaysInAdvance) * 24 * time.Hour)
	scheduledAtTime, _ := time.Parse("15:04", dropConfig.DropTime)
	scheduledAtLoc := time.UTC
	if restaurantTimezone != nil {
		scheduledAtLoc = restaurantTimezone
	}
	scheduledAt := time.Date(scheduledAtDate.Year(), scheduledAtDate.Month(), scheduledAtDate.Day(), scheduledAtTime.Hour(), scheduledAtTime.Minute(), 0, 0, scheduledAtLoc)
	return time.Now().After(scheduledAt)
}
