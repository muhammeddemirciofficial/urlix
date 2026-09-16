package repository

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/muhammeddemirciofficial/urlix/internal/service"
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

	_, _ = db.Exec("DELETE FROM urls WHERE code = $1", code)

	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM urls WHERE code = $1", code)
	})

	err = repo.Save(code, url, nil)
	if err != nil {
		t.Fatalf("failed to save URL: %v", err)
	}

	got, err := repo.Get(code)
	if err != nil {
		t.Fatalf("failed to get URL: %v", err)
	}

	if got.URL != url {
		t.Errorf("expected URL %s, got %s", url, got.URL)
	}
}

func TestPostgresURLRepository_SaveDuplicate(t *testing.T) {
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

	code := "dup01"
	url := "https://example.com"

	_, _ = db.Exec("DELETE FROM urls WHERE code = $1", code)

	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM urls WHERE code = $1", code)
	})

	if err := repo.Save(code, url, nil); err != nil {
		t.Fatalf("failed to save first URL: %v", err)
	}

	err = repo.Save(code, url, nil)
	if !errors.Is(err, service.ErrDuplicateCode) {
		t.Fatalf("expected ErrDuplicateCode, got %v", err)
	}
}

func TestPostgresURLRepository_RecordClick(t *testing.T) {
	dsn := os.Getenv("URLIX_TEST_DATABASE_URL")

	if dsn == "" {
		dsn = "postgres://urlix:urlix@localhost:5432/urlix?sslmode=disable"
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	repo := NewPostgresURLRepository(db)

	var urlID int64

	code := "analytics-test"

	err = db.QueryRow(`
		INSERT INTO urls (code, url)
		VALUES ($1, 'https://example.com')
		RETURNING id
	`, code).Scan(&urlID)

	if err != nil {
		t.Fatalf("failed to create URL: %v", err)
	}

	defer func() {
		_, _ = db.Exec(`DELETE FROM urls WHERE code = $1`, code)
	}()

	err = repo.RecordClick(
		urlID,
		"Mozilla/5.0",
		"https://google.com",
	)

	if err != nil {
		t.Fatalf("failed to record click: %v", err)
	}

	var (
		gotUserAgent string
		gotReferer   string
	)

	err = db.QueryRow(`
		SELECT user_agent, referer
		FROM url_clicks
		WHERE url_id = $1
	`, urlID).Scan(&gotUserAgent, &gotReferer)

	if err != nil {
		t.Fatalf("failed to query click: %v", err)
	}

	if gotUserAgent != "Mozilla/5.0" {
		t.Errorf("expected user agent Mozilla/5.0, got %s", gotUserAgent)
	}

	if gotReferer != "https://google.com" {
		t.Errorf("expected referer https://google.com, got %s", gotReferer)
	}
}
