package resy

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

// Represents a time slot available
// at a restaurant to book
// The Quantity field represents
// the amount of this time slot that
// are available to book
type Slot struct {
	Config   SlotConfig
	Date     SlotDate
	Quantity int
}

// Represents the configuration
// of the slot, and includes the
// slot token which is necessary
// for performing a booking
type SlotConfig struct {
	Id        int    `json:"id"`
	Token     string `json:"token"`
	Type      string `json:"type"`
	IsVisible bool   `json:"is_visible"`
}

// Represents the start and end
// datetime for a slot
// The start represents the reservation
// time, and end the expected end
// NOTE: The timezone value of the
// time fields is to UTC but the times
// are in local time
// e.g. 15:04 local time
// -> 15:04:00 +0000 UTC
type SlotDate struct {
	Start ResyDatetime `json:"start"`
	End   ResyDatetime `json:"end"`
}

// Represents the minimum and
// maximum size (amount of covers)
// for a particular slot
type SlotSize struct {
	Max int `json:"max"`
	Min int `json:"min"`
}

// Represents whether a slot
// requires a payment (deposit)
// and its details
// Pointers are required for all fields
// but the IsPaid field as if a restaurant
// does not have a deposit, Resy will still
// include this object and every field will
// be specified as `null`
type SlotPayment struct {
	IsPaid           bool     `json:"is_paid"`
	DepositFee       *float32 `json:"deposit_fee"`
	ServiceCharge    *string  `json:"service_charge"`
	VenueShare       *int     `json:"venue_share"`
	PaymentStructure *int     `json:"payment_structure"`
	SecsCancelCutOff *int     `json:"secs_cancel_cut_off"`
	TimeCancelCutOff *string  `json:"time_cancel_cut_off"`
	SecsChangeCutOff *int     `json:"secs_change_cut_off"`
	TimeChangeCutOff *string  `json:"time_change_cut_off"`
}

// Represents a slot's details that are
// returned when fetching a slot's details
// Numerous fields are returned but the booking
// token is the only thing handled at this point
// as a lot of the other fields contain
// repetitive data that is mostly unnecessary if
// you're not trying to render a frontend
type SlotDetails struct {
	BookingToken BookingToken `json:"book_token"`
	User         User         `json:"user"`
}

type BookingToken struct {
	Expiry time.Time `json:"date_expires"`
	Value  string    `json:"value"`
}

// Returns available slots for a specified venue ID, a Venue object, and an error that is nil if
// successful. If no venues are returned, `ErrNoVenues` is returned as the error.
func (c *Client) GetSlots(venueId int, day time.Time, partySize int) ([]Slot, *Venue, error) {
	type getSlotsRequest struct {
		Lat       int    `json:"lat"`
		Long      int    `json:"long"`
		Day       string `json:"day"`
		PartySize int    `json:"party_size"`
		VenueId   int    `json:"venue_id"`
	}

	reqUrl := Host + "/4/find"
	getSlotsReq := getSlotsRequest{
		Lat:       0,
		Long:      0,
		VenueId:   venueId,
		Day:       day.Format("2006-01-02"),
		PartySize: partySize,
	}
	reqBody, err := json.Marshal(getSlotsReq)
	if err != nil {
		return nil, nil, err
	}

	req, err := http.NewRequest(http.MethodPost, reqUrl, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, nil, err
	}

	type getSlotsResponse struct {
		Results struct {
			Venues []struct {
				Venue Venue  `json:"venue"`
				Slots []Slot `json:"slots"`
			} `json:"venues"`
		} `json:"results"`
	}

	var getSlotsRes getSlotsResponse
	err = c.Do(req, &getSlotsRes)
	if err != nil {
		return nil, nil, err
	}

	if len(getSlotsRes.Results.Venues) == 0 {
		return nil, nil, ErrNoVenues
	}
	return getSlotsRes.Results.Venues[0].Slots, &getSlotsRes.Results.Venues[0].Venue, nil
}

// Gets the details about a slot
// This creates a booking token that is valid for 5 minutes
// NOTE: Resy will allow you to get the slot details and create a booking
// token for a reservation that is not available. If a reservation is not available,
// you will get a 404 Not Found when trying to book
func (c *Client) GetSlotDetails(slotConfig string, day time.Time, partySize int) (*SlotDetails, error) {
	type getSlotDetailsRequest struct {
		ConfigId  string
		Day       string
		PartySize string
	}

	reqUrl := Host + "/3/details"
	getSlotDetailsReq := getSlotDetailsRequest{
		ConfigId:  slotConfig,
		Day:       day.Format("2006-01-02"),
		PartySize: strconv.Itoa(partySize),
	}
	reqBody, err := json.Marshal(getSlotDetailsReq)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, reqUrl, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	var slotDetails SlotDetails
	err = c.Do(req, &slotDetails)
	if err != nil {
		return nil, err
	}

	return &slotDetails, nil
}
