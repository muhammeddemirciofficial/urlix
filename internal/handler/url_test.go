package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/muhammeddemirciofficial/urlix/internal/repository"
	"github.com/muhammeddemirciofficial/urlix/internal/service"
)

func TestCreateURLInvalidURL(t *testing.T) {
	repo := repository.NewMemoryURLRepository()
	svc := service.NewURLService(repo)
	handler := NewURLHandler(svc)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/urls",
		strings.NewReader(`{"url":"not-a-url"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	handler.CreateURL(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCreateURLSuccess(t *testing.T) {
	repo := repository.NewMemoryURLRepository()
	svc := service.NewURLService(repo)
	handler := NewURLHandler(svc)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/urls",
		strings.NewReader(`{"url":"https://example.com"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	handler.CreateURL(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rec.Code)
	}

	if !strings.Contains(rec.Body.String(), `"code"`) {
		t.Fatalf("expected response to contain code, got %s", rec.Body.String())
	}
}

func TestCreateURLWithAlias(t *testing.T) {
	repo := repository.NewMemoryURLRepository()
	svc := service.NewURLService(repo)
	handler := NewURLHandler(svc)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/urls",
		strings.NewReader(`{"url":"https://example.com","alias":"google"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	handler.CreateURL(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rec.Code)
	}

	if !strings.Contains(rec.Body.String(), `"code":"google"`) {
		t.Fatalf("expected response to contain alias google, got %s", rec.Body.String())
	}
}

func TestCreateURLDuplicateAlias(t *testing.T) {
	repo := repository.NewMemoryURLRepository()
	svc := service.NewURLService(repo)
	handler := NewURLHandler(svc)

	firstReq := httptest.NewRequest(
		http.MethodPost,
		"/api/urls",
		strings.NewReader(`{"url":"https://example.com","alias":"google"}`),
	)
	firstReq.Header.Set("Content-Type", "application/json")

	firstRec := httptest.NewRecorder()
	handler.CreateURL(firstRec, firstReq)

	if firstRec.Code != http.StatusCreated {
		t.Fatalf("expected first request status 201, got %d", firstRec.Code)
	}

	secondReq := httptest.NewRequest(
		http.MethodPost,
		"/api/urls",
		strings.NewReader(`{"url":"https://example.org","alias":"google"}`),
	)
	secondReq.Header.Set("Content-Type", "application/json")

	secondRec := httptest.NewRecorder()
	handler.CreateURL(secondRec, secondReq)

	if secondRec.Code != http.StatusConflict {
		t.Fatalf("expected second request status 409, got %d", secondRec.Code)
	}
}
