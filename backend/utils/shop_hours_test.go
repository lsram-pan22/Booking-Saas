package utils

import (
	"testing"
	"time"
)

func TestShopOpeningHours(t *testing.T) {
	for _, tc := range []struct {
		open, close string
		valid       bool
	}{
		{"09:00", "18:00", true}, {"00:00", "24:00", true}, {"09:15", "18:45", true}, {"9:00", "18:00", false}, {"18:00", "09:00", false}, {"09:00", "09:00", false}, {"09:60", "18:00", false}, {"24:00", "24:00", false},
	} {
		if ValidShopHours(tc.open, tc.close) != tc.valid {
			t.Errorf("%s–%s validation", tc.open, tc.close)
		}
	}
	parse := func(value string) time.Time {
		t.Helper()
		v, e := time.Parse(time.RFC3339, value)
		if e != nil {
			t.Fatal(e)
		}
		return v
	}
	for _, tc := range []struct {
		start, end, open, close string
		valid                   bool
	}{
		{"2030-01-02T02:00:00Z", "2030-01-02T11:00:00Z", "09:00", "18:00", true},
		{"2030-01-02T08:59:00+07:00", "2030-01-02T09:30:00+07:00", "09:00", "18:00", false},
		{"2030-01-02T17:30:00+07:00", "2030-01-02T18:01:00+07:00", "09:00", "18:00", false},
		{"2030-01-02T23:30:00+07:00", "2030-01-03T00:00:00+07:00", "00:00", "24:00", true},
		{"2030-01-02T23:30:00+07:00", "2030-01-03T00:01:00+07:00", "00:00", "24:00", false},
	} {
		if WithinShopHours(parse(tc.start), parse(tc.end), tc.open, tc.close) != tc.valid {
			t.Errorf("interval %s–%s", tc.start, tc.end)
		}
	}
}
