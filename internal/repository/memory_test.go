package repository

import "testing"

func TestMemoryURLRepository_RecordClick(t *testing.T) {
	repo := NewMemoryURLRepository()

	if err := repo.RecordClick(
		42,
		"TestBrowser/1.0",
		"https://google.com",
	); err != nil {
		t.Fatalf("failed to record click: %v", err)
	}

	repo.mu.RLock()
	defer repo.mu.RUnlock()

	if len(repo.clicks) != 1 {
		t.Fatalf("expected 1 click, got %d", len(repo.clicks))
	}

	click := repo.clicks[0]

	if click.urlID != 42 {
		t.Errorf("expected URL ID 42, got %d", click.urlID)
	}

	if click.userAgent != "TestBrowser/1.0" {
		t.Errorf(
			"expected user agent TestBrowser/1.0, got %s",
			click.userAgent,
		)
	}

	if click.referer != "https://google.com" {
		t.Errorf(
			"expected referer https://google.com, got %s",
			click.referer,
		)
	}

	if click.clickedAt.IsZero() {
		t.Error("expected click timestamp to be set")
	}
}
