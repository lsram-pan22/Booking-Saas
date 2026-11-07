package utils

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"time"

	"booking-saas/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func ConnectDatabase(cfg *config.Config) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}

	pg := cfg.PostgreSQL

	dsn := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(pg.User, pg.Pass),
		Host:   net.JoinHostPort(pg.Host, pg.Port),
		Path:   "/" + pg.Name,
	}

	query := url.Values{}
	query.Set("sslmode", pg.SSLMode)
	dsn.RawQuery = query.Encode()

	poolConfig, err := pgxpool.ParseConfig(dsn.String())
	if err != nil {
		return fmt.Errorf("failed to parse database config: %w", err)
	}

	poolConfig.MaxConns = 10
	poolConfig.MinConns = 2
	poolConfig.MaxConnLifetime = time.Hour

	pool, err := pgxpool.NewWithConfig(
		context.Background(),
		poolConfig,
	)
	if err != nil {
		return fmt.Errorf("failed to create database pool: %w", err)
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return fmt.Errorf("failed to connect database: %w", err)
	}

	DB = pool

	return nil
}

func CloseDatabase() {
	if DB != nil {
		DB.Close()
		DB = nil
	}
}
