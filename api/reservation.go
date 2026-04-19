package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type Reservation struct {
	ID           uuid.UUID  `json:"id"`
	UserID       uuid.UUID  `json:"user_id"`
	JobID        *uuid.UUID `json:"job_id,omitempty"`
	RestaurantID uuid.UUID  `json:"restaurant_id"`

	Platform      string    `json:"platform"`
	Confirmation  *string   `json:"confirmation,omitempty"`
	ReservationAt time.Time `json:"reservation_at"`
	PartySize     int16     `json:"party_size"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Retrieve a given reservation
func (c *Client) GetReservation(reservationId uuid.UUID) (Reservation, error) {
	reqUrl := c.host + "/api/reservation/" + reservationId.String()
	req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
	if err != nil {
		return Reservation{}, err
	}

	var res Reservation
	err = c.Do(req, &res)
	if err != nil {
		return Reservation{}, err
	}

	return res, nil
}

// Retrieve reservations for the user
// If upcomingOnly is set to true, only upcoming reservations are returned
func (c *Client) GetReservations(upcomingOnly bool) ([]Reservation, error) {
	reqUrl := c.host + "/api/reservation/list?upcoming=" + strconv.FormatBool(upcomingOnly)
	req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
	if err != nil {
		return nil, err
	}

	var reservations []Reservation
	err = c.Do(req, &reservations)
	if err != nil {
		return nil, err
	}

	return reservations, nil
}
