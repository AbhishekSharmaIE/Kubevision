package db

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// MustConnectPostgres creates a connection pool from DATABASE_URL or exits.
func MustConnectPostgres() *pgxpool.Pool {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		url = "postgres://kubevision:kubevision@localhost:5432/kubevision?sslmode=disable"
	}
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		panic(fmt.Sprintf("postgres config: %v", err))
	}
	cfg.MaxConns = 20
	cfg.MinConns = 2
	cfg.MaxConnLifetime = time.Hour
	cfg.HealthCheckPeriod = time.Minute

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		panic(fmt.Sprintf("postgres connect: %v", err))
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		panic(fmt.Sprintf("postgres ping: %v", err))
	}
	return pool
}
