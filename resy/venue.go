package resy

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
)

var (
	ErrNoVenues = errors.New("no venues returned")
)

// Represents a restaurant venue
// NOTE: Not all fields may be populated
// on every request, the Resy API is
// inconsistent across endpoints and will
// not always include all the same fields
type Venue struct {
	Id              VenueId          `json:"id"`
	Group           VenueGroup       `json:"group"`
	Name            string           `json:"name"`
	Type            string           `json:"type"`
	UrlSlug         string           `json:"url_slug"`
	PriceRange      int              `json:"price_range"`
	AverageBillSize int              `json:"average_bill_size"`
	TaxIncluded     bool             `json:"tax_included"`
	Rating          Rating           `json:"rating"`
	TotalRatings    int              `json:"total_ratings"`
	Favorite        bool             `json:"favorite"`
	CurrencyCode    string           `json:"currency_code"`
	CurrencySymbol  string           `json:"currency_symbol"`
	Locale          VenueLocale      `json:"locale"`
	Location        VenueLocation    `json:"location"`
	Country         string           `json:"country"`
	Region          string           `json:"region"`
	Locality        string           `json:"locality"`
	Neighborhood    string           `json:"neighborhood"`
	Contact         VenueContact     `json:"contact"`
	Reopen          VenueReopen      `json:"reopen"`
	LastUpdatedAt   int              `json:"last_updated_at"`
	LeadTimeInDays  int              `json:"lead_time_in_days"`
	ServiceTypes    []KeyValueFilter `json:"service_types"`
	MinPartySize    int              `json:"min_party_size"`
	MaxPartySize    int              `json:"max_party_size"`
}

// Represents a venue's IDs
// The Resy field is the Resy
// venue ID
type VenueId struct {
	Resy   int    `json:"resy"`
	Google string `json:"google"`
}

// Represents a restaurant
// group, including all of their
// restaurants' venue ID
type VenueGroup struct {
	Id     int    `json:"id"`
	Name   string `json:"name"`
	Venues []int  `json:"venues"`
}

// Represents the locale of a
// venue including the timezone
// and currency
type VenueLocale struct {
	Currency string   `json:"currency"`
	Timezone Timezone `json:"time_zone"`
}

// Contact information of a restaurant
type VenueContact struct {
	PhoneNumber string `json:"phone_number"`
	Website     string `json:"url"`
}

// The reopening date of a venue
// This can be set if the restaurant is
// closed and has a planned reopening date,
// otherwise a nil value is present
type VenueReopen struct {
	Date *ResyDate `json:"date"`
}

// Represents the location of a venue,
// this is usually the city that the venue
// is in and not the exact location
// The Geo field however contains the
// coordinates of the restaurant itself
// (can sometimes have slight deviation)
type VenueLocation struct {
	Id           int         `json:"id"`
	Timezone     Timezone    `json:"time_zone"`
	Neighborhood string      `json:"neighborhood"`
	Geo          GeoLocation `json:"geo"`
	Code         string      `json:"code"`
	UrlSlug      string      `json:"url_slug"`
	Country      string      `json:"country"`
	CountryIso   string      `json:"country_iso3166"`
	Address1     *string     `json:"address_1"`
	Address2     *string     `json:"address_2"`
	PostalCode   string      `json:"postal_code"`
	Region       string      `json:"region"`
}

// Represents coordinates to
// a specific location
type GeoLocation struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

// Represents the inventory for
// a restaurant on a particular day
type CalendarSlot struct {
	Date      ResyDate      `json:"date"`
	Inventory SlotInventory `json:"inventory"`
}

// Represents the inventory of
// a reservation
// All of the fields are enums
// and have values of either
// "available", "not available",
// or "sold-out"
type SlotInventory struct {
	Reservation string `json:"reservation"`
	Event       string `json:"event"`
	WalkIn      string `json:"walk-in"`
}

// Default page limit for a venue search request
const defaultSearchPageLimit = 10

