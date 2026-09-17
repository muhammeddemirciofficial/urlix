package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
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

func (r *PostgresURLRepository) Save(code string, url string, expiresAt *time.Time) error {
	_, err := r.db.ExecContext(
		context.Background(),
		`INSERT INTO urls (code, url, expires_at)
		 VALUES ($1, $2, $3)`,
		code,
		url,
		expiresAt,
	)

	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return service.ErrDuplicateCode
	}

	return err
}

func (r *PostgresURLRepository) Get(code string) (service.URL, error) {
	var result service.URL

	err := r.db.QueryRowContext(
		context.Background(),
		`SELECT id, code, url, expires_at
		 FROM urls
		 WHERE code = $1`,
		code,
	).Scan(
		&result.ID,
		&result.Code,
		&result.URL,
		&result.ExpiresAt,
	)

	if err != nil {
		return service.URL{}, err
	}

	return result, nil
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

func (r *PostgresURLRepository) RecordClick(urlID int64, userAgent string, referer string) error {
	_, err := r.db.ExecContext(
		context.Background(),
		`INSERT INTO url_clicks (url_id, user_agent, referer)
         VALUES ($1, $2, $3)`,
		urlID,
		userAgent,
		referer,
	)

	return err
}

func (r *PostgresURLRepository) GetAnalytics(code string) (service.URLAnalytics, error) {
	var analytics service.URLAnalytics

	err := r.db.QueryRowContext(
		context.Background(),
		`SELECT code, url, click_count
		 FROM urls
		 WHERE code = $1`,
		code,
	).Scan(
		&analytics.Code,
		&analytics.URL,
		&analytics.ClickCount,
	)

	if err != nil {
		return service.URLAnalytics{}, err
	}

	rows, err := r.db.QueryContext(
		context.Background(),
		`SELECT clicked_at, user_agent, referer
		 FROM url_clicks
		 WHERE url_id = (
		     SELECT id FROM urls WHERE code = $1
		 )
		 ORDER BY clicked_at DESC`,
		code,
	)
	if err != nil {
		return service.URLAnalytics{}, err
	}
	defer rows.Close()

	analytics.Clicks = make([]service.Click, 0)

	for rows.Next() {
		var click service.Click

		if err := rows.Scan(
			&click.ClickedAt,
			&click.UserAgent,
			&click.Referer,
		); err != nil {
			return service.URLAnalytics{}, err
		}

		analytics.Clicks = append(analytics.Clicks, click)
	}

	if err := rows.Err(); err != nil {
		return service.URLAnalytics{}, err
	}

	return analytics, nil
}
