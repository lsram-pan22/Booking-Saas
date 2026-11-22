package migrations

import (
	"context"
	"crypto/sha256"
	"embed"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed *.sql
var files embed.FS

// Apply runs pending migrations atomically in public. Existing migrations must
// have matching checksums; unmanaged schemas are never silently adopted.
func Apply(ctx context.Context, pool *pgxpool.Pool) ([]string, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())
	if _, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(709130221); SET LOCAL search_path TO public,pg_catalog; SET LOCAL lock_timeout TO '10s'`); err != nil {
		return nil, err
	}
	if _, err = tx.Exec(ctx, `CREATE TABLE IF NOT EXISTS public.schema_migrations(name TEXT PRIMARY KEY,checksum TEXT NOT NULL,applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW())`); err != nil {
		return nil, err
	}
	entries, err := files.ReadDir(".")
	if err != nil {
		return nil, err
	}
	names := []string{}
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".sql") {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	applied := []string{}
	for _, name := range names {
		data, err := files.ReadFile(name)
		if err != nil {
			return nil, err
		}
		checksum := fmt.Sprintf("%x", sha256.Sum256(data))
		var recorded string
		err = tx.QueryRow(ctx, `SELECT checksum FROM public.schema_migrations WHERE name=$1`, name).Scan(&recorded)
		if err == nil {
			if recorded != checksum {
				return nil, fmt.Errorf("migration checksum mismatch: %s", name)
			}
			continue
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
		body, err := transactionBody(string(data))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		if _, err = tx.Exec(ctx, body); err != nil {
			return nil, fmt.Errorf("migration %s: %w", name, err)
		}
		if _, err = tx.Exec(ctx, `INSERT INTO public.schema_migrations(name,checksum) VALUES($1,$2)`, name, checksum); err != nil {
			return nil, err
		}
		applied = append(applied, name)
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return applied, nil
}

func transactionBody(sql string) (string, error) {
	sql = strings.TrimSpace(sql)
	if !strings.HasPrefix(sql, "BEGIN;") || !strings.HasSuffix(sql, "COMMIT;") {
		return "", errors.New("expected BEGIN/COMMIT wrapper")
	}
	return strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(sql, "BEGIN;"), "COMMIT;")), nil
}
