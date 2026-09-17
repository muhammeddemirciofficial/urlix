package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		panic("DATABASE_URL is required")
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := db.PingContext(ctx); err != nil {
		panic(err)
	}

	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version BIGINT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`); err != nil {
		panic(err)
	}

	entries, err := os.ReadDir("/app/migrations")
	if err != nil {
		panic(err)
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			files = append(files, entry.Name())
		}
	}
	sort.Strings(files)

	for _, name := range files {
		version, err := migrationVersion(name)
		if err != nil {
			panic(err)
		}

		var applied bool
		if err := db.QueryRowContext(ctx,
			`SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)`, version,
		).Scan(&applied); err != nil {
			panic(err)
		}
		if applied {
			continue
		}

		content, err := os.ReadFile(filepath.Join("/app/migrations", name))
		if err != nil {
			panic(err)
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			panic(err)
		}

		if _, err := tx.ExecContext(ctx, string(content)); err != nil {
			_ = tx.Rollback()
			panic(fmt.Errorf("migration %s: %w", name, err))
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO schema_migrations(version) VALUES ($1)`, version,
		); err != nil {
			_ = tx.Rollback()
			panic(err)
		}
		if err := tx.Commit(); err != nil {
			panic(err)
		}

		fmt.Printf("applied migration %s\n", name)
	}
}

func migrationVersion(name string) (int64, error) {
	parts := strings.SplitN(name, "_", 2)
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid migration filename: %s", name)
	}
	return strconv.ParseInt(parts[0], 10, 64)
}
