package resy

import (
	"testing"
)

func TestGeoip_Get(t *testing.T) {
	client, _ := newUnauthenticatedClient(t)

	geoip, err := client.GetGeoip()
	requireNoError(t, err, "GetGeoip failed")
	assertNotNil(t, geoip, "geoip should not be nil")

	// Validate response structure
	if geoip.Success != true {
		t.Errorf("Success: got %v, want true", geoip.Success)
	}

	if geoip.Ip == "" {
		t.Error("IP address should not be empty")
	}

	// Validate latitude and longitude ranges
	if geoip.Latitude < -90 || geoip.Latitude > 90 {
		t.Errorf("Latitude out of range: %v (expected -90 to 90)", geoip.Latitude)
	}

	if geoip.Longitude < -180 || geoip.Longitude > 180 {
		t.Errorf("Longitude out of range: %v (expected -180 to 180)", geoip.Longitude)
	}

	// Country code should be 2-3 characters
	if len(geoip.CountryCode) < 2 || len(geoip.CountryCode) > 3 {
		t.Errorf("CountryCode length unexpected: %q (expected 2-3 chars)", geoip.CountryCode)
	}

	if geoip.Source == "" {
		t.Error("Source should not be empty")
	}

	t.Logf("GeoIP response: IP=%s, Country=%s, Lat=%f, Lon=%f",
		geoip.Ip, geoip.CountryCode, geoip.Latitude, geoip.Longitude)
}
