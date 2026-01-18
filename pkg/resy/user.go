package resy

// Represents a Resy user
// NOTE: Not all fields are populated or included
// depending on the API endpoint that it is returned from
type User struct {
	Id                      int             `json:"id"`
	FirstName               string          `json:"first_name"`
	LastName                string          `json:"last_name"`
	Bio                     *string         `json:"bio"`
	MobileNumber            string          `json:"mobile_number"`
	EmailAddress            string          `json:"email_address"`
	EmailVerified           int             `json:"em_is_verified"`
	MobileNumberVerified    int             `json:"mobile_number_is_verified"`
	IsActive                int             `json:"is_active"`
	PaymentMethods          []PaymentMethod `json:"payment_methods"`
	PaymentMethodId         int             `json:"payment_method_id"`
	ReferralCode            string          `json:"referral_code"`
	IsMarketable            int             `json:"is_marketable"`
	IsConcierge             int             `json:"is_concierge"`
	DateUpdated             int             `json:"date_updated"`
	DateCreated             int             `json:"date_created"`
	HasSetPassword          int             `json:"has_set_password"`
	BookingsNumber          int             `json:"num_bookings"`
	FeatureFlags            map[string]any  `json:"feature_flags"`
	ProfileImageUrl         string          `json:"profile_image_url"`
	IsGlobalDiningAccess    bool            `json:"is_global_dining_access"`
	IsRGA                   bool            `json:"is_rga"`
	IsPlatinumNightEligible bool            `json:"is_platinum_night_eligible"`
	GuestId                 int             `json:"guest_id"`
	ResyMemberSince         int             `json:"resy_member_since"`
}

// Represents a user's payment method
// If it is a credit card, the Display
// field will contain the last 4 digits
// of the credit card
type PaymentMethod struct {
	Id              int    `json:"id"`
	IsDefault       bool   `json:"is_default"`
	ProviderId      int    `json:"provider_id"`
	ProviderName    string `json:"provider_name"`
	Display         string `json:"display"`
	Type            string `json:"type"`
	ExpirationYear  int    `json:"exp_year"`
	ExpirationMonth int    `json:"exp_month"`
}
