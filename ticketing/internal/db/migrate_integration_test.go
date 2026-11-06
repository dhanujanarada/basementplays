//go:build integration

package db_test

import (
	"context"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dhanujanarada/basementplays/ticketing/internal/db"
)

func TestMigrateIsIdempotent(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set; skipping Postgres integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	if err := db.Migrate(ctx, dsn, logger); err != nil {
		t.Fatalf("first migrate: %v", err)
	}
	if err := db.Migrate(ctx, dsn, logger); err != nil {
		t.Fatalf("second migrate: %v", err)
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	var schemaExists bool
	if err := pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.schemata WHERE schema_name = 'ticketing'
		)`).Scan(&schemaExists); err != nil {
		t.Fatalf("query schema: %v", err)
	}
	if !schemaExists {
		t.Fatal("ticketing schema was not created")
	}

	var versionTableExists bool
	if err := pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_schema = 'ticketing' AND table_name = 'goose_db_version'
		)`).Scan(&versionTableExists); err != nil {
		t.Fatalf("query version table: %v", err)
	}
	if !versionTableExists {
		t.Fatal("ticketing.goose_db_version was not created")
	}

	for _, table := range []string{
		"seat_plans",
		"shows",
		"show_seats",
		"orders",
		"order_items",
		"tickets",
		"payments",
		"audit_log",
	} {
		var exists bool
		if err := pool.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM information_schema.tables
				WHERE table_schema = 'ticketing' AND table_name = $1
			)`, table).Scan(&exists); err != nil {
			t.Fatalf("query table %s: %v", table, err)
		}
		if !exists {
			t.Errorf("ticketing.%s was not created", table)
		}
	}
}
