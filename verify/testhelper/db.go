package testhelper

import (
	"context"
	"fmt"
	"testing"

	"github.com/uptrace/bun"

	"github.com/he-end/verify-reverse/verify/repository"
)

func NewTestDB(t *testing.T) *bun.DB {
	t.Helper()

	dsn := "postgres://postgres:postgres@localhost:5433/verify_auth?sslmode=disable"

	db, err := repository.NewPostgresDB(t.Context(), dsn)
	if err != nil {
		t.Fatalf("failed to connect to postgres: %v", err)
	}

	if err := repository.RunMigrations(context.Background(), db); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

func TruncateAll(ctx context.Context, db *bun.DB) error {
	tables := []string{
		"verification_attempts",
		"verification_codes",
		"sessions",
		"users",
	}
	for _, table := range tables {
		if _, err := db.ExecContext(ctx, fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)); err != nil {
			return fmt.Errorf("failed to truncate %s: %w", table, err)
		}
	}
	return nil
}
