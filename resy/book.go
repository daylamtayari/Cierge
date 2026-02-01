package resy

import (
	"net/http"
	"net/url"
)

type BookingConfirmation struct {
	ReservationToken string `json:"resy_token"`
	ReservationId    string `json:"reservation_id"`
	VenueOptIn       bool   `json:"venue_opt_in"`
}

// Executes a reservation booking for a provided booking token
// A paymentMethodId representing the ID of the user's payment method
// should be passed if the restaurant requires a deposit or payment method
// on file, otherwise a 402 Payment Required will be returned
// If a reservation is no longer available, an ErrNotFound will be returned
// as the API returns a 404 in such cases
func (c *Client) BookReservation(bookingToken string, paymentMethodId *string) (*BookingConfirmation, error) {
	reqUrl := Host + "/3/book"

	reqForm := url.Values{
		"book_token": []string{bookingToken},
	}
	if paymentMethodId != nil {
		reqForm.Set("struct_payment_method", `{"id":"`+*paymentMethodId+`"}`)
	}

	req, err := c.NewFormRequest(http.MethodPost, reqUrl, &reqForm)
	if err != nil {
		return nil, err
	}

	var bookingConfirmation BookingConfirmation
	err = c.Do(req, &bookingConfirmation)
	if err != nil {
		return nil, err
	}

	return &bookingConfirmation, nil
}

// Cancels a specified booking and returns an error that is nil if successful
// A pointer to a byte slice can also be provided and the body of the response
// value will be unmarshalled into it
// The response body usually only returns a nested refund field that has value
// 1 to confirm that the refund will be issued, but uncertain if this response
// structure is consistent so keeping this designed as is
// NOTE: If the reservation token is invalid, it returns an Unauthorized error
// one more of those 'why...'
func (c *Client) CancelBooking(reservationToken string, body *[]byte) error {
	reqUrl := Host + "/3/cancel"

	reqForm := url.Values{
		"resy_token": []string{reservationToken},
	}

	req, err := c.NewFormRequest(http.MethodPost, reqUrl, &reqForm)
	if err != nil {
		return err
	}

	if body == nil {
		err = c.Do(req, nil)
	} else {
		err = c.Do(req, body)
	}

	if err != nil {
		return err
	}

	return nil
}
