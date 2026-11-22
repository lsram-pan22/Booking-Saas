package utils

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"testing"
)

type subscriptionRow struct{ err error }

func (r subscriptionRow) Scan(dest ...any) error {
	if r.err == nil {
		*(dest[0].(*string)) = "subscription-id"
	}
	return r.err
}

type subscriptionQuery struct{ err error }

func (q subscriptionQuery) QueryRow(context.Context, string, ...any) pgx.Row {
	return subscriptionRow{err: q.err}
}
func TestSubscriptionLookupFailsClosed(t *testing.T) {
	failure := errors.New("database unavailable")
	for _, tc := range []struct {
		name   string
		err    error
		active bool
	}{
		{"active subscription", nil, true}, {"no owner subscription", pgx.ErrNoRows, false}, {"database failure", failure, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			active, err := ShopSubscriptionActive(context.Background(), subscriptionQuery{tc.err}, "shop")
			if active != tc.active {
				t.Fatalf("active=%v", active)
			}
			if tc.err == failure && !errors.Is(err, failure) {
				t.Fatal("database failure was hidden")
			}
			if tc.err != failure && err != nil {
				t.Fatal(err)
			}
		})
	}
}
