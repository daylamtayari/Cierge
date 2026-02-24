package opentable

type Restaurant struct {
	Id             int                  `json:"restaurantId"`
	Name           string               `json:"name"`
	Description    string               `json:"description"`
	DiningStyle    string               `json:"diningStyle"`
	MaxAdvanceDays *int                 `json:"maxAdvanceDays"`
	Statistics     RestaurantStatistics `json:"statistics"`
	Neighborhood   struct {
		Id   *int   `json:"neighborhoodId"`
		Name string `json:"name"`
	} `json:"neighborhood"`
	Coordinates struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	} `json:"coordinates"`
	Address struct {
		Line1    string  `json:"line1"`
		Line2    string  `json:"line2"`
		City     string  `json:"city"`
		State    string  `json:"state"`
		PostCode string  `json:"postCode"`
		Country  *string `json:"country"`
	} `json:"address"`
	Website *string `json:"website"`
	Urls    struct {
		ProfileLink PageLink `json:"profileLink"`
	} `json:"urls"`
	ContactInformation struct {
		PhoneNumber          string `json:"phoneNumber"`
		FormattedPhoneNumber string `json:"formattedPhoneNumber"`
	} `json:"contactInformation"`
	Awards    []Award `json:"awards"`
	PriceBand struct {
		Id             int    `json:"priceBandId"`
		CurrencySymbol string `json:"currencySymbol"`
		Name           string `json:"name"`
	} `json:"priceBand"`
	Features struct {
		MaxPartySize int  `json:""`
		Bar          bool `json:"bar"`
		Counter      bool `json:"counter"`
		HighTop      bool `json:"high_top"`
		Outdoor      bool `json:"outdoor"`
	} `json:"features"`
	HasTakeout bool `json:"hasTakeout"`
}

type RestaurantStatistics struct {
	RecentReservationCount int `json:"recent_reservation_count"`
	Reviews                struct {
		AllTimeTextReviews int   `json:"allTimeTextReviewCount"`
		Total              *int  `json:"totalNumberOfReviews"`
		VerifiedReviews    *bool `json:"verifiedReviews"`
		Ratings            struct {
			Count        *int             `json:"count"`
			Overall      RatingStatistic  `json:"overall"`
			Ambience     *RatingStatistic `json:"ambience"`
			Food         *RatingStatistic `json:"food"`
			Service      *RatingStatistic `json:"service"`
			Value        *RatingStatistic `json:"value"`
			Noise        *RatingStatistic `json:"noise"`
			Distribution *struct {
				One   int `json:"one"`
				Two   int `json:"two"`
				Three int `json:"three"`
				Four  int `json:"four"`
				Five  int `json:"five"`
			} `json:"overallRatingsDistribution"`
		} `json:"ratings"`
	} `json:"reviews"`
}

type RatingStatistic struct {
	Rating float32 `json:"rating"`
	Count  *int    `json:"count"`
}

type Award struct {
	Name   string  `json:"name"`
	Year   string  `json:"year"`
	Rating *string `json:"rating"`
}
