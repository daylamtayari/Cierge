package resy

import (
	"bytes"
	"encoding/json"
	"net/http"
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

// Returns available slots for a specified venue ID, a Venue object, and an error that is nil if
// successful. If no venues are returned, `ErrNoVenues` is returned as the error.
func (c *Client) GetSlots(venueId int, day time.Time, partySize int) ([]Slot, *Venue, error) {
	type getSlotRequest struct {
		Lat       int    `json:"lat"`
		Long      int    `json:"long"`
		Day       string `json:"day"`
		PartySize int    `json:"party_size"`
		VenueId   int    `json:"venue_id"`
	}

	reqUrl := Host + "/4/find"
	getSlotReq := getSlotRequest{
		Lat:       0,
		Long:      0,
		VenueId:   venueId,
		Day:       day.Format("2006-01-02"),
		PartySize: partySize,
	}
	reqBody, err := json.Marshal(getSlotReq)
	if err != nil {
		return nil, nil, err
	}

	req, err := http.NewRequest(http.MethodPost, reqUrl, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, nil, err
	}

	type getSlotResponse struct {
		Results struct {
			Venues []struct {
				Venue Venue  `json:"venue"`
				Slots []Slot `json:"slots"`
			} `json:"venues"`
		} `json:"results"`
	}

	var getSlotRes getSlotResponse
	err = c.Do(req, &getSlotRes)
	if err != nil {
		return nil, nil, err
	}

	if len(getSlotRes.Results.Venues) == 0 {
		return nil, nil, ErrNoVenues
	}
	return getSlotRes.Results.Venues[0].Slots, &getSlotRes.Results.Venues[0].Venue, nil
}
