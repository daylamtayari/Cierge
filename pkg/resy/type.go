package resy

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// ResyDatetime wraps time.Time to
// handle Resy's datetime format
type ResyDatetime struct {
	time.Time
}

const ResyDatetimeFormat = "2006-01-02 15:04:05"

// Custom unmarshaller for the ResyDatetime type
func (t *ResyDatetime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	if s == "null" || s == "" {
		return nil
	}

	parsedTime, err := time.Parse(ResyDatetimeFormat, s)
	if err != nil {
		return err
	}

	t.Time = parsedTime
	return nil
}

// Custom marshaller for the ResyDatetime type
func (t ResyDatetime) MarshalJSON() ([]byte, error) {
	if t.IsZero() {
		return []byte("null"), nil
	}
	return fmt.Appendf(nil, "\"%s\"", t.Format(ResyDatetimeFormat)), nil
}

// ResyDate wraps time.Time to
// handle Resy's date format
type ResyDate struct {
	time.Time
}

const ResyDateFormat = "2006-01-02"

// Custom unmarshaller for the ResyDate type
func (t *ResyDate) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	if s == "null" || s == "" {
		return nil
	}

	parsedTime, err := time.Parse(ResyDateFormat, s)
	if err != nil {
		return err
	}

	t.Time = parsedTime
	return nil
}

// Custom marshaller for the ResyDatetime type
func (t ResyDate) MarshalJSON() ([]byte, error) {
	if t.IsZero() {
		return []byte("null"), nil
	}
	return fmt.Appendf(nil, "\"%s\"", t.Format(ResyDateFormat)), nil
}

// Rating represents a restaurant's rating
// Need to create a custom type as Resy is annoying and depending on the
// API endpoint, the `rating` field in the `venue` object can be either
// a float value representing the rating or an object containing the
// rating and the amount of reviews (which can also be represented
// by the `total_ratings` field in a `venue` object...)
type Rating struct {
	Score float32 `json:"score"`
	Count int     `json:"count"`
}

// UnmarshalJSON handles both rating formats
func (r *Rating) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as a float first
	var value float32
	if err := json.Unmarshal(data, &value); err == nil {
		r.Score = value
		return nil
	}

	// Otherwise, try the object
	var detail struct {
		Average float32 `json:"average"`
		Count   int     `json:"count"`
	}
	if err := json.Unmarshal(data, &detail); err != nil {
		return err
	}

	r.Score = detail.Average
	r.Count = detail.Count
	return nil
}
