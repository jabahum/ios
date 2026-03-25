package db

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"
)

type DBConfig struct {
	Driver          string
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	WaitTimeout     time.Duration
}

func InitDB(ctx context.Context, slog *slog.Logger, cfg DBConfig) (*sql.DB, error) {

	db, err := sql.Open(cfg.Driver, cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open db connection: %w", err)
	}

	// Pool tuning
	if cfg.MaxOpenConns > 0 {
		db.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		db.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}

	// Wait for DB readiness
	waitCtx := ctx
	if cfg.WaitTimeout > 0 {
		var cancel context.CancelFunc
		waitCtx, cancel = context.WithTimeout(ctx, cfg.WaitTimeout)
		defer cancel()
	}

	if err := WaitForDB(waitCtx, db); err != nil {
		return nil, fmt.Errorf("database not ready: %w", err)
	}

	slog.Info(
		"Database initialized successfully")

	return db, nil
}

func PingDB(ctx context.Context, db *sql.DB) error {
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return db.PingContext(pingCtx)
}

func WaitForDB(ctx context.Context, db *sql.DB) error {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timed out waiting for database: %w", ctx.Err())

		case <-ticker.C:
			if err := PingDB(ctx, db); err == nil {
				slog.Info("Database is ready")
				return nil
			} else {
				slog.Info("Waiting for database...")
			}
		}
	}
}
