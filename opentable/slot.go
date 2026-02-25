package opentable

type Slot struct {
	IsAvailable                  bool   `json:"isAvailable"`
	TimeOffsetMinutes            int    `json:"timeOffsetMinutes"`
	SlotHash                     string `json:"slotHash"`
	PointsType                   string `json:"pointsType"`
	PointsValue                  int    `json:"pointsValue"`
	HasPrivateDiningAvailability bool   `json:"hasPrivateDiningAvailability"`
	AvailableSpaceIds            []int  `json:"AvailableSpaceIds"`
	ExperienceIds                []int  `json:"experienceIds"`
	AvailabilityToken            string `json:"slotAvailabilityToken"`
	RedemptionTier               string `json:"redemptionTier"`
	Type                         string `json:"type"`
}
