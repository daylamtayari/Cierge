package api

import (
	"net/http"
	"time"

	"github.com/google/uuid"
)

type Restaurant struct {
	ID         uuid.UUID      `json:"id"`
	Platform   string         `json:"platform"`
	PlatformID string         `json:"platform_id"`
	Name       string         `json:"name"`
	Address    *string        `json:"address,omitempty"`
	City       *string        `json:"city,omitempty"`
	State      *string        `json:"state,omitempty"`
	Timezone   *time.Location `json:"timezone,omitempty"`
	Rating     *float32       `json:"rating,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

// Retrieves a restaurant with a given platform ID for a specified platform
func (c *Client) GetRestaurantByPlatform(platform string, platformId string) (Restaurant, error) {
	reqUrl := c.host + "/api/restaurant?platform=" + platform + "&platform-id=" + platformId
	req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
	if err != nil {
		return Restaurant{}, err
	}

	var restaurant Restaurant
	err = c.Do(req, &restaurant)
	if err != nil {
		return Restaurant{}, err
	}

	return restaurant, nil
}
