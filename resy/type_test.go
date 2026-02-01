package resy

import (
	"encoding/json"
	"testing"
)

func TestResyDatetime_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		json      string
		wantTime  string
		wantError bool
	}{
		{
			name:      "valid datetime",
			json:      `"2024-01-15 19:00:00"`,
			wantTime:  "2024-01-15 19:00:00",
			wantError: false,
		},
		{
			name:      "empty string",
			json:      `""`,
			wantTime:  "",
			wantError: false,
		},
		{
			name:      "null value",
			json:      `null`,
			wantTime:  "",
			wantError: false,
		},
		{
			name:      "invalid format ISO8601",
			json:      `"2024-01-15T19:00:00"`,
			wantTime:  "",
			wantError: true,
		},
		{
			name:      "malformed datetime",
			json:      `"invalid-datetime"`,
			wantTime:  "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var dt ResyDatetime
			err := json.Unmarshal([]byte(tt.json), &dt)

			if tt.wantError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			requireNoError(t, err, "UnmarshalJSON failed")

			if tt.wantTime == "" {
				// Check for zero value
				if !dt.Time.IsZero() {
					t.Errorf("expected zero time, got %v", dt.Time)
				}
			} else {
				got := dt.Format(ResyDatetimeFormat)
				if got != tt.wantTime {
					t.Errorf("got %q, want %q", got, tt.wantTime)
				}
			}
		})
	}
}

func TestResyDate_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		json      string
		wantDate  string
		wantError bool
	}{
		{
			name:      "valid date",
			json:      `"2024-01-15"`,
			wantDate:  "2024-01-15",
			wantError: false,
		},
		{
			name:      "empty string",
			json:      `""`,
			wantDate:  "",
			wantError: false,
		},
		{
			name:      "null value",
			json:      `null`,
			wantDate:  "",
			wantError: false,
		},
		{
			name:      "invalid format US style",
			json:      `"01/15/2024"`,
			wantDate:  "",
			wantError: true,
		},
		{
			name:      "date with time included",
			json:      `"2024-01-15 19:00"`,
			wantDate:  "",
			wantError: true,
		},
		{
			name:      "malformed date",
			json:      `"invalid-date"`,
			wantDate:  "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d ResyDate
			err := json.Unmarshal([]byte(tt.json), &d)

			if tt.wantError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			requireNoError(t, err, "UnmarshalJSON failed")

			if tt.wantDate == "" {
				// Check for zero value
				if !d.Time.IsZero() {
					t.Errorf("expected zero time, got %v", d.Time)
				}
			} else {
				got := d.Format(ResyDateFormat)
				if got != tt.wantDate {
					t.Errorf("got %q, want %q", got, tt.wantDate)
				}
			}
		})
	}
}

func TestResyTime_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		json      string
		wantTime  string
		wantError bool
	}{
		{
			name:      "full time format",
			json:      `"19:00:00"`,
			wantTime:  "19:00:00",
			wantError: false,
		},
		{
			name:      "midnight",
			json:      `"00:00:00"`,
			wantTime:  "00:00:00",
			wantError: false,
		},
		{
			name:      "empty string",
			json:      `""`,
			wantTime:  "",
			wantError: false,
		},
		{
			name:      "null value",
			json:      `null`,
			wantTime:  "",
			wantError: false,
		},
		{
			name:      "invalid hour",
			json:      `"25:00:00"`,
			wantTime:  "",
			wantError: true,
		},
		{
			name:      "invalid format",
			json:      `"19:00"`,
			wantTime:  "",
			wantError: true,
		},
		{
			name:      "malformed time",
			json:      `"invalid-time"`,
			wantTime:  "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rt ResyTime
			err := json.Unmarshal([]byte(tt.json), &rt)

			if tt.wantError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			requireNoError(t, err, "UnmarshalJSON failed")

			if tt.wantTime == "" {
				// Check for zero value
				if !rt.Time.IsZero() {
					t.Errorf("expected zero time, got %v", rt.Time)
				}
			} else {
				got := rt.Format(ResyTimeFormat)
				if got != tt.wantTime {
					t.Errorf("got %q, want %q", got, tt.wantTime)
				}
			}
		})
	}
}

func TestTimezone_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		json      string
		wantTZ    string
		wantError bool
	}{
		{
			name:      "America/New_York",
			json:      `"America/New_York"`,
			wantTZ:    "America/New_York",
			wantError: false,
		},
		{
			name:      "UTC",
			json:      `"UTC"`,
			wantTZ:    "UTC",
			wantError: false,
		},
		{
			name:      "America/Los_Angeles",
			json:      `"America/Los_Angeles"`,
			wantTZ:    "America/Los_Angeles",
			wantError: false,
		},
		{
			name:      "empty string",
			json:      `""`,
			wantTZ:    "",
			wantError: false,
		},
		{
			name:      "null value",
			json:      `null`,
			wantTZ:    "",
			wantError: false,
		},
		{
			name:      "invalid timezone",
			json:      `"Invalid/Timezone"`,
			wantTZ:    "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tz Timezone
			err := json.Unmarshal([]byte(tt.json), &tz)

			if tt.wantError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			requireNoError(t, err, "UnmarshalJSON failed")

			if tt.wantTZ == "" {
				// Check for nil location
				if tz.Location != nil {
					t.Errorf("expected nil location, got %v", tz.Location)
				}
			} else {
				if tz.Location == nil {
					t.Fatal("expected non-nil location, got nil")
				}
				got := tz.Location.String()
				if got != tt.wantTZ {
					t.Errorf("got %q, want %q", got, tt.wantTZ)
				}
			}
		})
	}
}

func TestRating_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		json      string
		wantScore float32
		wantCount int
		wantError bool
	}{
		{
			name:      "float format",
			json:      `4.5`,
			wantScore: 4.5,
			wantCount: 0,
			wantError: false,
		},
		{
			name:      "object format",
			json:      `{"average": 4.5, "count": 100}`,
			wantScore: 4.5,
			wantCount: 100,
			wantError: false,
		},
		{
			name:      "zero value",
			json:      `0`,
			wantScore: 0,
			wantCount: 0,
			wantError: false,
		},
		{
			name:      "high rating",
			json:      `5.0`,
			wantScore: 5.0,
			wantCount: 0,
			wantError: false,
		},
		{
			name:      "object with zero count",
			json:      `{"average": 3.8, "count": 0}`,
			wantScore: 3.8,
			wantCount: 0,
			wantError: false,
		},
		{
			name:      "invalid string",
			json:      `"invalid"`,
			wantScore: 0,
			wantCount: 0,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var r Rating
			err := json.Unmarshal([]byte(tt.json), &r)

			if tt.wantError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			requireNoError(t, err, "UnmarshalJSON failed")

			if r.Score != tt.wantScore {
				t.Errorf("Score: got %v, want %v", r.Score, tt.wantScore)
			}
			if r.Count != tt.wantCount {
				t.Errorf("Count: got %v, want %v", r.Count, tt.wantCount)
			}
		})
	}
}

