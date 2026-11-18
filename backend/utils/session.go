package utils

import (
	"context"
	"github.com/jackc/pgx/v5/pgconn"
)

type SessionExecutor interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
}

func RevokeAccountSessions(ctx context.Context, db SessionExecutor, claims *JWTClaims) (bool, error) {
	if claims == nil || !ValidUserID(claims.UserID) || claims.SessionVersion < 0 {
		return false, nil
	}
	result, err := db.Exec(ctx, `UPDATE users SET session_version=session_version+1 WHERE id=$1 AND session_version=$2 AND status='active' AND system_role=$3`, claims.UserID, claims.SessionVersion, claims.Role)
	return err == nil && result.RowsAffected() == 1, err
}
