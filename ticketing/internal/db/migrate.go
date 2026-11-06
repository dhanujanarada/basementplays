package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/dhanujanarada/basementplays/ticketing/migrations"
)

const (
	// migrationLockKey is a fixed application-wide key. It must never change, or two
	// deploys could migrate concurrently.
	migrationLockKey int64 = 0x5449434b45543031 // "TICKET01"

	migrationLockWait = 30 * time.Second
)

// Migrate applies every pending SQL migration under an advisory lock. The schema is
// created first so goose's version table can live inside `ticketing`. It is safe to
// call concurrently and safe to call repeatedly.
func Migrate(ctx context.Context, dsn string, logger *slog.Logger) error {
	cfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		return fmt.Errorf("parse DATABASE_URL: %w", err)
	}

	sqlDB := stdlib.OpenDB(*cfg)
	defer sqlDB.Close()
	sqlDB.SetMaxOpenConns(4)
	sqlDB.SetMaxIdleConns(2)
	sqlDB.SetConnMaxIdleTime(time.Minute)

	lockCtx, cancel := context.WithTimeout(ctx, migrationLockWait)
	defer cancel()

	conn, err := sqlDB.Conn(lockCtx)
	if err != nil {
		return fmt.Errorf("open migration connection: %w", err)
	}
	defer conn.Close()

	if _, err := conn.ExecContext(lockCtx, "SELECT pg_advisory_lock($1)", migrationLockKey); err != nil {
		return fmt.Errorf("acquire migration lock: %w", err)
	}
	defer func() {
		unlockCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, err := conn.ExecContext(unlockCtx, "SELECT pg_advisory_unlock($1)", migrationLockKey); err != nil {
			logger.Warn("release migration lock", "error", err)
		}
	}()

	if _, err := conn.ExecContext(ctx, "CREATE SCHEMA IF NOT EXISTS ticketing"); err != nil {
		return fmt.Errorf("create ticketing schema: %w", err)
	}

	goose.SetBaseFS(migrations.FS)
	goose.SetTableName("ticketing.goose_db_version")
	goose.SetLogger(gooseLogger{logger: logger})
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set migration dialect: %w", err)
	}

	if err := goose.UpContext(ctx, sqlDB, "."); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}

	return nil
}

// gooseLogger adapts goose's package-level logger to slog.
type gooseLogger struct {
	logger *slog.Logger
}

func (g gooseLogger) Print(v ...any) {
	g.logger.Info(fmt.Sprint(v...))
}

func (g gooseLogger) Printf(format string, v ...any) {
	g.logger.Info(fmt.Sprintf(format, v...))
}

func (g gooseLogger) Fatal(v ...any) {
	g.logger.Error(fmt.Sprint(v...))
}

func (g gooseLogger) Fatalf(format string, v ...any) {
	g.logger.Error(fmt.Sprintf(format, v...))
}
