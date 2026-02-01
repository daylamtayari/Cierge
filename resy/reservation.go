package resy

import (
	"net/http"
	"net/url"
	"strconv"
)

// Represents a Resy reservation
// NOTE: When field datetime is in UTC
type Reservation struct {
	ReservationId        int               `json:"reservation_id"`
	ReservationToken     string            `json:"resy_token"`
	Day                  ResyDate          `json:"day"`
	When                 ResyDatetime      `json:"when"`
	NumSeats             int               `json:"num_seats"`
	TimeSlot             ResyTime          `json:"time_slot"`
	ServiceTypeId        int               `json:"service_type_id"`
	Occasion             *string           `json:"occasion"`
	Price                float32           `json:"price"`
	Status               ReservationStatus `json:"status"`
	Venue                ReservationVenue  `json:"venue"`
	Config               ReservationConfig `json:"config"`
	IsPickup             bool              `json:"is_pickup"`
	AddOnsAvailable      bool              `json:"add_ons_available"`
	IsGlobalDiningAccess bool              `json:"is_global_dining_access"`
}

// Represents a party member of a reservation
// The only field populated in the User field
// is the party member's email address
type ReservationPartyMember struct {
	Token     string `json:"token"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	User      User   `json:"user"`
	Label     string `json:"label"`
	Type      string `json:"type"`
	Status    string `json:"status"`
}

// Represents the status of a reservation
// Finished has a value of 1 if past reservation
type ReservationStatus struct {
	Finished int `json:"finished"`
	NoShow   int `json:"no_show"`
}

// Represents a reservation's config object
// The Type field represents the type of
// seating for the reservation
// (e.g. "Dining Room", "Table", "Patio")
type ReservationConfig struct {
	BackgroundColor string `json:"background_color"`
	FontColor       string `json:"font_color"`
	Type            string `json:"type"`
}

// Represents an occasion for a reservation
type ReservationOccasion struct {
	Occasion   string `json:"occasion"`
	OccasionId string `json:"occasion_id"`
}

// Represents a reservation's venue
type ReservationVenue struct {
	VenueId  int    `json:"id"`
	Currency string `json:"currency"`
}

var (
	AnniversaryOccasion = ReservationOccasion{
		Occasion:   "Anniversary",
		OccasionId: "149470",
	}
	BirthdayOccasion = ReservationOccasion{
		Occasion:   "Birthday",
		OccasionId: "149474",
	}
	BusinessOccasion = ReservationOccasion{
		Occasion:   "Business",
		OccasionId: "149476",
	}
	GraduationOccasion = ReservationOccasion{
		Occasion:   "Graduation",
		OccasionId: "149480",
	}
	NoOccasion = ReservationOccasion{
		Occasion:   "",
		OccasionId: "",
	}
)

// Represents the type of a reservation
// (past or upcoming) used for queries
type ReservationType string

const (
	PastReservation     = ReservationType("past")
	UpcomingReservation = ReservationType("upcoming")
)

// Default limit parameter for the GetReservations method
// If you want to get more than 100,000 reservations
// at once, specify a custom limit value
const defaultReservationLimit = 100000

// Returns a slice of reservations
// If no reservations were found, a slice of length 0 is returned
// A reservation type parameter can be specified to fetchh only past or
// only upcoming reservations. A on behalf parameter can be specified to
// only retrieve reservations that include other people on the reservation.
// A limit and offset field can also be provided to limit the size of the
// response but if no limit is specified, a default limit of 100,000 is set.
// Similarly, if no offset is specified, an offset of 0 is specified
func (c *Client) GetReservations(reservationType *ReservationType, onBehalf *bool, limit *int, offset *int) ([]Reservation, error) {
	reqUrl, err := url.Parse(Host + "/3/user/reservations")
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	if reservationType != nil {
		params.Add("type", string(*reservationType))
	}
	if onBehalf != nil {
		params.Add("book_on_behalf_of", strconv.FormatBool(*onBehalf))
	}
	if limit != nil {
		params.Add("limit", strconv.Itoa(*limit))
	} else {
		params.Add("limit", strconv.Itoa(defaultReservationLimit))
	}
	if offset != nil {
		params.Add("offset", strconv.Itoa(*offset))
	}

	reqUrl.RawQuery = params.Encode()

	req, err := http.NewRequest(http.MethodGet, reqUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	type getReservationsResponse struct {
		Reservations []Reservation `json:"reservations"`
	}

	var getReservationsRes getReservationsResponse
	err = c.Do(req, &getReservationsRes)
	if err != nil {
		return nil, err
	}

	// Although the API currently returns an empty slice which
	// is unmarshalled into a slice of len 0, adding this nil
	// check in case that behaviour ever changes to make sure
	// that this method's behaviour is accurate
	if getReservationsRes.Reservations == nil {
		getReservationsRes.Reservations = make([]Reservation, 0)
	}
	return getReservationsRes.Reservations, nil
}

// Sets the reservation occasion for a specified reservation
func (c *Client) SetReservationOccasion(reservationToken string, occasion ReservationOccasion) error {
	type setReservationOccasionRequest struct {
		ReservationOccasion
		ReservationToken string `json:"resy_token"`
	}

	reqUrl := Host + "/2/reservation/special_request"
	setReservationOccasionReq := setReservationOccasionRequest{
		ReservationOccasion: occasion,
		ReservationToken:    reservationToken,
	}

	req, err := c.NewJsonRequest(http.MethodPost, reqUrl, setReservationOccasionReq)
	if err != nil {
		return err
	}

	err = c.Do(req, nil)
	if err != nil {
		return err
	}

	return nil
}

// Sets a special request for a specified reservation
func (c *Client) SetReservationSpecialRequest(reservationToken string, specialRequest string) error {
	type setReservationSpecialRequestRequest struct {
		ReservationToken string `json:"resy_token"`
		Description      string `json:"description"`
	}

	reqUrl := Host + "/2/reservation/special_request"
	setReservationSpecialRequestReq := setReservationSpecialRequestRequest{
		ReservationToken: reservationToken,
		Description:      specialRequest,
	}

	req, err := c.NewJsonRequest(http.MethodPost, reqUrl, setReservationSpecialRequestReq)
	if err != nil {
		return err
	}

	err = c.Do(req, nil)
	if err != nil {
		return err
	}

	return nil
}
