package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Open(ctx context.Context, url string) (*pgxpool.Pool, error) {
	if url == "" { return nil, fmt.Errorf("DATABASE_URL is required") }
	var pool *pgxpool.Pool
	var err error
	for attempt := 1; attempt <= 30; attempt++ {
		pool, err = pgxpool.New(ctx, url)
		if err == nil { err = pool.Ping(ctx); if err == nil { return pool, nil }; pool.Close() }
		select { case <-ctx.Done(): return nil, ctx.Err(); case <-time.After(2 * time.Second): }
	}
	return nil, fmt.Errorf("connect postgres: %w", err)
}
