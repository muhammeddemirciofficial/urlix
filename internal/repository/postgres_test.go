package repository

import (
	"context"
	"database/sql"
	"os"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestPostgresURLRepository_SaveAndGet(t *testing.T) {
	dsn := os.Getenv("URLIX_TEST_DATABASE_URL")

	if dsn == "" {
		dsn = "postgres://urlix:urlix@localhost:5432/urlix?sslmode=disable"
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.PingContext(context.Background()); err != nil {
		t.Fatalf("failed to ping database: %v", err)
	}

	repo := NewPostgresURLRepository(db)

	code := "test01"
	url := "https://example.com"

	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM urls WHERE code = $1", code)
	})

	err = repo.Save(code, url)
	if err != nil {
		t.Fatalf("failed to save URL: %v", err)
	}

	got, err := repo.Get(code)
	if err != nil {
		t.Fatalf("failed to get URL: %v", err)
	}

	if got != url {
		t.Errorf("expected URL %s, got %s", url, got)
	}
}
