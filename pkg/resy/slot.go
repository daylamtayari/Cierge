package resy

type Slot struct {
	Config SlotConfig
	Date   SlotDate
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
type SlotPayment struct {
	IsPaid           bool    `json:"is_paid"`
	DepositFee       float32 `json:"deposit_fee"`
	ServiceCharge    string  `json:"service_charge"`
	VenueShare       int     `json:"venue_share"`
	PaymentStructure int     `json:"payment_structure"`
	SecsCancelCutOff int     `json:"secs_cancel_cut_off"`
	TimeCancelCutOff string  `json:"time_cancel_cut_off"`
	SecsChangeCutOff int     `json:"secs_change_cut_off"`
	TimeChangeCutOff string  `json:"time_change_cut_off"`
}
