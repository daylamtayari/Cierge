package model

import (
	"database/sql/driver"
	"errors"
	"time"
)

var (
	ErrTimezoneNotString = errors.New("timezone value from database is not a string")
	ErrTimezoneInvalid   = errors.New("timezone value is not a valid timezone")
)

// Wraps time.Location for db compatibility
type Timezone struct {
	*time.Location
}

// Implements the scanner interface for reading from the db
func (t *Timezone) Scan(value any) error {
	if value == nil {
		t.Location = nil
		return nil
	}

	str, ok := value.(string)
	if !ok {
		return ErrTimezoneNotString
	}

	location, err := time.LoadLocation(str)
	if err != nil {
		return ErrTimezoneInvalid
	}

	t.Location = location
	return nil
}

// Implements the value interface for writing to the db
func (t Timezone) Value() (driver.Value, error) {
	if t.Location == nil {
		return nil, nil
	}

	return t.String(), nil
}
