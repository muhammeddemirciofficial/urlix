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

func TestRedirectRecordsClick(t *testing.T) {
	repo := repository.NewMemoryURLRepository()
	svc := service.NewURLService(repo)
	handler := NewURLHandler(svc)

	createReq := httptest.NewRequest(
		http.MethodPost,
		"/api/urls",
		strings.NewReader(`{"url":"https://example.com","alias":"trackme"}`),
	)
	createReq.Header.Set("Content-Type", "application/json")

	createRec := httptest.NewRecorder()
	handler.CreateURL(createRec, createReq)

	if createRec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", createRec.Code)
	}

	redirectReq := httptest.NewRequest(
		http.MethodGet,
		"/r/trackme",
		nil,
	)
	redirectReq.SetPathValue("code", "trackme")
	redirectReq.Header.Set("User-Agent", "TestBrowser/1.0")
	redirectReq.Header.Set("Referer", "https://google.com")

	redirectRec := httptest.NewRecorder()
	handler.Redirect(redirectRec, redirectReq)

	if redirectRec.Code != http.StatusFound {
		t.Fatalf("expected status 302, got %d", redirectRec.Code)
	}

	location := redirectRec.Header().Get("Location")
	if location != "https://example.com" {
		t.Fatalf("expected redirect location https://example.com, got %s", location)
	}
}

func TestGetAnalytics(t *testing.T) {
	repo := repository.NewMemoryURLRepository()
	svc := service.NewURLService(repo)
	handler := NewURLHandler(svc)

	createReq := httptest.NewRequest(
		http.MethodPost,
		"/api/urls",
		strings.NewReader(`{"url":"https://example.com","alias":"analytics"}`),
	)
	createReq.Header.Set("Content-Type", "application/json")

	createRec := httptest.NewRecorder()
	handler.CreateURL(createRec, createReq)

	if createRec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", createRec.Code)
	}

	result, err := repo.Get("analytics")
	if err != nil {
		t.Fatalf("failed to get created URL: %v", err)
	}

	if err := svc.RecordClick(
		result.ID,
		"TestBrowser/1.0",
		"https://google.com",
	); err != nil {
		t.Fatalf("failed to record click: %v", err)
	}

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/urls/analytics/analytics",
		nil,
	)
	req.SetPathValue("code", "analytics")

	rec := httptest.NewRecorder()
	handler.GetAnalytics(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	body := rec.Body.String()

	if !strings.Contains(body, `"code":"analytics"`) {
		t.Errorf("expected response to contain code, got %s", body)
	}

	if !strings.Contains(body, `"url":"https://example.com"`) {
		t.Errorf("expected response to contain URL, got %s", body)
	}

	if !strings.Contains(body, `"click_count":1`) {
		t.Errorf("expected click count 1, got %s", body)
	}

	if !strings.Contains(body, `"user_agent":"TestBrowser/1.0"`) {
		t.Errorf("expected user agent, got %s", body)
	}

	if !strings.Contains(body, `"referer":"https://google.com"`) {
		t.Errorf("expected referer, got %s", body)
	}
}

func TestGetAnalyticsNotFound(t *testing.T) {
	repo := repository.NewMemoryURLRepository()
	svc := service.NewURLService(repo)
	handler := NewURLHandler(svc)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/urls/notfound/analytics",
		nil,
	)
	req.SetPathValue("code", "notfound")

	rec := httptest.NewRecorder()
	handler.GetAnalytics(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}
}
