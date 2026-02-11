package api

import "net/http"

type Health struct {
	Status string
	Server string
}

// Retrieve the health of the server
func (c *Client) GetHealth() (Health, error) {
	reqUrl := c.host + "/health"
	req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
	if err != nil {
		return Health{}, err
	}

	var health Health
	err = c.Do(req, &health)
	if err != nil {
		return Health{}, err
	}

	return health, nil
}
