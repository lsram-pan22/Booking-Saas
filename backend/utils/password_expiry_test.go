package utils

import (
	"testing"
	"time"
)

func TestTemporaryPasswordExpiryBoundary(t *testing.T) {
	now := time.Now()
	past := now.Add(-time.Second)
	future := now.Add(time.Second)
	for _, tc := range []struct {
		name    string
		must    bool
		expires *time.Time
		want    bool
	}{
		{"regular", false, nil, true}, {"temporary valid", true, &future, true}, {"expired", true, &past, false}, {"exact expiry", true, &now, false}, {"missing deadline", true, nil, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if TemporaryPasswordValid(tc.must, tc.expires, now) != tc.want {
				t.Fatal("unexpected password expiry decision")
			}
		})
	}
}
