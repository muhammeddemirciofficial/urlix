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
		strings.NewReader(`{"url":"not-a-url}`),
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
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"code"`) {
		t.Fatalf("expected response to contain code, got %s", rec.Body.String())
	}

}
