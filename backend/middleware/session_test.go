package middleware

import (
	"testing"
	"time"
)

func TestSessionPolicy(t *testing.T) {
	future := time.Now().Add(time.Hour)
	for _, tc := range []struct {
		name    string
		state   SessionState
		version int64
		role    string
		allow   bool
		want    int
	}{
		{"current", SessionState{Version: 2, Role: "user", Status: "active"}, 2, "user", false, 0},
		{"revoked", SessionState{Version: 3, Role: "user", Status: "active"}, 2, "user", true, 401},
		{"suspended", SessionState{Role: "user", Status: "suspended"}, 0, "user", true, 401},
		{"role changed", SessionState{Role: "user", Status: "active"}, 0, "super_admin", true, 401},
		{"temporary password cannot book", SessionState{Role: "user", Status: "active", MustChange: true, ResetExpiresAt: &future}, 0, "user", false, 403},
		{"temporary password can change", SessionState{Role: "user", Status: "active", MustChange: true, ResetExpiresAt: &future}, 0, "user", true, 0},
		{"temporary password without expiry denied", SessionState{Role: "user", Status: "active", MustChange: true}, 0, "user", true, 401},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := SessionPolicy(tc.state, tc.version, tc.role, tc.allow); got != tc.want {
				t.Fatalf("got %d want %d", got, tc.want)
			}
		})
	}
}
