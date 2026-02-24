package opentable

type SearchResult struct {
	Id               int     `json:"id"`
	Type             string  `json:"type"`
	LocationSubtype  *string `json:"locationSubtype"` //TODO: Verify
	Name             string  `json:"name"`
	SearchCenterType *string `json:"searchCenterType"` //TODO: Verify
	Country          string  `json:"country"`
	CountryId        *int    `json:"countryId"`
	MetroId          int     `json:"metroId"`
	MetroName        string  `json:"metroName"`
	MacroId          int     `json:"macroId"`
	MacroName        string  `json:"macroName"`
	Neighborhood     string  `json:"neighborhoodName"`
	Latitude         float64 `json:"latitude"`
	Longitude        float64 `json:"longitude"`
}
