package repository

import (
	"context"
	"database/sql"

	"github.com/muhammeddemirciofficial/urlix/internal/service"
)

type PostgresURLRepository struct {
	db *sql.DB
}

func NewPostgresURLRepository(db *sql.DB) *PostgresURLRepository {
	return &PostgresURLRepository{
		db: db,
	}
}

func (r *PostgresURLRepository) Save(code string, url string) error {
	_, err := r.db.ExecContext(
		context.Background(),
		`INSERT INTO urls (code, url) VALUES ($1, $2)`,
		code,
		url,
	)
	return err
}

func (r *PostgresURLRepository) Get(code string) (string, error) {
	var url string

	err := r.db.QueryRowContext(
		context.Background(),
		`SELECT url FROM urls WHERE code = $1`,
		code,
	).Scan(&url)

	if err != nil {
		return "", err
	}

	return url, nil
}

func (r *PostgresURLRepository) IncrementClickCount(code string) error {
	_, err := r.db.ExecContext(
		context.Background(),
		`UPDATE urls
		SET click_count = click_count + 1
		WHERE code = $1`,
		code,
	)

	return err
}

func (r *PostgresURLRepository) GetStats(code string) (service.URLStats, error) {
	var stats service.URLStats

	err := r.db.QueryRowContext(
		context.Background(),
		`SELECT code, url, click_count FROM urls WHERE code = $1`,
		code,
	).Scan(&stats.Code, &stats.URL, &stats.ClickCount)

	if err != nil {
		return service.URLStats{}, err
	}

	return stats, nil
}
