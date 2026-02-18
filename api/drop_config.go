package api

import (
	"net/http"
	"time"

	"github.com/google/uuid"
)

type DropConfig struct {
	ID uuid.UUID `json:"id"`

	DaysInAdvance int16  `json:"days_in_advance"`
	DropTime      string `json:"drop_time"`
	Confidence    int16  `json:"confidence"`

	CreatedAt time.Time `json:"created_at"`
}

type dropConfigCreateRequest struct {
	Restaurant    uuid.UUID
	DaysInAdvance int16
	DropTime      string
}

// Retrieves all drop configs for a specified restaurant, ordered
// by confidence in descending order
// The confidence value represents how many times a drop config has
// been used to schedule a job for the given restaurant
// If none exist, an empty slice is returned
func (c *Client) GetDropConfigs(restaurantId uuid.UUID) ([]DropConfig, error) {
	reqUrl := c.host + "/api/drop-config?restaurant=" + restaurantId.String()
	req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
	if err != nil {
		return nil, err
	}

	var dropConfigs []DropConfig
	err = c.Do(req, &dropConfigs)
	if err != nil {
		return nil, err
	}

	return dropConfigs, nil
}

// Creates a new drop config and returns the drop config object and an error that is nil if successful
// NOTE: The drop time value must be in HH:mm format otherwise the server will return an error
func (c *Client) CreateDropConfig(restaurantId uuid.UUID, daysInAdvance int16, dropTime string) (DropConfig, error) {
	reqUrl := c.host + "/api/drop-config"
	dropConfigCreateReq := dropConfigCreateRequest{
		Restaurant:    restaurantId,
		DaysInAdvance: daysInAdvance,
		DropTime:      dropTime,
	}
	req, err := c.NewJsonRequest(http.MethodPost, reqUrl, dropConfigCreateReq)
	if err != nil {
		return DropConfig{}, err
	}

	var dropConfig DropConfig
	err = c.Do(req, &dropConfig)
	if err != nil {
		return DropConfig{}, err
	}

	return dropConfig, nil
}
