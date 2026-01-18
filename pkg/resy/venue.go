package resy

import "errors"

var (
	ErrNoVenues = errors.New("no venues returned")
)

// Represents a restaurant venue
// NOTE: Not all fields may be populated
// on every request, the Resy API is
// inconsistent across endpoints and will
// not always include all the same fields
type Venue struct {
	Id              VenueId    `json:"id"`
	Group           VenueGroup `json:"group"`
	Name            string     `json:"name"`
	Type            string     `json:"type"`
	UrlSlug         string     `json:"url_slug"`
	PriceRange      int        `json:"price_range"`
	AverageBillSize int        `json:"average_bill_size"`
	TaxIncluded     bool       `json:"tax_included"`
	Rating          Rating     `json:"rating"`
	TotalRatings    int        `json:"total_ratings"`
	Favorite        bool       `json:"favorite"`
	CurrencyCode    string     `json:"currency_code"`
	CurrencySymbol  string     `json:"currency_symbol"`
	Country         string     `json:"country"`
	Region          string     `json:"region"`
	Locality        string     `json:"locality"`
	Neighborhood    string     `json:"neighborhood"`
	LastUpdatedAt   int        `json:"last_updated_at"`
}

// Represents a venue's IDs
// The Resy field is the Resy
// venue ID
type VenueId struct {
	Resy int `json:"resy"`
}

// Represents a restaurant
// group, including all of their
// restaurants' venue ID
type VenueGroup struct {
	Id     int    `json:"id"`
	Name   string `json:"name"`
	Venues []int  `json:"venues"`
}

// Represents the location of a venue,
// this is usually the city that the venue
// is in and not the exact location
// The Geo field however contains the
// coordinates of the restaurant itself
// (can sometimes have slight deviation)
type Location struct {
	Id           int         `json:"id"`
	Timezone     string      `json:"time_zone"`
	Neighborhood string      `json:"neighborhood"`
	Geo          GeoLocation `json:"geo"`
	Code         string      `json:"code"`
	UrlSlug      string      `json:"url_slug"`
}

// Represents coordinates to
// a specific location
type GeoLocation struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}
