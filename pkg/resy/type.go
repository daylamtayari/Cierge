package resy

import (
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
