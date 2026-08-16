package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Open(ctx context.Context, url string) (*pgxpool.Pool, error) {
	if url == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	var pool *pgxpool.Pool
	var err error
	for attempt := 1; attempt <= 30; attempt++ {
		pool, err = pgxpool.New(ctx, url)
		if err == nil {
			err = pool.Ping(ctx)
			if err == nil {
				err = ensureStaffSchema(ctx, pool)
				if err == nil {
					return pool, nil
				}
			}
			pool.Close()
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
	return nil, fmt.Errorf("connect postgres: %w", err)
}

func ensureStaffSchema(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS restaurants (id BIGSERIAL PRIMARY KEY, name TEXT NOT NULL, branch TEXT NOT NULL DEFAULT 'Main Branch', address TEXT NOT NULL DEFAULT '', phone TEXT NOT NULL DEFAULT '', created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()); ALTER TABLE restaurants ADD COLUMN IF NOT EXISTS branch TEXT NOT NULL DEFAULT 'Main Branch'; DROP INDEX IF EXISTS one_restaurant_account; ALTER TABLE users ADD COLUMN IF NOT EXISTS phone TEXT NOT NULL DEFAULT ''; ALTER TABLE users ADD COLUMN IF NOT EXISTS aadhar TEXT NOT NULL DEFAULT ''; ALTER TABLE users ADD COLUMN IF NOT EXISTS address TEXT NOT NULL DEFAULT ''; ALTER TABLE users ADD COLUMN IF NOT EXISTS image_url TEXT NOT NULL DEFAULT ''; ALTER TABLE users ADD COLUMN IF NOT EXISTS branch TEXT NOT NULL DEFAULT 'Main Branch'; ALTER TABLE users ADD COLUMN IF NOT EXISTS parent_id BIGINT REFERENCES users(id) ON DELETE SET NULL; ALTER TABLE users ADD COLUMN IF NOT EXISTS staff_type TEXT NOT NULL DEFAULT ''; ALTER TABLE users ADD COLUMN IF NOT EXISTS active BOOLEAN NOT NULL DEFAULT TRUE; ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check; ALTER TABLE users ADD CONSTRAINT users_role_check CHECK (role IN ('customer','order_manager','stock_manager','staff_manager','admin','branch_manager','special_admin','hr_vp','hr_manager','other','owner')); CREATE UNIQUE INDEX IF NOT EXISTS one_owner_account ON users ((role)) WHERE role='owner'; CREATE UNIQUE INDEX IF NOT EXISTS one_hr_vp_account ON users ((role)) WHERE role='hr_vp'; CREATE TABLE IF NOT EXISTS salary_payments (id BIGSERIAL PRIMARY KEY, user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE, month TEXT NOT NULL, amount NUMERIC(12,2) NOT NULL CHECK (amount >= 0), paid_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), UNIQUE(user_id, month));`)
	return err
}
