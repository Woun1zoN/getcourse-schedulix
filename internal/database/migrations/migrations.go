package migrations

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Run(ctx context.Context, dbURL, migrationsPath string) error {
	if err := WaitForDB(ctx, dbURL, 30*time.Second); err != nil {
		return fmt.Errorf("database not ready: %w", err)
	}

	m, err := migrate.New(migrationsPath, dbURL)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("run migrations: %w", err)
	}

	return nil
}

func WaitForDB(ctx context.Context, dbURL string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return fmt.Errorf("create pool: %w", err)
	}
	defer pool.Close()

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	var lastErr error

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("waiting for db: %w (last ping error: %v)", ctx.Err(), lastErr)
		case <-ticker.C:
			pingCtx, cancelPing := context.WithTimeout(ctx, time.Second)
			lastErr = pool.Ping(pingCtx)
			cancelPing()

			if lastErr == nil {
				return nil
			}
		}
	}
}
