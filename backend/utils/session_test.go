package utils

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
	"testing"
)

type revokerStub struct {
	tag  string
	err  error
	args []any
}

func (s *revokerStub) Exec(_ context.Context, _ string, args ...any) (pgconn.CommandTag, error) {
	s.args = args
	return pgconn.NewCommandTag(s.tag), s.err
}
func TestSessionRevocationFailsClosed(t *testing.T) {
	claims := &JWTClaims{UserID: "11111111-1111-4111-8111-111111111111", Role: "user", SessionVersion: 4}
	for _, tc := range []struct {
		name, tag string
		err       error
		want      bool
	}{
		{"current", "UPDATE 1", nil, true}, {"already revoked", "UPDATE 0", nil, false}, {"database failure", "", errors.New("offline"), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &revokerStub{tag: tc.tag, err: tc.err}
			revoked, err := RevokeAccountSessions(context.Background(), repo, claims)
			if revoked != tc.want || (tc.err != nil && err == nil) {
				t.Fatal("revocation result incorrect")
			}
			if repo.args[0] != claims.UserID || repo.args[1] != claims.SessionVersion || repo.args[2] != claims.Role {
				t.Fatal("wrong account/session revoked")
			}
		})
	}
}
