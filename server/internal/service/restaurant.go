package service

import (
	"context"
	"errors"
	"strconv"

	"github.com/daylamtayari/cierge/resy"
	"github.com/daylamtayari/cierge/server/internal/model"
	"github.com/daylamtayari/cierge/server/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrRestaurantDNE = errors.New("restaurant does not exist")
)

type Restaurant struct {
	restaurantRepo *repository.Restaurant
	resyClient     *resy.Client
}

func NewRestaurant(restaurantRepo *repository.Restaurant, resyClient *resy.Client) *Restaurant {
	return &Restaurant{
		restaurantRepo: restaurantRepo,
		resyClient:     resyClient,
	}
}

// Retrieves a restaurant from its ID
func (s *Restaurant) GetByID(ctx context.Context, restaurantId uuid.UUID) (*model.Restaurant, error) {
	restaurant, err := s.restaurantRepo.GetByID(ctx, restaurantId)
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrReservationDNE
	} else if err != nil {
		return nil, err
	}
	return restaurant, nil
}

// Retrieves a restaurant from its platform ID
func (s *Restaurant) GetByPlatformID(ctx context.Context, platform string, platformID string) (*model.Restaurant, error) {
	restaurant, err := s.restaurantRepo.GetByPlatformID(ctx, platform, platformID)
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrRestaurantDNE
	} else if err != nil {
		return nil, err
	}
	return restaurant, nil
}

// Create a restaurant
func (s *Restaurant) Create(ctx context.Context, platform string, platformID string) (*model.Restaurant, error) {
	restaurant := model.Restaurant{
		Platform:   platform,
		PlatformID: platformID,
	}
	switch platform {
	case "resy":
		resyVenueId, err := strconv.Atoi(platformID)
		if err != nil {
			return nil, err
		}
		venue, err := s.resyClient.GetVenue(resyVenueId)
		if err != nil {
			return nil, err
		}

		restaurant.Name = venue.Name
		restaurant.Timezone = &model.Timezone{Location: venue.Locale.Timezone.Location}

		var addressStr string
		if venue.Location.Address1 != nil {
			addressStr = *venue.Location.Address1
		}
		if venue.Location.Address2 != nil {
			if addressStr != "" {
				addressStr += " " + *venue.Location.Address2
			} else {
				addressStr = *venue.Location.Address2
			}
		}
		if addressStr != "" {
			restaurant.Address = &addressStr
		}
		restaurant.City = &venue.Location.Neighborhood
		restaurant.State = &venue.Location.Region

	case "opentable":
		// TODO: Implement opentable
	}

	if err := s.restaurantRepo.Create(ctx, &restaurant); err != nil {
		return nil, err
	}

	return &restaurant, nil
}
