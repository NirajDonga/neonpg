package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func Connect(ctx context.Context, dbUrl string) (*pgxpool.Pool, error) {
	currCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	pool, err := pgxpool.New(currCtx, dbUrl)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err := pool.Ping(currCtx); err != nil {
		return nil, fmt.Errorf("unable to connect to database (ping failed): %w", err)
	}

	DB = pool

	fmt.Println("Connected to Postgres successfully!")
	return pool, nil

}
