package database

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(databaseURL string) *pgxpool.Pool {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		log.Fatalf("❌ Failed to parse database URL: %v", err)
	}

	config.MaxConns = 10
	config.MinConns = 2

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Fatalf("❌ Failed to create connection pool: %v", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("❌ Failed to ping database: %v", err)
	}

	fmt.Println("✅ Database connected successfully")
	return pool
}