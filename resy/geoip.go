package resy

import "net/http"

type Geoip struct {
	Ip          string  `json:"ip"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	CountryCode string  `json:"country_iso_code"`
	InEU        bool    `json:"in_eu"`
	Source      string  `json:"source"`
	Success     bool    `json:"success"`
}

// Retrieves the geoip information about the
// IP that is making the request
// NOTE: This is a very light request that only
// requires the API key and not user auth
func (c *Client) GetGeoip() (*Geoip, error) {
	reqUrl := Host + "/3/geoip"
	req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
	if err != nil {
		return nil, err
	}

	var geoip Geoip
	err = c.Do(req, &geoip)
	if err != nil {
		return nil, err
	}

	return &geoip, nil
}