// Searches for a venue based on a specific query
func (c *Client) SearchVenue(query string, pageLimit *int) ([]Venue, error) {
	type searchVenueRequest struct {
		PageLimit int    `json:"per_page"`
		Query     string `json:"query"`
	}

	reqUrl := Host + "/3/venuesearch/search"
	searchVenueReq := searchVenueRequest{
		Query:     query,
		PageLimit: defaultSearchPageLimit,
	}
	if pageLimit != nil {
		searchVenueReq.PageLimit = *pageLimit
	}

	req, err := c.NewJsonRequest(http.MethodPost, reqUrl, searchVenueReq)
	if err != nil {
		return nil, err
	}

	type searchVenueResponse struct {
		Search struct {
			Hits []Venue
		}
	}

	var searchVenueRes searchVenueResponse
	err = c.Do(req, &searchVenueRes)
	if err != nil {
		return nil, err
	}

	// Although the API currently returns an empty slice which
	// is unmarshalled into a slice of len 0, adding this nil
	// check in case that behaviour ever changes to make sure
	// that this method's behaviour is accurate
	if searchVenueRes.Search.Hits != nil {
		searchVenueRes.Search.Hits = make([]Venue, 0)
	}

	return searchVenueRes.Search.Hits, nil
}

// Retrieves information about a specified venue
func (c *Client) GetVenue(venueId int) (*Venue, error) {
	reqUrl := Host + "/3/venue?id=" + strconv.Itoa(venueId)
	req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
	if err != nil {
		return nil, err
	}

	var venue Venue
	err = c.Do(req, &venue)
	if err != nil {
		return nil, err
	}

	return &venue, nil
}

// Gets the configuration for a venue and returns a Venue type with as
// much of the data returned. If a pointer to a Venue type is provided,
// that Venue type will be augmented with the data retrieved and any
// new data will overwrite existing data.
// The main use of this method is to retrieve the lead time in days for a
// reservation, which is how many days in advance reservations open
func (c *Client) GetVenueConfig(venueId int, providedVenue *Venue) (*Venue, error) {
	reqUrl := Host + "/2/config?venue_id=" + strconv.Itoa(venueId)
	req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
	if err != nil {
		return nil, err
	}

	type getVenueConfigResponse struct {
		LeadTimeInDays int `json:"lead_time_in_days"`
		Venue          struct {
			Name         string `json:"name"`
			MinPartySize int    `json:"min_party_size"`
			MaxPartySize int    `json:"max_party_size"`
		} `json:"venue"`
		ServiceTypes []KeyValueFilter `json:"service_types"`
	}

	var getVenueConfigRes getVenueConfigResponse
	err = c.Do(req, &getVenueConfigRes)
	if err != nil {
		return nil, err
	}

	venue := &Venue{
		Id: VenueId{
			Resy: venueId,
		},
	}
	if providedVenue != nil {
		venue = providedVenue
	}

	venue.LeadTimeInDays = getVenueConfigRes.LeadTimeInDays
	venue.MinPartySize = getVenueConfigRes.Venue.MinPartySize
	venue.MaxPartySize = getVenueConfigRes.Venue.MaxPartySize
	venue.Name = getVenueConfigRes.Venue.Name
	venue.ServiceTypes = getVenueConfigRes.ServiceTypes

	return venue, nil
}

// Retrieves a venue's inventory for a range of dates and for a specified number of seats
// This returns whether reservations, events, and walk ins are available, sold out, or unavailable on a given day
func (c *Client) GetVenueCalendar(venueId int, numSeats int, startDate ResyDate, endDate ResyDate) ([]CalendarSlot, error) {
	reqUrl, err := url.Parse(Host + "/4/venue/calendar")
	if err != nil {
		return nil, err
	}

	params := url.Values{
		"venue_id":   []string{strconv.Itoa(venueId)},
		"num_seats":  []string{strconv.Itoa(numSeats)},
		"start_date": []string{startDate.Format(ResyDateFormat)},
		"end_date":   []string{endDate.Format(ResyDateFormat)},
	}
	reqUrl.RawQuery = params.Encode()

	req, err := http.NewRequest(http.MethodGet, reqUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	type getVenueCalendarResponse struct {
		Scheduled []CalendarSlot `json:"scheduled"`
	}

	var getVenueCalendarRes getVenueCalendarResponse
	err = c.Do(req, &getVenueCalendarRes)
	if err != nil {
		return nil, err
	}

	return getVenueCalendarRes.Scheduled, nil
}
